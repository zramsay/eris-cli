package util

import (
	"bytes"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path"
	"runtime"

	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/fsouza/go-dockerclient"
)

// Docker Client initialization
var DockerClient *docker.Client

func DockerConnect(verbose bool) { // TODO: return an error...?
	var err error

	if runtime.GOOS == "linux" {
		endpoint := "unix:///var/run/docker.sock"

		logger.Debugln("Connecting to the Docker Client via:", endpoint)
		DockerClient, err = docker.NewClient(endpoint)
		if err != nil {
			logger.Printf("%v\n", mustInstallError())
			os.Exit(1)
		}

		logger.Debugln("Successfully connected to Docker daemon.")

	} else {

		var dockerHost string
		var dockerCertPath string

		dockerHost, dockerCertPath, err = getMachineDeets("eris") // we'd want this to be a flag in the future

		if err != nil {
			logger.Debugf("Could not connect to the eris docker-machine. Trying default docker-machine.\n")
			dockerHost, dockerCertPath, err = getMachineDeets("default") // during toolbox setup this is the machine that is created
			if err != nil {
				logger.Debugf("Could not connect to the default docker-machine. Trying to set up a new machine.\n")
				if e2 := CheckDockerClient(); e2 != nil {
					logger.Printf("%v\n", e2)
					os.Exit(1)
				}
				dockerHost, dockerCertPath, _ = getMachineDeets("eris")
			}
		}

		logger.Debugln("Connecting to the Docker Client via:", dockerHost)
		logger.Debugln("Docker Certificate Path:", dockerCertPath)

		DockerClient, err = docker.NewTLSClient(dockerHost, path.Join(dockerCertPath, "cert.pem"), path.Join(dockerCertPath, "key.pem"), path.Join(dockerCertPath, "ca.pem"))
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		logger.Debugln("Successfully connected to Docker daemon")
		logger.Debugln("Setting IPFS Host")
		os.Setenv("ERIS_IPFS_HOST", dockerHost)
	}
}

func CheckDockerClient() error {
	if runtime.GOOS == "linux" {
		_, err := net.Dial("unix", "/var/run/docker.sock")
		if err != nil {
			return mustInstallError()
		}
	} else {
		dockerHost, dockerCertPath := popPathAndHost()

		if dockerCertPath == "" || dockerHost == "" {
			driver := "virtualbox" // when we use agents we'll wanna turn this driver into a flag

			if runtime.GOOS == "windows" {
				if err := prepWin(); err != nil {
					return fmt.Errorf("Could not add ssh.exe to PATH.\nError:%v\n", err)
				}
			}

			if _, _, err := getMachineDeets("default"); err == nil {

				var input string
				fmt.Print("A docker-machine exists, which eris can use.\nHowever, our marmots recommend that you have a vm dedicated to eris dev-ing.\nWould you like the marmots to create a machine for you? (Y/n): ")
				fmt.Scanln(&input)

				if input == "Y" || input == "y" || input == "YES" || input == "Yes" || input == "yes" {
					logger.Infof("The marmots will create an eris machine.\n")
					if err := setupErisMachine(driver); err != nil {
						return err
					}

					logger.Debugf("New docker machine created using %s driver. Getting the proper environment variables.\n", driver)
					if _, _, err := getMachineDeets("eris"); err != nil {
						return err
					}
				} else {
					logger.Infof("No eris docker-machine will be created.")
				}

			} else {
				logger.Debugf("Could not find DOCKER_HOST or DOCKER_CERT_PATH. The marmots will create an eris machine.\n")
				if err := setupErisMachine(driver); err != nil {
					return err
				}

				logger.Debugf("New docker machine created using %s driver. Getting the proper environment variables.\n", driver)
				if _, _, err := getMachineDeets("eris"); err != nil {
					return err
				}
			}
		}
	}

	logger.Infof("Docker client connects correctly.\n")
	return nil
}

func popPathAndHost() (string, string) {
	return os.Getenv("DOCKER_HOST"), os.Getenv("DOCKER_CERT_PATH")
}

func setupErisMachine(driver string) error {
	cmd := exec.Command("docker-machine", "create", "--driver", driver, "eris")
	if err := cmd.Run(); err != nil {
		logger.Debugf("There was an error creating a new docker-machine.\nError:\t%v\n", err)
		return mustInstallError()
	}
	return nil
}

func getMachineDeets(machName string) (string, string, error) {
	var out bytes.Buffer
	noConnectError := fmt.Errorf("Could not evaluate the env vars for the %s docker-machine.\n", machName)
	dPath, dHost := popPathAndHost()

	if dPath != "" && dHost != "" {
		return dPath, dHost, nil
	}

	// TODO: when go-dockerclient adds machine API endpoints use those instead.
	cmd := exec.Command("docker-machine", "url", machName)
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return "", "", fmt.Errorf("%v\nError:\t\n", noConnectError, err)
	}
	dPath = out.String()

	// TODO: when go-dockerclient adds machine API endpoints use those instead.
	cmd = exec.Command("docker-machine", "inspect", "--format='{{.Driver.CaCertPath}}'", machName)
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return "", "", fmt.Errorf("%v\nError:\t\n", noConnectError, err)
	}
	dHost = out.String()
	dHost = path.Dir(dHost)

	if dPath == "" || dHost == "" {
		return "", "", noConnectError
	}

	// technically, do not *have* to do this, but it will make repetitive tasks faster
	os.Setenv("DOCKER_HOST", dHost)
	os.Setenv("DOCKER_CERT_PATH", dPath)
	os.Setenv("DOCKER_TLS_VERIFY", "1")
	os.Setenv("DOCKER_MACHINE_NAME", machName)

	return dPath, dHost, nil
}

func mustInstallError() error {
	errBase := "The marmots cannot connect to Docker.\nDo you have docker installed?\nIf not please visit here:\t"
	dInst := "https://docs.docker.com/installation/"

	switch runtime.GOOS {
	case "linux":
		return fmt.Errorf("%s%s\nDo you have docker running?\nIf not please [sudo services start docker] on Ubuntu.\n", errBase, dInst)
	case "darwin":
		return fmt.Errorf("%s%s\n", errBase, (dInst + "mac/"))
	case "windows":
		return fmt.Errorf("%s%s\n", errBase, (dInst + "windows/"))
	default:
		return fmt.Errorf("%s%s\n", errBase, dInst)
	}

	return nil
}

// need to add ssh.exe to PATH, it resides in GIT dir.
// see: https://docs.docker.com/installation/windows/#from-your-shell
func prepWin() error {
	// note this is for running from cmd.exe ... watch out for powershell....
	cmd := exec.Command("set", `PATH=%PATH%;"c:\Program Files (x86)\Git\bin"`)
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}
