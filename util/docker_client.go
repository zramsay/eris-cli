package util

import (
	"bytes"
	"fmt"
	"net"
	"net/url"
	"os"
	"os/exec"
	"path"
	"runtime"
	"strconv"
	"strings"

	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/fsouza/go-dockerclient"

	. "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common/go/common"
)

// Docker Client initialization
var DockerClient *docker.Client

func DockerConnect(verbose bool, machName string) { // TODO: return an error...?
	var err error
	var dockerHost string
	var dockerCertPath string
	if runtime.GOOS == "linux" {
		// this means we aren't gonna use docker-machine (kind of)
		if (machName == "eris" || machName == "default") && (os.Getenv("DOCKER_HOST") == "" && os.Getenv("DOCKER_CERT_PATH") == "") {
			//if os.Getenv("DOCKER_HOST") == "" && os.Getenv("DOCKER_CERT_PATH") == "" {
			endpoint := "unix:///var/run/docker.sock"

			logger.Debugf("Checking The Linux Docker Socket =>\t%s\n", endpoint)
			u, _ := url.Parse(endpoint)
			_, err := net.Dial(u.Scheme, u.Path)
			if err != nil {
				IfExit(fmt.Errorf("%v\n", mustInstallError()))
			}

			logger.Debugf("Connecting to the Docker Client via =>\t%s\n", endpoint)
			DockerClient, err = docker.NewClient(endpoint)
			if err != nil {
				IfExit(fmt.Errorf("%v\n", mustInstallError()))
			}
		} else {
			//if machName !="eris/default" OR if docker env vars set (via, e.g., eval)
			logger.Debugf("Connecting to the Docker Client via =>\t%s:%s\n", os.Getenv("DOCKER_HOST"), os.Getenv("DOCKER_CERT_PATH"))
			logger.Debugf("Checking Details via Docker-Machine =>\t%s\n", machName)
			dockerHost, dockerCertPath, err = getMachineDeets(machName)
			if err != nil {
				IfExit(fmt.Errorf("Error getting Docker-Machine Details for connection over TLS.\nERROR =>\t\t\t%v\n\nEither re-run the command without a machine or correct your machine name.\n", err))
			}

			logger.Debugf("Connecting to the Docker Client via =>\t%s:%s\n", dockerHost, dockerCertPath)
			if err := connectDockerTLS(dockerHost, dockerCertPath); err != nil {
				IfExit(fmt.Errorf("Error connecting to Docker Backend over TLS.\nERROR =>\t\t\t%v\n", err))
			}
			logger.Debugln("Successfully connected to Docker daemon.")

			logger.Debugln("Setting IPFS Host")
			setIPFSHostViaDockerHost(dockerHost)
		}

		logger.Debugln("Successfully connected to Docker daemon.")

	} else {
		logger.Debugf("Connecting to the Docker Client via =>\t%s:%s\n", os.Getenv("DOCKER_HOST"), os.Getenv("DOCKER_CERT_PATH"))
		logger.Debugf("Checking Details via Docker-Machine =>\t%s\n", machName)
		dockerHost, dockerCertPath, err = getMachineDeets(machName) // machName is "eris" by default

		if err != nil {

			logger.Debugf("Could not connect to the eris docker-machine.\nError:\t%vTrying \"default\" docker-machine.\n", err)
			dockerHost, dockerCertPath, err = getMachineDeets("default") // during toolbox setup this is the machine that is created
			if err != nil {

				logger.Debugf("Could not connect to the \"default\" docker-machine.\nError:\t%vTrying to set up a new machine.\n", err)
				if e2 := CheckDockerClient(); e2 != nil {
					IfExit(fmt.Errorf("%v\n", e2))
				}
				dockerHost, dockerCertPath, _ = getMachineDeets("eris")
			}

		}

		logger.Debugf("Connecting to the Docker Client via =>\t%s:%s\n", dockerHost, dockerCertPath)
		if err := connectDockerTLS(dockerHost, dockerCertPath); err != nil {
			IfExit(fmt.Errorf("Error connecting to Docker Backend over TLS.\nERROR =>\t\t\t%v\n", err))
		}
		logger.Debugln("Successfully connected to Docker daemon.")

		logger.Debugln("Setting IPFS Host")
		setIPFSHostViaDockerHost(dockerHost)
	}
}

func CheckDockerClient() error {
	if runtime.GOOS == "linux" {
		return nil
	}

	var input string
	dockerHost, dockerCertPath := popHostAndPath()

	if dockerCertPath == "" || dockerHost == "" {
		driver := "virtualbox" // when we use agents we'll wanna turn this driver into a flag

		if runtime.GOOS == "windows" {
			if err := prepWin(); err != nil {
				return fmt.Errorf("Could not add ssh.exe to PATH.\nError:%v\n", err)
			}
		}

		if _, _, err := getMachineDeets("default"); err == nil {

			fmt.Print("A docker-machine virtual machine exists, which eris can use.\nHowever, our marmots recommend that you have a vm dedicated to eris dev-ing.\nWould you like the marmots to create a machine for you? (y/n): ")
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

			fmt.Print("The marmots could not find a docker-machine virtual machine they could connect to.\nOur marmots recommend that you have a vm dedicated to eris dev-ing.\nWould you like the marmots to create a machine for you? (y/n): ")
			fmt.Scanln(&input)

			if input == "Y" || input == "y" || input == "YES" || input == "Yes" || input == "yes" {
				logger.Printf("The marmots will create an eris machine.\n")
				if err := setupErisMachine(driver); err != nil {
					return err
				}

				logger.Infof("New docker machine created using %s driver.\nGetting the proper environment variables.\n", driver)
				if _, _, err := getMachineDeets("eris"); err != nil {
					return err
				}
			}

		}
	}

	logger.Infof("Docker client connects correctly.\n")
	return nil
}

func getMachineDeets(machName string) (string, string, error) {
	var out = new(bytes.Buffer)
	var out2 = new(bytes.Buffer)

	noConnectError := fmt.Errorf("Could not evaluate the env vars for the %s docker-machine.\n", machName)
	dHost, dPath := popHostAndPath()

	if (dHost != "" && dPath != "") && (machName == "eris" || machName == "default") {
		return dHost, dPath, nil
	}

	// TODO: when go-dockerclient adds machine API endpoints use those instead.
	logger.Debugf("Querying the %s docker-machine's url.\n", machName)
	cmd := exec.Command("docker-machine", "url", machName)
	cmd.Stdout = out
	if err := cmd.Run(); err != nil {
		return "", "", fmt.Errorf("%vError:\t%v\n", noConnectError, err)
	}
	dHost = strings.TrimSpace(out.String())
	logger.Debugf("\tURL =>\t\t\t%s\n", dHost)

	// TODO: when go-dockerclient adds machine API endpoints use those instead.
	logger.Debugf("Querying the %s docker-machine's certificate path.\n", machName)
	cmd2 := exec.Command("docker-machine", "inspect", machName, "--format", "{{.HostOptions.AuthOptions.ServerCertPath}}")
	cmd2.Stdout = out2
	//cmd2.Stderr = os.Stderr
	if err := cmd2.Run(); err != nil {
		return "", "", fmt.Errorf("%vError:\t%v\n", noConnectError, err)
	}
	dPath = out2.String()
	dPath = strings.Replace(dPath, "'", "", -1)
	dPath = path.Dir(dPath)
	logger.Debugf("\tCertificate Path =>\t%s\n", dPath)

	if dPath == "" || dHost == "" {
		return "", "", noConnectError
	}

	logger.Infof("Querying whether the host and user have access to the right files for TLS connection to docker.\n")
	if err := checkKeysAndCerts(dPath); err != nil {
		return "", "", err
	}
	logger.Debugf("\tCertificate files look good.\n")

	// technically, do not *have* to do this, but it will make repetitive tasks faster
	logger.Debugf("Setting the environment variables for quick future development.\n")
	os.Setenv("DOCKER_HOST", dHost)
	os.Setenv("DOCKER_CERT_PATH", dPath)
	os.Setenv("DOCKER_TLS_VERIFY", "1")
	os.Setenv("DOCKER_MACHINE_NAME", machName)

	logger.Debugf("Finished getting machine details =>\t%s\n", machName)
	return dHost, dPath, nil
}

func DockerClientVersion() (float64, error) {
	verR, err := DockerClient.Version()
	if err != nil {
		return 0, err
	}

	verD := verR.Get("Version")
	v := strings.Split(verD, ".")
	v = v[:len(v)-1] // we only want 1.8 so we can marshal that into a float64

	return strconv.ParseFloat(strings.Join(v, "."), 10)
}

func DockerAPIVersion() (float64, error) {
	verR, err := DockerClient.Version()
	if err != nil {
		return 0, err
	}

	verD := verR.Get("APIVersion")
	return strconv.ParseFloat(verD, 10)
}

func setupErisMachine(driver string) error {
	cmd := "docker-machine"
	args := []string{"status", "eris"}
	if err := exec.Command(cmd, args...).Run(); err == nil {
		// if err == nil this means the machine is created. if err != nil that means machine doesn't exist.
		logger.Debugf("Eris docker-machine exists. Starting.\n")
		return startErisMachine()
	}
	logger.Debugf("Eris docker-machine doesn't exist.\n")

	return createErisMachine(driver)
}

func createErisMachine(driver string) error {
	logger.Printf("Creating the eris docker-machine.\nThis will take some time, please feel free to go feed your marmot.\n")
	logger.Debugf("\tDriver =>\t\t\t%s\n", driver)
	cmd := "docker-machine"
	args := []string{"create", "--driver", driver, "eris"}
	if err := exec.Command(cmd, args...).Run(); err != nil {
		logger.Debugf("There was an error creating the eris docker-machine.\nError:\t%v\n", err)
		return mustInstallError()
	}
	logger.Debugf("Eris docker-machine created.\n")

	return startErisMachine()
}

func startErisMachine() error {
	logger.Infof("Starting eris docker-machine.\n")
	cmd := "docker-machine"
	args := []string{"start", "eris"}
	if err := exec.Command(cmd, args...).Run(); err != nil {
		return fmt.Errorf("There was an error starting the newly created docker-machine.\nError:\t%v\n", err)
	}
	logger.Debugf("Eris docker-machine started.\n")

	return nil
}

func connectDockerTLS(dockerHost, dockerCertPath string) error {
	var err error

	logger.Debugf("Connecting to the Docker Client via TLS.\n")
	logger.Debugf("\tURL =>\t\t\t%s\n", dockerHost)
	logger.Debugf("\tDocker Cert Path =>\t%s\n", dockerCertPath)

	DockerClient, err = docker.NewTLSClient(dockerHost, path.Join(dockerCertPath, "cert.pem"), path.Join(dockerCertPath, "key.pem"), path.Join(dockerCertPath, "ca.pem"))
	if err != nil {
		return err
	}

	logger.Debugf("Connected over TLS.\n")
	return nil
}

func popHostAndPath() (string, string) {
	return os.Getenv("DOCKER_HOST"), os.Getenv("DOCKER_CERT_PATH")
}

func checkKeysAndCerts(dPath string) error {
	toCheck := []string{"cert.pem", "key.pem", "ca.pem"}
	for _, f := range toCheck {
		f = path.Join(dPath, f)
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
	errBase := "The marmots cannot connect to Docker.\nDo you have docker installed?\nIf not please visit here:\t"
	dInst := "https://docs.docker.com/installation/"

	switch runtime.GOOS {
	case "linux":
		return fmt.Errorf("%s%s\nDo you have docker installed and running?\nIf not please [sudo services start docker] on Ubuntu.\nAlso check that your user is in the docker group (or rerun with sudo).\nTo fix this please run [sudo usermod -a -G docker $USER] on Ubuntu with your user substituted.", errBase, dInst)
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

func setIPFSHostViaDockerHost(dockerHost string) {
	u, err := url.Parse(dockerHost)
	if err != nil {
		IfExit(fmt.Errorf("The marmots could not parse the URL for the DockerHost to populate the IPFS Host.\nPlease check that your docker-machine VM is running with [docker-machine ls]\nError:\t%v\n", err))
	}
	dIP, _, err := net.SplitHostPort(u.Host)
	if err != nil {
		IfExit(fmt.Errorf("The marmots could not split the host and port for the DockerHost to populate the IPFS Host.\nPlease check that your docker-machine VM is running with [docker-machine ls]\nError:\t%v\n", err))

	}
	dockerIP := fmt.Sprintf("%s%s", "http://", dIP)
	logger.Debugf("Set ERIS_IPFS_HOST to =>\t%s\n", dockerIP)
	os.Setenv("ERIS_IPFS_HOST", dockerIP)
}
