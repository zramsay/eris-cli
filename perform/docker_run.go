package perform

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/eris-ltd/eris-cli/config"
	def "github.com/eris-ltd/eris-cli/definitions"
	"github.com/eris-ltd/eris-cli/util"

	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/docker/docker/pkg/term"
	dirs "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common/go/common"
	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/fsouza/go-dockerclient"
)

// Tested against Docker API versions: 1.18, 1.19, 1.20
func DockerCreateDataContainer(srvName string, containerNumber int) error {
	logger.Infof("Creating Data Container for =>\t%s\n", srvName)

	srv := def.BlankServiceDefinition()
	srv.Operations.DataContainerName = util.DataContainersName(srvName, containerNumber)
	optsData, err := configureDataContainer(srv.Service, srv.Operations, nil)
	if err != nil {
		return err
	}

	srv.Operations.SrvContainerName = srv.Operations.DataContainerName // mock for the query function
	if _, exists := ContainerExists(srv.Operations); exists {
		logger.Infoln("Data container exists. Not creating.")
		return nil
	}

	cont, err := createContainer(optsData)
	if err != nil {
		return err
	}

	logger.Infof("Data Container ID =>\t\t%s\n", cont.ID)
	return nil
}

// create a container with volumes-from the srvName data container
// and either attach interactively or execute a command
// container should be destroyed on exit
func DockerRunVolumesFromContainer(volumesFrom string, interactive bool, args []string, service *def.Service) (result []byte, err error) {
	logger.Infof("DockerRunVolumesFromContnr =>\t%s:%v\n", volumesFrom, args)
	opts := configureVolumesFromContainer(volumesFrom, interactive, args, service)
	cont, err := createContainer(opts)
	if err != nil {
		return nil, err
	}
	id_main := cont.ID

	// trap signals so we can drop out of the container
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)
	go func() {
		<-c
		logger.Infof("\nCaught signal. Stopping container %s\n", id_main)
		if err = stopContainer(id_main, 5); err != nil {
			logger.Errorf("Error stopping container: %v\n", err)
		}
	}()

	defer func() {
		logger.Infof("Removing container =>\t\t%s\n", id_main)
		if err2 := removeContainer(id_main, true); err2 != nil {
			err = fmt.Errorf("Tragic! Error removing data container after executing (%v): %v", err, err2)
		}
		logger.Infof("Container removed =>\t\t%s\n", id_main)
	}()

	attached := make(chan struct{})
	if interactive {
		logger.Debugf("Attaching to container =>\t%s\n", id_main)

		savedState, err := term.SetRawTerminal(os.Stdin.Fd())
		if err != nil {
			logger.Errorln("Cannot set the terminal into raw mode")
		} else {
			defer term.RestoreTerminal(os.Stdin.Fd(), savedState)
		}

		// attachContainer uses hijack so we need to run this in a goroutine.
		go func(chan struct{}) {
			attachContainer(id_main, attached)
		}(attached)
	}

	// start the container (either interactive or one off command)
	logger.Infof("Exec Container ID =>\t\t%s\n", id_main)
	if err = startContainer(id_main, &opts); err != nil {
		return nil, err
	}

	if interactive {
		// Wait for the console prompt to appear.
		_, ok := <-attached
		if ok {
			attached <- struct{}{}
		}
	}

	logger.Infof("Waiting to exit for removal =>\t%s\n", id_main)
	if err = waitContainer(id_main); err != nil {
		return nil, err
	}

	if !interactive {
		logger.Debugf("Getting logs for container =>\t%s\n", id_main)
		if err = logsContainer(id_main, true, "all"); err != nil {
			return nil, err
		}
		// now lets get the logs out
		// XXX: we only do this if the global config writer is a bytes.Buffer
		if config.GlobalConfig.Writer != nil {
			writer := config.GlobalConfig.Writer
			reader, ok := writer.(*bytes.Buffer)
			if !ok {
				return nil, nil
			}

			done := make(chan struct{}, 1)
			var b []byte
			go func() {
				// TODO: this routine will hang forever if  ReadAll doesn't complete
				// need to be smarter
				logger.Debugln("Attempting to read log reader.")
				b, err = ioutil.ReadAll(reader)
				done <- struct{}{}
			}()
			ticker := time.NewTicker(time.Second * 2)
		LOOP:
			for {
				select {
				case <-ticker.C:
					logger.Debugln("tick!")
					if reader.Len() == 0 {
						// nothing to read means dont bother waiting
						break LOOP
					} else {
						logger.Debugln("Read something", reader.Len())
					}
				case <-done:
					return b, err
				}
			}
		}
	}

	return nil, nil
}

func DockerRun(srv *def.Service, ops *def.Operation) error {
	var id_main, id_data string
	var optsData docker.CreateContainerOptions
	var dataCont docker.APIContainers
	var dataContCreated *docker.Container

	_, running := ContainerRunning(ops)
	if running {
		logger.Infof("Service already Started. Skipping.\n\tService Name=>\t\t%s\n", srv.Name)
		return nil
	}

	logger.Infof("Starting Service =>\t\t%s\n", srv.Name)

	// copy service config into docker client config
	optsServ, err := configureServiceContainer(srv, ops)
	if err != nil {
		return err
	}

	// fix volume paths
	srv.Volumes, err = fixDirs(srv.Volumes)
	if err != nil {
		return err
	}

	// setup data container
	logger.Infof("Manage data containers? =>\t%t\n", srv.AutoData)
	if srv.AutoData {
		optsData, err = configureDataContainer(srv, ops, &optsServ)
		if err != nil {
			return err
		}
	}

	// check existence || create the container
	if servCont, exists := ContainerExists(ops); exists {
		logger.Infoln("Service Container already exists, am not creating.")

		if srv.AutoData {
			if dataCont, exists = util.ParseContainers(ops.DataContainerName, true); exists {
				logger.Infoln("Data Container already exists, am not creating.")
				id_data = dataCont.ID
			} else {
				logger.Infoln("Data Container does not exist, creating.")
				dataContCreated, err := createContainer(optsData)
				if err != nil {
					return err
				}
				id_data = dataContCreated.ID
			}
		}

		id_main = servCont.ID

	} else {
		logger.Infof("Service Container does not exist, creating from image (%s).\n", srv.Image)

		if srv.AutoData {
			if dataCont, exists = util.ParseContainers(ops.DataContainerName, true); exists {
				logger.Infoln("Data Container already exists, am not creating.")
				id_data = dataCont.ID
			} else {
				logger.Infoln("Data Container does not exist, creating.")
				dataContCreated, err = createContainer(optsData)
				if err != nil {
					return err
				}
				id_data = dataContCreated.ID
			}
		}

		servContCreated, err := createContainer(optsServ)
		if err != nil {
			return err
		}
		id_main = servContCreated.ID
	}

	// start the container
	logger.Infof("Starting Service Contanr ID =>\t%s:%s\n", optsServ.Name, id_main)
	if srv.AutoData {
		logger.Infof("\twith DataContanr ID =>\t%s\n", id_data)
	}
	logger.Debugf("\twith EntryPoint =>\t%v\n", optsServ.Config.Entrypoint)
	logger.Debugf("\twith CMD =>\t\t%v\n", optsServ.Config.Cmd)
	logger.Debugf("\twith Image =>\t\t%v\n", optsServ.Config.Image)
	// logger.Debugf("\twith Environment =>\t%s\n", optsServ.Config.Env)
	logger.Debugf("\twith AllPortsPubl'd =>\t%v\n", optsServ.HostConfig.PublishAllPorts)
	if err := startContainer(id_main, &optsServ); err != nil {
		return err
	}

	// XXX: setting Remove causes us to block here!
	if ops.Remove {

		// dump the logs (TODO: options about this)
		doneLogs := make(chan struct{}, 1)
		go func() {
			logger.Debugln("DockerRun. Following logs.")
			if err := logsContainer(id_main, true, "all"); err != nil {
				logger.Errorf("Unable to follow logs for %s\n", id_main)
			}
			logger.Debugln("DockerRun. Finished following logs.")
			doneLogs <- struct{}{}
		}()

		logger.Infof("Waiting to exit for removal =>\t%s\n", id_main)
		if err := waitContainer(id_main); err != nil {
			return err
		}

		logger.Debugln("DockerRun. Waiting for logs to finish.")
		// let the logs finish
		<-doneLogs

		logger.Infof("DockerRun. Removing cont =>\t%s\n", id_main)
		if err := removeContainer(id_main, false); err != nil {
			return err
		}

	} else {
		logger.Infof("Successfully started service =>\t%s\n", srv.Name)
	}

	return nil
}

func DockerRunInteractive(srv *def.Service, ops *def.Operation, args []string, interactive bool, volume string) error {
	var id_main, id_data string
	var optsData docker.CreateContainerOptions
	var dataContCreated *docker.Container

	defer func() {
		logger.Infof("Removing container =>\t\t%s\n", id_main)
		if err := removeContainer(id_main, false); err != nil {
			fmt.Errorf("Tragic! Error removing data container after executing (%v): %v", id_main, err)
		}
		logger.Infof("Container removed =>\t\t%s\n", id_main)
	}()

	logger.Infof("Starting Service =>\t\t%s\n", srv.Name)
	optsServ, err := configureInteractiveContainer(srv, ops, args, interactive, volume)

	// Fix volume paths.
	srv.Volumes, err = fixDirs(srv.Volumes)
	if err != nil {
		return err
	}

	logger.Infof("Creating from image (%s)\n", srv.Image)
	if srv.AutoData {
		optsData, err = configureDataContainer(srv, ops, &optsServ)
		if err != nil {
			return err
		}
		if dataCont, exists := util.ParseContainers(ops.DataContainerName, true); exists {
			logger.Infoln("Data Container already exists, am not creating.")
			id_data = dataCont.ID
		} else {
			logger.Infoln("Data Container does not exist, creating.")
			dataContCreated, err = createContainer(optsData)
			if err != nil {
				return err
			}
			id_data = dataContCreated.ID
		}
	}

	servContCreated, err := createContainer(optsServ)
	if err != nil {
		return err
	}
	id_main = servContCreated.ID

	// Set terminal into row mode, and restore upon container exit.
	savedState, err := term.SetRawTerminal(os.Stdin.Fd())
	if err != nil {
		logger.Errorln("Cannot set the terminal into raw mode")
	} else {
		defer term.RestoreTerminal(os.Stdin.Fd(), savedState)
	}

	// Trap signals so we can drop out of the container.
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)
	go func() {
		<-c
		logger.Infof("\nCaught signal. Stopping container %s\n", id_main)
		if err = stopContainer(id_main, 5); err != nil {
			logger.Errorf("Error stopping container: %v\n", err)
		}
	}()

	// This channel receives an event that the container is attached.
	attached := make(chan struct{})
	go func(chan struct{}) {
		attachContainer(id_main, attached)
	}(attached)

	// Start the container.
	logger.Infof("Starting Service Contanr ID =>\t%s:%s\n", optsServ.Name, id_main)
	if srv.AutoData {
		logger.Infof("\twith DataContanr ID =>\t%s\n", id_data)
	}
	logger.Debugf("\twith EntryPoint =>\t%v\n", optsServ.Config.Entrypoint)
	logger.Debugf("\twith CMD =>\t\t%v\n", optsServ.Config.Cmd)
	logger.Debugf("\twith Image =>\t\t%v\n", optsServ.Config.Image)
	logger.Debugf("\twith AllPortsPubl'd =>\t%v\n", optsServ.HostConfig.PublishAllPorts)

	if err := startContainer(id_main, &optsServ); err != nil {
		return err
	}

	// Wait for a console prompt to appear.
	_, ok := <-attached
	if ok {
		attached <- struct{}{}
	}

	logger.Infof("Waiting to exit for removal =>\t%s\n", id_main)
	if err := waitContainer(id_main); err != nil {
		return err
	}

	return nil
}

func DockerRebuild(srv *def.Service, ops *def.Operation, skipPull bool, timeout uint) error {
	var id string
	var wasRunning bool = false

	logger.Infof("Starting Docker Rebuild =>\t%s\n", srv.Name)

	if service, exists := ContainerExists(ops); exists {
		if _, running := ContainerRunning(ops); running {
			wasRunning = true
			err := DockerStop(srv, ops, timeout)
			if err != nil {
				return err
			}
		}

		logger.Infof("Removing old container =>\t%s\n", service.ID)
		err := removeContainer(service.ID, true)
		if err != nil {
			return err
		}

	} else {
		logger.Infoln("Service did not previously exist. Nothing to rebuild.")
		return nil
	}

	if !skipPull {
		logger.Infof("Pulling new image =>\t\t%s\n", srv.Image)
		err := DockerPull(srv, ops)
		if err != nil {
			return err
		}
	}

	opts, err := configureServiceContainer(srv, ops)
	if err != nil {
		return err
	}
	srv.Volumes, err = fixDirs(srv.Volumes)
	if err != nil {
		return err
	}

	logger.Infof("Creating new cont for srv =>\t%s\n", srv.Name)
	cont, err := createContainer(opts)
	if err != nil {
		return err
	}
	id = cont.ID

	if wasRunning {
		logger.Infof("Restarting srv with new ID =>\t%s\n", id)
		err := startContainer(id, &opts)
		if err != nil {
			return err
		}
	}

	logger.Infof("Finished rebuilding service =>\t%s\n", srv.Name)

	return nil
}

func DockerPull(srv *def.Service, ops *def.Operation) error {
	logger.Infof("Pulling an image (%s) for the service (%s)\n", srv.Image, srv.Name)

	var wasRunning bool = false

	if service, exists := ContainerExists(ops); exists {
		logger.Infoln("Found Service ID: " + service.ID)
		if _, running := ContainerRunning(ops); running {
			wasRunning = true
			err := DockerStop(srv, ops, 10)
			if err != nil {
				return err
			}
		}
		err := removeContainer(service.ID, false)
		if err != nil {
			return err
		}
	}

	if logger.Level > 0 {
		err := pullImage(srv.Image, logger.Writer)
		if err != nil {
			return err
		}
	} else {
		err := pullImage(srv.Image, bytes.NewBuffer([]byte{}))
		if err != nil {
			return err
		}
	}

	if wasRunning {
		err := DockerRun(srv, ops)
		if err != nil {
			return err
		}
	}

	return nil
}

func DockerLogs(srv *def.Service, ops *def.Operation, follow bool, tail string) error {
	if service, exists := ContainerExists(ops); exists {
		logger.Infof("Getting Logs for Service ID =>\t%s:%v:%v\n", service.ID, follow, tail)
		if err := logsContainer(service.ID, follow, tail); err != nil {
			return err
		}
	} else {
		logger.Infoln("Service does not exist. Cannot display logs.")
	}

	return nil
}

func DockerInspect(srv *def.Service, ops *def.Operation, field string) error {
	if service, exists := ContainerExists(ops); exists {
		logger.Infof("Inspecting Service ID =>\t%s\n", service.ID)
		err := inspectContainer(service.ID, field)
		if err != nil {
			return err
		}
	} else {
		logger.Infoln("Service container does not exist. Cannot inspect.")
	}
	return nil
}

func DockerStop(srv *def.Service, ops *def.Operation, timeout uint) error {
	// don't limit this to verbose because it takes a few seconds
	logger.Printf("Docker is Stopping =>\t\t%s\tThis may take a few seconds.\n", srv.Name)
	logger.Debugf("\twith ContainerNumber =>\t%d\n", ops.ContainerNumber)
	logger.Debugf("\twith SrvContnerName =>\t%s\n", ops.SrvContainerName)

	dockerAPIContainer, running := ContainerExists(ops)

	if running {
		logger.Infof("Service is running =>\t\t%s:%d\n", srv.Name, ops.ContainerNumber)
		err := stopContainer(dockerAPIContainer.ID, timeout)
		if err != nil {
			return err
		}
	} else {
		logger.Infof("Service is not running =>\t%s:%d\n", srv.Name, ops.ContainerNumber)
	}

	logger.Infof("Finished stopping service =>\t%s\n", srv.Name)
	return nil
}

func DockerRename(srv *def.Service, ops *def.Operation, oldName, newName string) error {
	// don't limit this to verbose because it takes a few seconds
	logger.Debugf("Docker is Renaming =>\t\t%s:%s:%d\n", srv.Name, newName, ops.ContainerNumber)
	logger.Debugf("\twith ContainerNumber =>\t%d\n", ops.ContainerNumber)
	logger.Debugf("\twith SrvContnerName =>\t%s\n", ops.SrvContainerName)

	if service, exists := ContainerExists(ops); exists {
		logger.Infof("Renaming Service ID =>\t\t%s\n", service.ID)
		newName = strings.Replace(service.Names[0], oldName, newName, 1)
		err := renameContainer(service.ID, newName)
		if err != nil {
			return err
		}
	} else {
		logger.Infoln("Service container does not exist. Cannot rename.")
	}

	return nil
}

func DockerRemove(srv *def.Service, ops *def.Operation, withData, volumes bool) error {
	if service, exists := ContainerExists(ops); exists {
		logger.Infof("Removing Service ID =>\t\t%s\n", service.ID)
		if err := removeContainer(service.ID, volumes); err != nil {
			return err
		}
		if withData {
			if srv, ext := ContainerDataContainerExists(ops); ext {
				logger.Infof("\t with DataContanr ID =>\t%s\n", srv.ID)
				if err := removeContainer(srv.ID, volumes); err != nil {
					return err
				}
			}
		}
	} else {
		logger.Infoln("Service container does not exist. Cannot remove.")
	}

	return nil
}

func ContainerExists(ops *def.Operation) (docker.APIContainers, bool) {
	return util.ParseContainers(ops.SrvContainerName, true)
}

func ContainerRunning(ops *def.Operation) (docker.APIContainers, bool) {
	return util.ParseContainers(ops.SrvContainerName, false)
}

func ContainerDataContainerExists(ops *def.Operation) (docker.APIContainers, bool) {
	return util.ParseContainers(ops.DataContainerName, true)
}

// ----------------------------------------------------------------------------
// ---------------------    Images Core    ------------------------------------
// ----------------------------------------------------------------------------
func pullImage(name string, writer io.Writer) error {
	var tag string = "latest"
	var reg string = ""

	nameSplit := strings.Split(name, ":")
	if len(nameSplit) == 2 {
		tag = nameSplit[1]
	}
	if len(nameSplit) == 3 {
		tag = nameSplit[2]
	}
	name = nameSplit[0]

	repoSplit := strings.Split(nameSplit[0], "/")
	if len(repoSplit) > 2 {
		reg = repoSplit[0]
	}

	opts := docker.PullImageOptions{
		Repository:   name,
		Registry:     reg,
		Tag:          tag,
		OutputStream: os.Stdout,
	}

	if os.Getenv("ERIS_PULL_APPROVE") == "true" {
		opts.OutputStream = nil
	}

	auth := docker.AuthConfiguration{}

	err := util.DockerClient.PullImage(opts, auth)
	if err != nil {
		return err
	}

	return nil
}

// ----------------------------------------------------------------------------
// ---------------------    Container Core ------------------------------------
// ----------------------------------------------------------------------------
func createContainer(opts docker.CreateContainerOptions) (*docker.Container, error) {
	dockerContainer, err := util.DockerClient.CreateContainer(opts)
	if err != nil {
		if err == docker.ErrNoSuchImage {
			if os.Getenv("ERIS_PULL_APPROVE") != "true" {
				var input string
				logger.Printf("The docker image (%s) is not found locally.\nWould you like the marmots to pull it from the repository? (y/n) ", opts.Config.Image)
				fmt.Scanln(&input)

				if input == "Y" || input == "y" || input == "YES" || input == "Yes" || input == "yes" {
					logger.Debugf("\nUser assented to pull.\n")
				} else {
					logger.Debugf("\nUser refused to pull.\n")
					return nil, fmt.Errorf("Cannot start a container based on an image you will not let me pull.\n")
				}
			} else {
				logger.Printf("The docker image (%s) is not found locally.\nThe marmots are approved to pull from the repository on your behalf.\nThis could take a second.\n", opts.Config.Image)
			}
			if err := pullImage(opts.Config.Image, nil); err != nil {
				return nil, err
			}
			dockerContainer, err = util.DockerClient.CreateContainer(opts)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}
	return dockerContainer, nil
}

func startContainer(id string, opts *docker.CreateContainerOptions) error {
	return util.DockerClient.StartContainer(id, opts.HostConfig)
}

func attachContainer(id string, attached chan struct{}) error {
	opts := docker.AttachToContainerOptions{
		Container:    id,
		InputStream:  os.Stdin,
		OutputStream: config.GlobalConfig.Writer,
		ErrorStream:  config.GlobalConfig.ErrorWriter,
		Logs:         true,
		Stream:       true,
		Stdin:        true,
		Stdout:       true,
		Stderr:       true,
		RawTerminal:  true,
		Success:      attached,
	}

	return util.DockerClient.AttachToContainer(opts)
}

func waitContainer(id string) error {
	exitCode, err := util.DockerClient.WaitContainer(id)
	if exitCode != 0 {
		err1 := fmt.Errorf("Container %s exited with status %d", id, exitCode)
		if err != nil {
			err = fmt.Errorf("%s. Error: %v", err1.Error(), err)
		} else {
			err = err1
		}
	}
	return err
}

func logsContainer(id string, follow bool, tail string) error {
	var writer io.Writer
	var eWriter io.Writer

	if config.GlobalConfig != nil {
		writer = config.GlobalConfig.Writer
		eWriter = config.GlobalConfig.ErrorWriter
	} else {
		writer = os.Stdout
		eWriter = os.Stderr
	}

	opts := docker.LogsOptions{
		Container:    id,
		OutputStream: writer,
		ErrorStream:  eWriter,
		Follow:       follow,
		Stdout:       true,
		Stderr:       true,
		Since:        0,
		Timestamps:   false,
		Tail:         tail,

		RawTerminal: true, // Usually true when the container contains a TTY.
	}

	if err := util.DockerClient.Logs(opts); err != nil {
		return err
	}
	return nil
}

func inspectContainer(id, field string) error {
	cont, err := util.DockerClient.InspectContainer(id)
	if err != nil {
		return err
	}
	util.PrintInspectionReport(cont, field)

	return nil
}

func stopContainer(id string, timeout uint) error {
	logger.Debugf("\twith ContainerID =>\t%s\n", id)
	logger.Debugf("\twith Timeout =>\t\t%d\n", timeout)
	err := util.DockerClient.StopContainer(id, timeout)
	if err != nil {
		return err
	}
	return nil
}

func renameContainer(id, newName string) error {
	opts := docker.RenameContainerOptions{
		ID:   id,
		Name: newName,
	}

	err := util.DockerClient.RenameContainer(opts)
	if err != nil {
		return err
	}

	return nil
}

func removeContainer(id string, volumes bool) error {
	opts := docker.RemoveContainerOptions{
		ID:            id,
		RemoveVolumes: volumes,
		Force:         false,
	}

	err := util.DockerClient.RemoveContainer(opts)
	if err != nil {
		return err
	}

	return nil
}

func configureInteractiveContainer(srv *def.Service, ops *def.Operation, args []string, interactive bool, volume string) (docker.CreateContainerOptions, error) {
	opts, err := configureServiceContainer(srv, ops)
	if err != nil {
		return docker.CreateContainerOptions{}, err
	}

	opts.Name = "eris_interactive_" + opts.Name
	opts.Config.User = "root"
	opts.Config.OpenStdin = true
	opts.Config.Tty = true
	if interactive {
		// if there are args, we overwrite the entrypoint
		// else we just start an interactive shell
		if len(args) > 0 {
			opts.Config.Entrypoint = args
		} else {
			opts.Config.Entrypoint = []string{"/bin/bash"}
		}
	} else {
		// use the image's own entrypoint
		opts.Config.Cmd = args
	}

	// Mount a volume.
	if volume != "" {
		bind := filepath.Join(dirs.ErisRoot, volume) + ":" +
			filepath.Join(dirs.ErisContainerRoot, volume)

		if opts.HostConfig.Binds == nil {
			opts.HostConfig.Binds = []string{bind}
		} else {
			opts.HostConfig.Binds = append(opts.HostConfig.Binds, bind)
		}
	}

	// we expect to link to the main service container
	opts.HostConfig.Links = srv.Links

	// temporary hack.
	opts.HostConfig.PortBindings = make(map[docker.Port][]docker.PortBinding)

	return opts, nil
}

func configureServiceContainer(srv *def.Service, ops *def.Operation) (docker.CreateContainerOptions, error) {
	if ops.ContainerNumber == 0 {
		ops.ContainerNumber = 1
	}

	opts := docker.CreateContainerOptions{
		Name: ops.SrvContainerName,
		Config: &docker.Config{
			Hostname:        srv.HostName,
			Domainname:      srv.DomainName,
			User:            srv.User,
			Memory:          srv.MemLimit,
			CPUShares:       srv.CPUShares,
			AttachStdin:     false,
			AttachStdout:    false,
			AttachStderr:    false,
			Tty:             true,
			OpenStdin:       false,
			Env:             srv.Environment,
			Labels:          ops.Labels,
			Image:           srv.Image,
			NetworkDisabled: false,
		},
		HostConfig: &docker.HostConfig{
			Binds:           srv.Volumes,
			Links:           srv.Links,
			PublishAllPorts: ops.PublishAllPorts,
			Privileged:      ops.Privileged,
			ReadonlyRootfs:  false,
			DNS:             srv.DNS,
			DNSSearch:       srv.DNSSearch,
			VolumesFrom:     srv.VolumesFrom,
			CapAdd:          ops.CapAdd,
			CapDrop:         ops.CapDrop,
			RestartPolicy:   docker.NeverRestart(),
			NetworkMode:     "bridge",
		},
	}

	// some fields may be set in the dockerfile and we only want to overwrite if they are present in the service def
	if srv.EntryPoint != "" {
		opts.Config.Entrypoint = strings.Fields(srv.EntryPoint)
	}
	if srv.Command != "" {
		opts.Config.Cmd = strings.Fields(srv.Command)
	}
	if srv.WorkDir != "" {
		opts.Config.WorkingDir = srv.WorkDir
	}

	if ops.Attach {
		opts.Config.AttachStdin = true
		opts.Config.AttachStdout = true
		opts.Config.AttachStderr = true
		opts.Config.OpenStdin = true
	}

	if ops.Restart == "always" {
		opts.HostConfig.RestartPolicy = docker.AlwaysRestart()
	} else if strings.Contains(ops.Restart, "max") {
		times, err := strconv.Atoi(strings.Split(ops.Restart, ":")[1])
		if err != nil {
			return docker.CreateContainerOptions{}, err
		}
		opts.HostConfig.RestartPolicy = docker.RestartOnFailure(times)
	}

	opts.Config.ExposedPorts = make(map[docker.Port]struct{})
	opts.HostConfig.PortBindings = make(map[docker.Port][]docker.PortBinding)
	opts.Config.Volumes = make(map[string]struct{})

	// Don't fill in port bindings if randomizing the ports.
	if !ops.PublishAllPorts {
		for _, port := range srv.Ports {
			pS := strings.Split(port, ":")
			pC := docker.Port(util.PortAndProtocol(pS[len(pS)-1]))

			opts.Config.ExposedPorts[pC] = struct{}{}
			if len(pS) > 1 {
				pH := docker.PortBinding{
					HostPort: pS[len(pS)-2],
				}

				if len(pS) == 3 {
					// ipv4
					pH.HostIP = pS[0]
				} else if len(pS) > 3 {
					// ipv6
					pH.HostIP = strings.Join(pS[:len(pS)-2], ":")
				}

				opts.HostConfig.PortBindings[pC] = []docker.PortBinding{pH}
			} else {
				pH := docker.PortBinding{
					HostPort: pS[0],
				}
				opts.HostConfig.PortBindings[pC] = []docker.PortBinding{pH}
			}
		}
	}

	for _, vol := range srv.Volumes {
		opts.Config.Volumes[strings.Split(vol, ":")[1]] = struct{}{}
	}

	return opts, nil
}

func configureVolumesFromContainer(volumesFrom string, interactive bool, args []string, service *def.Service) docker.CreateContainerOptions {
	// set the defaults
	opts := docker.CreateContainerOptions{
		Name: "eris_exec_" + volumesFrom,
		Config: &docker.Config{
			Image:           "quay.io/eris/base",
			User:            "root",
			WorkingDir:      dirs.ErisContainerRoot,
			AttachStdout:    true,
			AttachStderr:    true,
			AttachStdin:     true,
			Tty:             true,
			NetworkDisabled: false,
		},
		HostConfig: &docker.HostConfig{
			VolumesFrom: []string{volumesFrom},
		},
	}
	if interactive {
		opts.Config.OpenStdin = true
		opts.Config.Cmd = []string{"/bin/bash"}
	} else {
		opts.Config.Cmd = args
	}

	// overwrite some things
	if service != nil {
		opts.Config.NetworkDisabled = false
		opts.Config.Image = service.Image
		opts.Config.User = service.User
		opts.Config.Env = service.Environment
		opts.HostConfig.Links = service.Links
		opts.Config.Entrypoint = strings.Fields(service.EntryPoint)
	}

	return opts
}

func configureDataContainer(srv *def.Service, ops *def.Operation, mainContOpts *docker.CreateContainerOptions) (docker.CreateContainerOptions, error) {
	// by default data containers will rely on the image used by
	//   the base service. sometimes, tho, especially for testing
	//   that base image will not be present. in such cases use
	//   the base eris data container.
	if srv.Image == "" {
		srv.Image = "quay.io/eris/data"
	}

	opts := docker.CreateContainerOptions{
		Name: ops.DataContainerName,
		Config: &docker.Config{
			Image:           srv.Image,
			User:            srv.User,
			AttachStdin:     false,
			AttachStdout:    false,
			AttachStderr:    false,
			Tty:             false,
			OpenStdin:       false,
			NetworkDisabled: true, // data containers do not need to talk to the outside world.
			Entrypoint:      []string{},
			Cmd:             []string{"false"}, // just gracefully exit. data containers just need to "exist" not run.
		},
		HostConfig: &docker.HostConfig{},
	}

	if mainContOpts != nil {
		mainContOpts.HostConfig.VolumesFrom = append(mainContOpts.HostConfig.VolumesFrom, ops.DataContainerName)
	}

	return opts, nil
}

// ----------------------------------------------------------------------------
// ---------------------    Exec Core -----------------------------------------
// ----------------------------------------------------------------------------
func createExec(container string, cmd []string, srv *def.Service) (*docker.Exec, error) {
	opts := docker.CreateExecOptions{
		AttachStdin:  false,
		AttachStdout: true,
		AttachStderr: true,
		Tty:          true,
		Cmd:          cmd,
		Container:    container,
		User:         srv.User,
	}

	if srv.User == "" {
		opts.User = "eris"
	}

	return util.DockerClient.CreateExec(opts)
}

func startExec(id string) error {
	opts := docker.StartExecOptions{
		Detach:       false,
		Tty:          true,
		InputStream:  nil,
		OutputStream: os.Stdout,
		ErrorStream:  os.Stderr,
		RawTerminal:  true,
	}

	return util.DockerClient.StartExec(id, opts)
}

// ----------------------------------------------------------------------------
// ---------------------    Util Funcs ----------------------------------------
// ----------------------------------------------------------------------------

// $(pwd) doesn't execute properly in golangs subshells; replace it
// use $eris as a shortcut
func fixDirs(arg []string) ([]string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return []string{}, err
	}

	for n, a := range arg {
		if strings.Contains(a, "$eris") {
			tmp := strings.Split(a, ":")[0]
			keep := strings.Replace(a, tmp+":", "", 1)
			if runtime.GOOS == "windows" {
				winTmp := strings.Split(tmp, "/")
				tmp = path.Join(winTmp...)
			}
			tmp = strings.Replace(tmp, "$eris", dirs.ErisRoot, 1)
			arg[n] = strings.Join([]string{tmp, keep}, ":")
			continue
		}

		if strings.Contains(a, "$pwd") {
			arg[n] = strings.Replace(a, "$pwd", dir, 1)
		}
	}

	return arg, nil
}
