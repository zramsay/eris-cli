package data

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"regexp"

	"github.com/monax/monax/config"
	"github.com/monax/monax/definitions"
	"github.com/monax/monax/loaders"
	"github.com/monax/monax/log"
	"github.com/monax/monax/perform"
	"github.com/monax/monax/util"

	docker "github.com/fsouza/go-dockerclient"
)

// used by tests; could be refactored/deprecated
func RmData(do *definitions.Do) (err error) {
	if len(do.Operations.Args) == 0 {
		do.Operations.Args = []string{do.Name}
	}
	for _, name := range do.Operations.Args {
		do.Name = name
		if util.IsData(do.Name) {
			log.WithField("=>", do.Name).Info("Removing data container")

			srv := definitions.BlankServiceDefinition()
			srv.Operations.SrvContainerName = util.ContainerName("data", do.Name)

			if err = perform.DockerRemove(srv.Service, srv.Operations, false, do.Volumes, false); err != nil {
				log.Errorf("Error removing %s: %v", do.Name, err)
				return err
			}

		} else {
			err = fmt.Errorf("I cannot find that data container for %s. Please check the data container name you sent me.", do.Name)
			log.Error(err)
			return err
		}

		if do.RmHF {
			log.WithField("=>", do.Name).Warn("Removing host directory")
			if err = os.RemoveAll(filepath.Join(config.DataContainersPath, do.Name)); err != nil {
				return err
			}
		}
	}
	return err
}

// ImportData does what it says. It imports from a host's Source to a Dest
// in a data container. It returns an error.
//
//  do.Name                       - name of the data container to use (required)
//  do.Source                     - directory which should be imported (required)
//  do.Destination                - directory to _unload_ the payload into (required)
//
// If the named data container does not exist, it will be created
// If do.Destination does not exist, it will be created
func ImportData(do *definitions.Do) error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	do.Source = config.AbsolutePath(wd, do.Source)

	// Check the source path exists.
	if _, err := os.Stat(do.Source); err != nil {
		return err
	}

	log.WithFields(log.Fields{
		"from": do.Source,
		"to":   do.Destination,
	}).Debug("Importing")

	if util.IsData(do.Name) {
		srv := PretendToBeAService(do.Name)
		exists := perform.ContainerExists(srv.Operations.SrvContainerName)

		if !exists {
			return fmt.Errorf("There is no data container for service %q", do.Name)
		}
		if err := checkMonaxContainerRoot(do, "import"); err != nil {
			return err
		}

		containerName := util.DataContainerName(do.Name)

		doCheck := definitions.NowDo()
		doCheck.Name = do.Name
		doCheck.Operations.Args = []string{"test", "-d", do.Destination}
		_, err := ExecData(doCheck)
		if err != nil {
			log.WithError(err).WithFields(log.Fields{
				"destination": do.Destination,
			}).Info("Directory missing")
			if err := runData(containerName, []string{"/bin/mkdir", "-p", do.Destination}); err != nil {
				return err
			}
			return ImportData(do)
		}

		reader, err := util.TarForDocker(do.Source, 0)
		if err != nil {
			return err
		}
		defer reader.Close()

		opts := docker.UploadToContainerOptions{
			InputStream:          reader,
			Path:                 do.Destination,
			NoOverwriteDirNonDir: true,
		}

		log.WithField("=>", containerName).Info("Copying into container")
		log.WithField("path", do.Source).Debug()
		if err := util.DockerClient.UploadToContainer(srv.Operations.SrvContainerName, opts); err != nil {
			return util.DockerError(err)
		}

		//required b/c `docker cp` (UploadToContainer) goes in as root
		// and monax images have the `monax` user by default
		if err := runData(containerName, []string{"chown", "-R", "monax", do.Destination}); err != nil {
			return util.DockerError(err)
		}

	} else {
		log.WithField("name", do.Name).Info("Data container does not exist, creating it")
		ops := loaders.LoadDataDefinition(do.Name)
		if err := perform.DockerCreateData(ops); err != nil {
			return fmt.Errorf("Error creating data container %v.", err)
		}

		return ImportData(do)
	}
	return nil
}

func ExecData(do *definitions.Do) (buf *bytes.Buffer, err error) {
	if util.IsData(do.Name) {
		log.WithField("=>", do.Operations.DataContainerName).Info("Executing data container")

		ops := loaders.LoadDataDefinition(do.Name)
		util.Merge(ops, do.Operations)
		buf, err = perform.DockerExecData(ops, nil)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, fmt.Errorf("That data container does not exist")
	}
	return buf, nil
}

//export from: do.Source(in container), to: do.Destination(on host)
func ExportData(do *definitions.Do) error {
	if util.IsData(do.Name) {
		wd, err := os.Getwd()
		if err != nil {
			return err
		}
		do.Destination = config.AbsolutePath(wd, do.Destination)
		log.WithField("=>", do.Name).Info("Exporting data container")

		// we want to export to a temp directory.
		exportPath, err := ioutil.TempDir(os.TempDir(), do.Name)
		defer os.Remove(exportPath)
		if err != nil {
			return err
		}

		containerName := util.DataContainerName(do.Name)
		srv := PretendToBeAService(do.Name)
		exists := perform.ContainerExists(srv.Operations.SrvContainerName)

		if !exists {
			return fmt.Errorf("There is no data container for that service")
		}

		reader, writer := io.Pipe()
		defer reader.Close()

		if !do.Operations.SkipCheck { // sometimes you want greater flexibility
			if err := checkMonaxContainerRoot(do, "export"); err != nil {
				return err
			}
		}

		opts := docker.DownloadFromContainerOptions{
			OutputStream: writer,
			Path:         do.Source,
		}

		go func() {
			log.WithField("=>", containerName).Info("Copying out of container")
			log.WithField("path", do.Source).Debug()
			util.IfExit(util.DockerClient.DownloadFromContainer(srv.Operations.SrvContainerName, opts))
			writer.Close()
		}()

		log.WithField("=>", exportPath).Debug("Untarring package from container")
		if err = util.UntarForDocker(reader, do.Name, exportPath); err != nil {
			return err
		}

		// now if docker dumps to exportPath/.monax we should remove
		// move everything from .monax to exportPath
		if err := MoveOutOfDirAndRmDir(filepath.Join(exportPath, ".monax"), exportPath); err != nil {
			return err
		}

		// finally remove everything in the data directory and move
		// the temp contents there
		if _, err := os.Stat(do.Destination); os.IsNotExist(err) {
			if e2 := os.MkdirAll(do.Destination, 0755); e2 != nil {
				return fmt.Errorf("The marmots could neither find, nor had access to make the directory %s", do.Destination)
			}
		}
		if err := MoveOutOfDirAndRmDir(exportPath, do.Destination); err != nil {
			return err
		}
	} else {
		return fmt.Errorf("That data container does not exist")
	}
	return nil
}

func PretendToBeAService(serviceYourPretendingToBe string) *definitions.ServiceDefinition {
	srv := definitions.BlankServiceDefinition()
	srv.Name = serviceYourPretendingToBe

	giveMeAllTheNames(serviceYourPretendingToBe, srv)
	return srv
}

func MoveOutOfDirAndRmDir(src, dest string) error {
	log.WithFields(log.Fields{
		"from": src,
		"to":   dest,
	}).Info("Move all files/dirs out of a dir and `rm -rf` that dir")
	toMove, err := filepath.Glob(filepath.Join(src, "*"))
	if err != nil {
		return err
	}

	if len(toMove) == 0 {
		log.Debug("No files to move")
	}

	for _, f := range toMove {
		// using a copy (read+write) strategy to get around swap partitions and other
		//   problems that cause a simple rename strategy to fail. it is more io overhead
		//   to do this, but for now that is preferable to alternative solutions.
		config.Copy(f, filepath.Join(dest, filepath.Base(f)))
	}

	log.WithField("=>", src).Info("Removing directory")
	return os.RemoveAll(src)
}

func runData(name string, args []string) error {
	doRun := definitions.NowDo()
	doRun.Operations.DataContainerName = name
	doRun.Operations.ContainerType = "data"
	doRun.Operations.Args = args
	out, err := perform.DockerRunData(doRun.Operations, nil)
	if err != nil {
		log.Debug("Dumping failed data container logs")
		log.Debug(out)

		return fmt.Errorf("Error running args %q: %v", args, err)
	}
	return nil
}

// check path for config.MonaxContainerRoot
func checkMonaxContainerRoot(do *definitions.Do, typ string) error {
	r, err := regexp.Compile(config.MonaxContainerRoot)
	if err != nil {
		return err
	}

	switch typ {
	case "import":
		if !r.MatchString(do.Destination) { //if not there join it
			do.Destination = path.Join(config.MonaxContainerRoot, do.Destination)
			return nil
		} else { // matches: do nothing
			return nil
		}
	case "export":
		if !r.MatchString(do.Source) {
			do.Source = path.Join(config.MonaxContainerRoot, do.Source)
			return nil
		} else {
			return nil
		}
	}
	return nil
}

func giveMeAllTheNames(name string, srv *definitions.ServiceDefinition) {
	log.WithField("=>", name).Debug("Giving myself all the names")
	srv.Name = name
	srv.Service.Name = name
	srv.Operations.SrvContainerName = util.DataContainerName(srv.Name)
	srv.Operations.DataContainerName = util.DataContainerName(srv.Name)
	log.WithFields(log.Fields{
		"data container":    srv.Operations.DataContainerName,
		"service container": srv.Operations.SrvContainerName,
	}).Debug("Using names")
}
