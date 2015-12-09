package testutils

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/fsouza/go-dockerclient"
	"github.com/eris-ltd/eris-cli/config"
	def "github.com/eris-ltd/eris-cli/definitions"
	ini "github.com/eris-ltd/eris-cli/initialize"
	"github.com/eris-ltd/eris-cli/util"

	log "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/Sirupsen/logrus"
)

var erisDir = path.Join(os.TempDir(), "eris")

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
	config.ChangeErisDir(erisDir)

	util.DockerConnect(false, "eris")

	// this dumps the ipfs and keys services defs into the temp dir which
	// has been set as the erisRoot.
	do := def.NowDo()
	do.Pull = false
	do.Yes = true
	do.Quiet = true
	do.Source = "toadserver"
	if err := ini.Initialize(do); err != nil {
		IfExit(fmt.Errorf("TRAGIC. Could not initialize the eris dir.\n"))
	}

	log.Info("Test init completed. Starting main test sequence now")
	return nil
}

//return to handle failings in each pkg
//typ = type of test for dealing with do.() details
func TestExistAndRun(name, typ string, contNum int, toExist, toRun bool) bool {
	var exist, run bool
	if typ == "actions" {
		name = strings.Replace(name, " ", "_", -1) // dirty
	}

	log.WithFields(log.Fields{
		"=>":       name,
		"running":  toRun,
		"existing": toExist,
	}).Info("Checking container")
	if typ == "chains" {
		name = util.ChainContainersName(name, 1) // not worried about containerNumbers, deal with multiple containers in services tests
	} else if typ == "services" {
		name = util.ServiceContainersName(name, contNum)

	} else {
		name = util.DataContainersName(name, 1)
	}
	do := def.NowDo()
	do.Quiet = true
	do.Operations.Args = []string{"testing"}

	if typ == "data" || typ == "chains" || typ == "services" {
		do.Existing = true
	} else if typ == "actions" {
		do.Known = true
	}
	if err := util.ListAll(do, typ); err != nil {
		log.Error(err)
		return true
	}

	res := strings.Split(do.Result, "\n")
	for _, r := range res {
		if r == util.ContainersShortName(name) {
			exist = true
		}
	}

	if toExist != exist {
		if toExist {
			log.WithField("=>", name).Info("Could not find existing")
		} else {
			log.WithField("=>", name).Info("Found existing when shouldn't be")
		}
		return true
	}
	//func should always be testing for toExist, only sometimes tested for runining
	if typ == "chains" || typ == "services" {
		do.Running = true
		do.Existing = false //unset
		if err := util.ListAll(do, typ); err != nil {
			return true
		}
		res = strings.Split(do.Result, "\n")
		for _, r := range res {
			if r == util.ContainersShortName(name) {
				run = true
			}
		}

		if toRun != run {
			if toRun {
				log.WithField("=>", name).Info("Could not find running")
			} else {
				log.WithField("=>", name).Info("Found running when shouldn't be")
			}
			return true
		}
	}

	return false
}

// Remove a container of some name, type, and number.
func RemoveContainer(name string, t string, n int) error {
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
func Links(name string, t string, n int) []string {
	container, err := util.DockerClient.InspectContainer(util.ContainersName(t, name, n))
	if err != nil {
		return []string{}
	}
	return container.HostConfig.Links
}

// Write a fake service definition file in a tmpDir Eris home directory.
func FakeServiceDefinition(tmpDir, name, definition string) error {
	filename := filepath.Join(tmpDir, "services", name+".toml")

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
	if err := os.RemoveAll(erisDir); err != nil {
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
	do.Operations.Args = []string{"testing"}
	log.Debug("Finding the running services")
	if err := util.ListAll(do, "services"); err != nil {
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
	do.Operations.Args = []string{"testing"}
	log.Debug("Finding the existing services")
	if err := util.ListAll(do, "services"); err != nil {
		IfExit(err)
	}
	res = strings.Split(do.Result, "\n")
	for _, r := range res {
		if r == "ipfs" {
			IfExit(fmt.Errorf("IPFS service exists.\nPlease remove it with\neris services rm ipfs\n"))
		}
	}
}
