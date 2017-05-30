package testutil

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/monax/monax/config"
	"github.com/monax/monax/definitions"
	"github.com/monax/monax/initialize"
	"github.com/monax/monax/log"
	"github.com/monax/monax/util"

	docker "github.com/fsouza/go-dockerclient"
)

var (
	TmpMonaxRoot = filepath.Join(os.TempDir(), "monax")

	ErrContainerExistMismatch = errors.New("container existence status check mismatch")
	ErrContainerRunMismatch   = errors.New("container run status check mismatch")
	ErrUnsupportedType        = errors.New("expected a Pull struct as a parameter")
)

// Pull type is used as an argument to Init function:
// which definitions and services to pull.
type Pull struct {
	Images   []string
	Services []string
}

// Init initializes environment to run Docker related package tests.
// It accepts either none or Pull struct as arguments.
//
//  Init()
//    - connect to Docker
//
//  Init(Pull{Services: []string{...}})
//    - connect to Docker and pull selected service definition files
//
//  Init(Pull{Services: []string{...}, Images: []string{...}})
//    - connect to Docker and pull select images and service
//      definition files.
//
//  Init(Pull{})
//    - connect to Docker and pull all service definition files
//      (unspecified Services means all services, not none).
//
func Init(args ...interface{}) (err error) {
	config.ChangeMonaxRoot(TmpMonaxRoot)
	config.InitMonaxDir()

	config.Global, err = config.New(os.Stdout, os.Stderr)
	if err != nil {
		IfExit(fmt.Errorf("Could not set global config"))
	}

	util.DockerConnect(false, "monax")

	// Just connect.
	if len(args) == 0 {
		return nil
	}

	do := definitions.NowDo()
	do.Yes = true
	do.Quiet = true

	pull, ok := args[0].(Pull)
	if !ok {
		return ErrUnsupportedType
	}

	do.ServicesSlice = pull.Services
	do.ImagesSlice = pull.Images

	if len(pull.Images) != 0 {
		do.Pull = true
	}

	os.Setenv("MONAX_PULL_APPROVE", "true")

	if err := initialize.Initialize(do); err != nil {
		IfExit(fmt.Errorf("Could not initialize Monax root: %v", err))
	}

	log.Info("Test init completed. Starting main test sequence now")

	return nil
}

func ExistAndRun(name, t string, toExist, toRun bool) error {
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

func NumbersExistAndRun(servName string, containerExist, containerRun bool) error {
	log.WithFields(log.Fields{
		"=>":        servName,
		"existing#": containerExist,
		"running#":  containerRun,
	}).Info("Checking number of containers for")

	log.WithField("=>", servName).Debug("Checking existing containers for")
	exist := util.Exists(definitions.TypeService, servName)

	log.WithField("=>", servName).Debug("Checking running containers for")
	run := util.Running(definitions.TypeService, servName)

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

	return util.DockerClient.RemoveContainer(opts)
}

// Remove everything Monax.
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

// Write a fake service definition file in a tmpDir Monax home directory.
func FakeServiceDefinition(name, definition string) error {
	return FakeDefinitionFile(config.ServicesPath, name, definition)
}

// Write a fake definition file in a tmpDir Monax home directory.
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

// TearDown removes all Monax containers and temporary Monax root
// directory on exit.
func TearDown() error {
	// Move out of MonaxDir before deleting it.
	parentPath := filepath.Join(TmpMonaxRoot, "..")
	os.Chdir(parentPath)

	return os.RemoveAll(TmpMonaxRoot)
}

// IfExit exits with an exit code 1 if err is not nil.
func IfExit(err error) {
	if err != nil {
		log.Error(err)
		if err := TearDown(); err != nil {
			log.Error(err)
		}
		os.Exit(1)
	}
}
