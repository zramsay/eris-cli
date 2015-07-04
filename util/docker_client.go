package util

import (
	"fmt"
	"os"
	"path"
	"regexp"
	"runtime"

	def "github.com/eris-ltd/eris-cli/definitions"

	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/fsouza/go-dockerclient"
)

// Docker Client initialization
var DockerClient *docker.Client

func DockerConnect(verbose bool) {
	var err error

	if runtime.GOOS == "linux" {
		endpoint := "unix:///var/run/docker.sock"

		logger.Debugln("Connecting to the Docker Client via:", endpoint)

		DockerClient, err = docker.NewClient(endpoint)
		if err != nil {
			// TODO: better error handling
			fmt.Println(err)
			os.Exit(1)
		}

		logger.Debugln("Successfully connected to Docker daemon.")

	} else {

		dockerCertPath := os.Getenv("DOCKER_CERT_PATH")

		logger.Debugln("Connecting to the Docker Client via:", os.Getenv("DOCKER_HOST"))
		logger.Debugln("Docker Certificate Path:", dockerCertPath)

		DockerClient, err = docker.NewTLSClient(os.Getenv("DOCKER_HOST"), path.Join(dockerCertPath, "cert.pem"), path.Join(dockerCertPath, "key.pem"), path.Join(dockerCertPath, "ca.pem"))

		if err != nil {
			// TODO: better error handling
			fmt.Println(err)
			os.Exit(1)
		}

		logger.Debugln("Successfully connected to Docker daemon")
	}
}

func NameAndNumber(name string, number int) string {
	return fmt.Sprintf("%s_%d", name, number)
}

func erisRegExp(typ string) *regexp.Regexp {
	return regexp.MustCompile(fmt.Sprintf(`\/eris_%s_(.+)_(\d+)`, typ))
}

// docker has this weird thing where it returns links as individual
// container (as in there is the container of two linked services and
// the linkage between them is actually its own containers). this explains
// the leading hash on containers. the q regexp is to filer out these
// links from the return list as they are irrelevant to the information
// desired by this function. and frankly they give false answers to
// IsServiceRunning and ls,ps,known functions.
func erisRegExpLinks() *regexp.Regexp {
	return regexp.MustCompile(`\A\/eris_service_(.+?)_\d/(.+?)\z`)
}

func ParseContainerNames(typ string, running bool) []string {
	containers := []string{}
	r := erisRegExp(typ)   // eris containers
	q := erisRegExpLinks() // skip past these

	contns, err := DockerClient.ListContainers(docker.ListContainersOptions{All: running})
	if len(contns) == 0 || err != nil {
		logger.Infoln("There are no containers.")
		return nil
	}
	for _, con := range contns {
		for _, c := range con.Names {
			match := r.FindAllStringSubmatch(c, 1)
			if typ == "service" {
				m2 := q.FindAllStringSubmatch(c, 1)
				if len(m2) != 0 {
					continue
				}
			}
			if len(match) != 0 {
				m := r.FindAllStringSubmatch(c, 1)[0]
				cNameNum := fmt.Sprintf("%s_%s", m[1], m[2])
				containers = append(containers, cNameNum)
			}
		}
	}

	return containers

}

// need to be alot smarter with this
func OverwriteOps(opsBase, opsOver *def.ServiceOperation) {
	if opsOver.PublishAllPorts {
		opsBase.PublishAllPorts = true
	}
}
