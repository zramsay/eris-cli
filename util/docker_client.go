package util

import (
	"fmt"
	"os"
	"path"
	"runtime"

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
