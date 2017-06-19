package util

import (
	"bytes"
	"fmt"
	"net"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/monax/monax/definitions"
	"github.com/monax/monax/log"
	"github.com/monax/monax/version"

	docker "github.com/fsouza/go-dockerclient"
)

// Docker Client initialization
var DockerClient *docker.Client

// XXX [zr] machName should be default always and ultimately we should remove
// and force users to run eval if not using DFM/DFW
// TODO: return an error...?
// and refator this function generally
func DockerConnect(verbose bool, machName string) {
	var err error
	var dockerHost string
	var dockerCertPath string

	// This means we aren't gonna use docker-machine (kind of).
	if (machName == "monax" || machName == "default") && (os.Getenv("DOCKER_HOST") == "" && os.Getenv("DOCKER_CERT_PATH") == "") {
		//if os.Getenv("DOCKER_HOST") == "" && os.Getenv("DOCKER_CERT_PATH") == "" {
		endpoint := "unix:///var/run/docker.sock"

		log.WithField("=>", endpoint).Debug("Checking Linux Docker socket")
		u, _ := url.Parse(endpoint)
		_, err := net.Dial(u.Scheme, u.Path)
		if err != nil {
			IfExit(fmt.Errorf("%v\n", mustInstallError()))
		}
		log.WithField("=>", endpoint).Debug("Connecting to Docker")
		DockerClient, err = docker.NewClient(endpoint)
		if err != nil {
			IfExit(DockerError(mustInstallError()))
		}
	} else {
		log.WithFields(log.Fields{
			"host":      os.Getenv("DOCKER_HOST"),
			"cert path": os.Getenv("DOCKER_CERT_PATH"),
		}).Debug("Getting connection details from environment")
		log.WithField("machine", machName).Debug("Getting connection details from Docker Machine")

		dockerHost, dockerCertPath, err = getMachineDeets(machName) // machName is "default" (or eval was used)
		if err != nil {
			log.Debug("Could not connect to Docker Machine")
			IfExit(mustInstallError())
		}

		log.WithFields(log.Fields{
			"host":      dockerHost,
			"cert path": dockerCertPath,
		}).Debug()

		if err := connectDockerTLS(dockerHost, dockerCertPath); err != nil {
			IfExit(fmt.Errorf("Error connecting to Docker Backend via TLS:\nError:%v\n", err))
		}
		log.Debug("Successfully connected to Docker daemon")
	}
}

func CheckDockerClient() error {
	if runtime.GOOS == "linux" {
		return nil
	}

	dockerHost, dockerCertPath := popHostAndPath()

	if dockerCertPath == "" || dockerHost == "" {
		// TODO [zr] there is probably nicer login we can go with generally
		if _, _, err := getMachineDeets("default"); err == nil {
			fmt.Println("A Docker Machine VM named 'default' exists, which Monax can use")
			fmt.Println("Please run [eval $(docker-machine env default)] to get sorted")
		} else {
			fmt.Println("The marmots could not find a Docker Machine VM they could connect to")
			fmt.Println("Please set one up (see https://monax.io/docs/getting-started)")
		}
	}

	return nil
}

func getMachineDeets(machName string) (string, string, error) {
	var out = new(bytes.Buffer)
	var out2 = new(bytes.Buffer)

	noConnectError := fmt.Errorf("Could not evaluate the env vars for the %s docker-machine.\n", machName)
	dHost, dPath := popHostAndPath()

	if (dHost != "" && dPath != "") && (machName == "monax" || machName == "default") {
		return dHost, dPath, nil
	}

	// TODO: when go-dockerclient adds machine API endpoints use those instead.
	log.WithField("machine", machName).Debug("Querying Docker Machine URL")
	cmd := exec.Command("docker-machine", "url", machName)
	cmd.Stdout = out
	if err := cmd.Run(); err != nil {
		return "", "", fmt.Errorf("%vError:\t%v\n", noConnectError, err)
	}
	dHost = strings.TrimSpace(out.String())
	log.WithField("host", dHost).Debug()

	// TODO: when go-dockerclient adds machine API endpoints use those instead.
	log.WithField("machine", machName).Debug("Querying Docker Machine cert path")
	cmd2 := exec.Command("docker-machine", "inspect", machName, "--format", "{{.HostOptions.AuthOptions.ServerCertPath}}")
	cmd2.Stdout = out2
	//cmd2.Stderr = os.Stderr
	if err := cmd2.Run(); err != nil {
		return "", "", fmt.Errorf("%vError:\t%v\n", noConnectError, err)
	}
	dPath = out2.String()
	dPath = strings.Replace(dPath, "'", "", -1)
	dPath = filepath.Dir(dPath)
	log.WithField("cert path", dPath).Debug()

	if dPath == "" || dHost == "" {
		return "", "", noConnectError
	}

	log.Info("Querying host and user have access to the right files for TLS connection to Docker")
	if err := checkKeysAndCerts(dPath); err != nil {
		return "", "", err
	}
	log.Debug("Certificate files look good")

	// technically, do not *have* to do this, but it will make repetitive tasks faster
	log.Debug("Setting environment variables for quick future development")
	os.Setenv("DOCKER_HOST", dHost)
	os.Setenv("DOCKER_CERT_PATH", dPath)
	os.Setenv("DOCKER_TLS_VERIFY", "1")
	os.Setenv("DOCKER_MACHINE_NAME", machName)

	log.WithField("machine", machName).Debug("Finished getting machine details")
	return dHost, dPath, nil
}

func DockerClientVersion() (string, error) {
	info, err := DockerClient.Version()
	if err != nil {
		return "", DockerError(err)
	}

	return info.Get("Version"), nil
}

func DockerAPIVersion() (string, error) {
	info, err := DockerClient.Version()
	if err != nil {
		return "", DockerError(err)
	}

	return info.Get("APIVersion"), nil
}

// IsMinimalDockerClientVersion returns true if the connected Docker client
// version is at least equal to the mimimal required for Monax.
func IsMinimalDockerClientVersion() bool {
	v, err := DockerClientVersion()
	if err != nil {
		return false
	}
	return CompareVersions(v, version.DOCKER_VER_MIN)
}

// CompareVersions returns true if the version1 is larger or equal the version2,
// for example CompareVersions("1.10", "1.9") returns true.
func CompareVersions(version1, version2 string) bool {
	v1 := strings.Split(version1, ".")
	v2 := strings.Split(version2, ".")

	// Comparing just the major.minor scheme versions against each other (like "1.8").
	if len(v1) < 2 || len(v2) < 2 {
		return false
	}

	major1, err := strconv.Atoi(v1[0])
	if err != nil {
		return false
	}
	minor1, err := strconv.Atoi(v1[1])
	if err != nil {
		return false
	}

	major2, err := strconv.Atoi(v2[0])
	if err != nil {
		return false
	}
	minor2, err := strconv.Atoi(v2[1])
	if err != nil {
		return false
	}

	// Comparing major versions.
	if major1 < major2 {
		return false
	}
	if major1 > major2 {
		return true
	}

	// Majors are equal. Comparing minor versions.
	if minor1 < minor2 {
		return false
	}

	// Otherwise true.

	// NOTE: this means that CompareVersions("1.9.19", "1.9.23") will return true,
	// because "1.9" equals "1.9".

	return true
}

func connectDockerTLS(dockerHost, dockerCertPath string) error {
	var err error

	log.WithFields(log.Fields{
		"host":      dockerHost,
		"cert path": dockerCertPath,
	}).Debug("Connecting to Docker via TLS")
	DockerClient, err = docker.NewTLSClient(dockerHost, filepath.Join(dockerCertPath, "cert.pem"), filepath.Join(dockerCertPath, "key.pem"), filepath.Join(dockerCertPath, "ca.pem"))
	if err != nil {
		return DockerError(err)
	}

	log.Debug("Connected via TLS")
	return nil
}

func popHostAndPath() (string, string) {
	return os.Getenv("DOCKER_HOST"), os.Getenv("DOCKER_CERT_PATH")
}

func checkKeysAndCerts(dPath string) error {
	toCheck := []string{"cert.pem", "key.pem", "ca.pem"}
	for _, f := range toCheck {
		f = filepath.Join(dPath, f)
		if _, err := os.Stat(f); err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("The marmots could not find a file that was required to connect to Docker.\nThey get a file does not exist error from the OS.\nFile needed:\t%s\n", f)
			} else if os.IsNotExist(err) {
				return fmt.Errorf("The marmots could not find a file that was required to connect to Docker.\nThey get a permissions error for the file.\nPlease check your file permissions.\nFile needed:\t%s\n", f)
			} else {
				return fmt.Errorf("The marmots could not find a file that was required to connect to Docker.\nThe file exists and the user has the right permissions.\nColor the marmots confused.\nFile needed:\t%s\nError:\t%v\n", f, err)
			}
		}
	}
	return nil
}

func mustInstallError() error {
	install := `The marmots cannot connect to Docker. Do you have Docker installed?
If not, please visit here: https://docs.docker.com/engine/installation/`

	if runtime.GOOS == "linux" {
		run := `Do you have Docker running? If not, please type [sudo service docker start].
Also check that your user is in the "docker" group. If not, you can add it
using the [sudo usermod -a -G docker $USER] command or rerun as [sudo monax]`

		return fmt.Errorf("%s\n\n%s", install, run)
	} else {
		return fmt.Errorf("%s\n\n%s", install, "Note that [monax] does not yet support Docker For Mac/Windows. These platforms require docker-machine.")
	}
	return fmt.Errorf(install)
}

func DockerError(err error) error {
	if _, ok := err.(*docker.Error); ok {
		return fmt.Errorf("Docker: %v", err.(*docker.Error).Message)
	}
	return err
}

func DockerWindowsAndMacIP(do *definitions.Do) (string, error) {
	if runtime.GOOS == "windows" || runtime.GOOS == "darwin" {
		if os.Getenv("DOCKER_HOST") == "" && os.Getenv("DOCKER_CERT_PATH") == "" {
			return "127.0.0.1", nil
		} else {
			cmd := exec.Command("docker-machine", "ip", "default")
			if out, err := cmd.CombinedOutput(); err != nil {
				return "", fmt.Errorf("%v", DockerError(err))
			} else {
				return strings.TrimSpace(string(out)), nil
			}
		}
	} else {
		return "", fmt.Errorf("Linux machine, should not reach here")
	}
}
