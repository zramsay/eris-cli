package main

import (
	"bytes"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/code.google.com/p/go-uuid/uuid"
)

var docker18s = []string{"1.8.0", "1.8.1", "1.8.2", "1.8.3"}
var docker19s = []string{"1.9.0", "1.9.1"}
var dockerAll = [][]string{docker18s, docker19s}

var dmDriver = "amazonec2"
var script = "docker.sh"
var vars map[string]string

var buildAllBranches = []string{"master", "staging", "develop"}
var maxTimeout = 15 * time.Minute

var wg sync.WaitGroup

func main() {
	if toShuffle() {
		for d := range dockerAll {
			shuffle(dockerAll[d])
		}
	}

	if err := vetAndPopulate(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	branch, err := getBranch()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	allOrFew := false
	for _, b := range buildAllBranches {
		if branch == b {
			allOrFew = true
		}
	}

	var machines []string

	if runtime.GOOS == "linux" {
		if allOrFew {
			machines = allBackends()
		} else {
			machines = curBackend()
		}
	} else {
		machines = curBackend()
	}

	failOut := make(chan bool, len(machines))
	go timeOutTicker(machines)

	wg.Add(len(machines))
	for _, m := range machines {
		go startMachine(m, failOut)
	}

	go func(failOut chan bool) {
		if _, ok := <-failOut; ok {
			for _, m := range machines {
				fmt.Println(m)
			}
			os.Exit(1)
		}
	}(failOut)
	wg.Wait()

	for _, m := range machines {
		fmt.Println(m)
	}
}

func allBackends() []string {
	var machines []string
	for d := range dockerAll {
		machines = append(machines, makeMachName(dockerAll[d][0]))
	}
	return machines
}

func curBackend() []string {
	return []string{makeMachName(dockerAll[len(dockerAll)-1][0])}
}

func timeOutTicker(machines []string) {
	time.Sleep(maxTimeout)
	for _, m := range machines {
		fmt.Println(m)
	}
	os.Exit(1)
}

func toShuffle() bool {
	t := time.Now()
	rand.Seed(int64(t.Nanosecond()))
	return rand.Int()%4 != 0
}

func shuffle(arr []string) {
	t := time.Now()
	rand.Seed(int64(t.Nanosecond()))
	for i := len(arr) - 1; i > 0; i-- {
		j := rand.Intn(i)
		arr[i], arr[j] = arr[j], arr[i]
	}
}

func vetAndPopulate() error {
	vars = make(map[string]string)

	// aws vars
	vars["akey"] = os.Getenv("AWS_ACCESS_KEY_ID")
	vars["asec"] = os.Getenv("AWS_SECRET_ACCESS_KEY")
	vars["avpc"] = os.Getenv("AWS_VPC_ID")
	vars["agrp"] = "eris-test-ire"
	vars["areg"] = "eu-west-1"
	switch runtime.GOOS {
	case "windows":
		vars["azon"] = "b"
	case "darwin":
		vars["azon"] = "b"
	default:
		vars["azon"] = "a"
	}

	// set aws default vars into env
	os.Setenv("AWS_VPC_ID", vars["avpc"])
	os.Setenv("AWS_SECURITY_GROUP", vars["agrp"])
	os.Setenv("AWS_DEFAULT_REGION", vars["areg"])
	os.Setenv("AWS_ZONE", vars["azon"])

	// this setting is atrocious in AWS
	os.Setenv("AWS_SSH_USER", "")

	// check populated based on driver
	for k, v := range vars {
		if k[0] == dmDriver[0] {
			if !checkExists(v) {
				return fmt.Errorf("Variable (%s) does not exist. Cannot proceed", k)
			}
		}
	}

	// all set
	return nil
}

func checkExists(toTest string) bool {
	if toTest == "" {
		return false
	}
	return true
}

func uuidMake() string {
	return strings.Split(uuid.New(), "-")[0]
}

func makeMachName(dockVer string) string {
	return strings.Join([]string{"eris", "test", runtime.GOOS, dockVer, uuidMake()}, "-")
}

func getBranch() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("Cannot find what branch this repository is in: (%s)", err)
	}
	return strings.TrimSpace(out.String()), nil
}

func startMachine(machine string, failOut chan<- bool) {
	defer wg.Done()
	if err := makeMachine(machine); err != nil {
		fmt.Fprintf(os.Stderr, "makeMachine error: %v\n", err)
		failOut <- true
		return
	}
	if err := setUpMachine(machine); err != nil {
		fmt.Fprintf(os.Stderr, "setUpMachine error: %v\n", err)
		failOut <- true
	}
}

func makeMachine(machine string) error {
	var cmd *exec.Cmd
	cmd = exec.Command("docker-machine", "create", "--driver", dmDriver, "--amazonec2-access-key", vars["akey"], "--amazonec2-secret-key", vars["asec"], "--amazonec2-region", vars["areg"], "--amazonec2-vpc-id", vars["avpc"], "--amazonec2-security-group", vars["agrp"], "--amazonec2-zone", vars["azon"], machine)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("Cannot make the machine (%s): (%s)\n\n%s", machine, err, out.String())
	}
	return nil
}

func setUpMachine(machine string) error {
	dir, err := os.Getwd()
	if err != nil {
		return err
	}
	file := filepath.Join(dir, script)
	cmd := exec.Command("docker-machine", "scp", file, fmt.Sprintf("%s:", machine))
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("Cannot scp into the machine (%s): (%s)\n\n%s", machine, err, out.String())
	}

	file = filepath.Base(file)
	cmd = exec.Command("docker-machine", "ssh", machine, fmt.Sprintf("sudo $HOME/%s", file))
	cmd.Stdout = &out
	cmd.Stderr = &out
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("Cannot execute the command to change docker daemon on machine (%s): (%s)\n\n%s", machine, err, out.String())
	}
	return nil
}
