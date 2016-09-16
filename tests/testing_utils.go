package tests

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/eris-ltd/eris-cli/config"
	def "github.com/eris-ltd/eris-cli/definitions"
	ini "github.com/eris-ltd/eris-cli/initialize"
	"github.com/eris-ltd/eris-cli/util"

	"github.com/eris-ltd/common/go/common"
	log "github.com/eris-ltd/eris-logger"
	docker "github.com/fsouza/go-dockerclient"
)

var (
	ErisDir = filepath.Join(os.TempDir(), "eris")

	ErrContainerExistMismatch = errors.New("container existence status check mismatch")
	ErrContainerRunMismatch   = errors.New("container run status check mismatch")
)

const (
	ConnectAndPull = iota
	DontPull
	Quick
)

func TestsInit(steps int, services ...string) (err error) {
	common.ChangeErisRoot(ErisDir)
	common.InitErisDir()

	config.Global, err = config.New(os.Stdout, os.Stderr)
	if err != nil {
		IfExit(fmt.Errorf("TRAGIC. Could not set global config.\n"))
	}

	// Don't connect to Docker daemon and don't pull default definitions.
	if steps == Quick {
		return nil
	}

	util.DockerConnect(false, "eris")

	// Don't pull default definition files.
	if steps == DontPull {
		return nil
	}

	// This dumps the ipfs and keys services defs into the temp dir which
	// has been set as the erisRoot.
	do := def.NowDo()
	do.Pull = false //don't pull imgs
	do.Yes = true   //over-ride command-line prompts
	do.Quiet = true
	do.ServicesSlice = services
	if err := ini.Initialize(do); err != nil {
		IfExit(fmt.Errorf("TRAGIC. Could not initialize the eris dir: %s.\n", err))
	}

	log.Info("Test init completed. Starting main test sequence now")
	return nil
}

func TestExistAndRun(name, t string, toExist, toRun bool) error {
	log.WithFields(log.Fields{
		"=>":       name,
		"running":  toRun,
		"existing": toExist,
	}).Info("Checking container")

	if existing := util.Exists(t, name); existing != toExist {
		log.WithFields(log.Fields{
			"=>":       name,
			"expected": toExist,
			"got":      existing,
		}).Info("Checking container existing")

		return ErrContainerExistMismatch
	}

	if running := util.Running(t, name); running != toRun {
		log.WithFields(log.Fields{
			"=>":       name,
			"expected": toExist,
			"got":      running,
		}).Info("Checking container running")

		return ErrContainerRunMismatch
	}

	return nil
}

func TestNumbersExistAndRun(servName string, containerExist, containerRun bool) error {
	log.WithFields(log.Fields{
		"=>":        servName,
		"existing#": containerExist,
		"running#":  containerRun,
	}).Info("Checking number of containers for")

	log.WithField("=>", servName).Debug("Checking existing containers for")
	exist := util.Exists(def.TypeService, servName)

	log.WithField("=>", servName).Debug("Checking running containers for")
	run := util.Running(def.TypeService, servName)

	if exist != containerExist {
		log.WithFields(log.Fields{
			"name":     servName,
			"expected": containerExist,
			"got":      exist,
		}).Info("Wrong number of existing containers")
		return fmt.Errorf("Wrong number of existing containers")
	}

	if run != containerRun {
		log.WithFields(log.Fields{
			"name":     servName,
			"expected": containerExist,
			"got":      run,
		}).Info("Wrong number of running containers")
		return fmt.Errorf("Wrong number of existing containers")
	}
	log.Info("All good")
	return nil
}

// Remove a container of some name, type, and number.
func RemoveContainer(name, t string) error {
	opts := docker.RemoveContainerOptions{
		ID:            util.ContainerName(t, name),
		RemoveVolumes: true,
		Force:         true,
	}

	if err := util.DockerClient.RemoveContainer(opts); err != nil {
		return err
	}

	return nil
}

// Remove everything Eris.
func RemoveAllContainers() error {
	toClean := map[string]bool{
		"yes":        true,
		"containers": true,
		"scratch":    true,
		"chains":     true,
		"all":        false,
		"rmd":        false,
		"images":     false,
	}
	return util.Clean(toClean)
}

// Return container links. For sake of simplicity, don't expose
// anything else.
func Links(name, t string) []string {
	container, err := util.DockerClient.InspectContainer(util.ContainerName(t, name))
	if err != nil {
		return []string{}
	}
	return container.HostConfig.Links
}

// Write a fake service definition file in a tmpDir Eris home directory.
func FakeServiceDefinition(name, definition string) error {
	return FakeDefinitionFile(common.ServicesPath, name, definition)
}

// Write a fake definition file in a tmpDir Eris home directory.
func FakeDefinitionFile(tmpDir, name, definition string) error {
	if !util.DoesDirExist(tmpDir) {
		if err := os.MkdirAll(tmpDir, 0755); err != nil {
			return err
		}
	}

	filename := filepath.Join(tmpDir, name+".toml")
	out, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = out.WriteString(definition)
	if err != nil {
		return err
	}

	return err
}

// Remove the Docker image. A wrapper over Docker client's library.
func RemoveImage(name string) error {
	return util.DockerClient.RemoveImage(name)
}

// FileContents returns the contents of a file as a string
// or panics on error.
func FileContents(filename string) string {
	f, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	content, err := ioutil.ReadAll(f)
	if err != nil {
		panic(err)
	}

	return string(content)
}

// each package will need its own custom stuff if need be
// do it through a custom pre-process ifExit in each package that
// calls tests.IfExit()
func TestsTearDown() error {
	// Move out of ErisDir before deleting it.
	parentPath := filepath.Join(ErisDir, "..")
	os.Chdir(parentPath)

	if err := os.RemoveAll(ErisDir); err != nil {
		return err
	}
	return nil
}

func IfExit(err error) {
	if err != nil {
		log.Error(err)
		if err := TestsTearDown(); err != nil {
			log.Error(err)
		}
		os.Exit(1)
	}
}

//------- helpers --------
func checkIPFSnotRunning() {
	//os.Setenv("ERIS_IPFS_HOST", "http://0.0.0.0") //conflicts with docker-machine
	do := def.NowDo()
	do.Known = false
	do.Existing = false
	do.Running = true
	do.Quiet = true
	log.Debug("Finding the running services")
	if util.IsService("ipfs", true) {
		IfExit(fmt.Errorf("IPFS service is running.\nPlease stop it with.\neris services stop -rx ipfs\n"))
	}
	// make sure ipfs container does not exist
	do = def.NowDo()
	do.Known = false
	do.Existing = true
	do.Running = false
	do.Quiet = true
	log.Debug("Finding the existing services")
	if util.IsService("ipfs", false) {
		IfExit(fmt.Errorf("IPFS service exists.\nPlease remove it with\neris services rm ipfs\n"))
	}
}
