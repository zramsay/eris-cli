package tests

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/eris-ltd/eris-cli/config"
	def "github.com/eris-ltd/eris-cli/definitions"
	ini "github.com/eris-ltd/eris-cli/initialize"
	"github.com/eris-ltd/eris-cli/list"
	"github.com/eris-ltd/eris-cli/util"

	log "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/Sirupsen/logrus"
	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common/go/common"
	docker "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/fsouza/go-dockerclient"
)

var (
	ErisDir = filepath.Join(os.TempDir(), "eris")

	ErrContainerExistMismatch = errors.New("container existence status check mismatch")
	ErrContainerRunMismatch   = errors.New("container run status check mismatch")
)

//testType = one of each package, will switch over it for
//make additional tempDirs and vars as needed -> [zr] or not, TBD
func TestsInit(testType string) (err error) {
	// TODO: make a reader/pipe so we can see what is written from tests.
	config.GlobalConfig, err = config.SetGlobalObject(os.Stdout, os.Stderr)
	if err != nil {
		IfExit(fmt.Errorf("TRAGIC. Could not set global config.\n"))
	}

	// common is initialized on import so
	// we have to manually override these
	// variables to ensure that the tests
	// run correctly.
	config.ChangeErisDir(ErisDir)
	common.InitErisDir()
	util.DockerConnect(false, "eris")

	// this dumps the ipfs and keys services defs into the temp dir which
	// has been set as the erisRoot.
	do := def.NowDo()
	do.Pull = false //don't pull imgs
	do.Yes = true   //over-ride command-line prompts
	do.Quiet = true
	do.Source = "rawgit" //use "rawgit" if ts down
	if err := ini.Initialize(do); err != nil {
		IfExit(fmt.Errorf("TRAGIC. Could not initialize the eris dir: %s.\n", err))
	}

	log.Info("Test init completed. Starting main test sequence now")
	return nil
}

func TestActionDefinitionFile(name string) bool {
	name = strings.Replace(name, " ", "_", -1)

	if util.GetFileByNameAndType("actions", name) == "" {
		return false
	}
	return true
}

func TestExistAndRun(name, t string, contNum int, toExist, toRun bool) error {
	log.WithFields(log.Fields{
		"=>":       name,
		"running":  toRun,
		"existing": toExist,
	}).Info("Checking container")

	if existing := FindContainer(name, t, contNum, false); existing != toExist {
		log.WithFields(log.Fields{
			"=>":       name,
			"expected": toExist,
			"got":      existing,
		}).Info("Checking container existing")

		return ErrContainerExistMismatch
	}

	if running := FindContainer(name, t, contNum, true); running != toRun {
		log.WithFields(log.Fields{
			"=>":       name,
			"expected": toExist,
			"got":      running,
		}).Info("Checking container running")

		return ErrContainerRunMismatch
	}

	return nil
}

// FindContainer returns true if the container with a given
// short name, type, number, and status exists.
func FindContainer(name, t string, n int, running bool) bool {
	containers := util.ErisContainersByType(t, !running)

	for _, c := range containers {
		if c.ShortName == name && c.Number == n {
			return true
		}
	}
	return false
}

// Remove a container of some name, type, and number.
func RemoveContainer(name, t string, n int) error {
	opts := docker.RemoveContainerOptions{
		ID:            util.ContainersName(t, name, n),
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
	return util.Clean(false, false, false, false)
}

// Return container links. For sake of simplicity, don't expose
// anything else.
func Links(name, t string, n int) []string {
	container, err := util.DockerClient.InspectContainer(util.ContainersName(t, name, n))
	if err != nil {
		return []string{}
	}
	return container.HostConfig.Links
}

// Write a fake service definition file in a tmpDir Eris home directory.
func FakeServiceDefinition(tmpDir, name, definition string) error {
	return FakeDefinitionFile(filepath.Join(tmpDir, "services"), name, definition)
}

// Write a fake definition file in a tmpDir Eris home directory.
func FakeDefinitionFile(tmpDir, name, definition string) error {
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

// each pacakge will need its own custom stuff if need be
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
	if err := list.ListAll(do, "services"); err != nil {
		IfExit(err)
	}
	res := strings.Split(do.Result, "\n")
	for _, r := range res {
		if r == "ipfs" {
			IfExit(fmt.Errorf("IPFS service is running.\nPlease stop it with.\neris services stop -rx ipfs\n"))
		}
	}
	// make sure ipfs container does not exist
	do = def.NowDo()
	do.Known = false
	do.Existing = true
	do.Running = false
	do.Quiet = true
	log.Debug("Finding the existing services")
	if err := list.ListAll(do, "services"); err != nil {
		IfExit(err)
	}
	res = strings.Split(do.Result, "\n")
	for _, r := range res {
		if r == "ipfs" {
			IfExit(fmt.Errorf("IPFS service exists.\nPlease remove it with\neris services rm ipfs\n"))
		}
	}
}
