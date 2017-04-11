package compilers

import (
	"bytes"
	"fmt"
	"os"

	"github.com/monax/cli/log"
	"github.com/monax/cli/util"

	docker "github.com/fsouza/go-dockerclient"
)

func Install(language, version string) error {
	return util.DockerClient.PullImage(
		docker.PullImageOptions{
			Repository:   language,
			Tag:          version,
			OutputStream: os.Stdout,
		},
		docker.AuthConfiguration{},
	)
}

func List(image string) ([]string, error) {
	return []string{}, nil
}

//if we want to we can move this into perform, just that for now I would like to have
//more fine grained control over the process and perform is unfortunately something of
//a clusterfuck to wade through atm.
func executeCompilerCommand(image string, command []string) ([]byte, error) {
	pwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	//create container with volumes premounted
	opts := docker.CreateContainerOptions{
		Name: util.UniqueName("compiler"),
		Config: &docker.Config{
			Image:           image,
			User:            "root",
			AttachStdout:    true,
			AttachStderr:    true,
			AttachStdin:     true,
			Tty:             true,
			NetworkDisabled: false,
			WorkingDir:      "/home/",
			Cmd:             command,
		},
		HostConfig: &docker.HostConfig{
			Binds: []string{pwd + ":" + "/home/"},
		},
	}
	container, err := util.DockerClient.CreateContainer(opts)
	if err != nil {
		return nil, util.DockerError(err)
	}
	removeOpts := docker.RemoveContainerOptions{
		ID:            container.ID,
		RemoveVolumes: true,
		Force:         true,
	}
	defer util.DockerClient.RemoveContainer(removeOpts)
	// Start the container.
	log.WithField("=>", opts.Name).Debug("Starting data container")
	if err = util.DockerError(util.DockerClient.StartContainer(opts.Name, opts.HostConfig)); err != nil {
		return nil, err
	}

	log.WithField("=>", opts.Name).Debug("Waiting for data container to exit")
	exitCode, err := util.DockerClient.WaitContainer(container.ID)
	if err != nil {
		return nil, err
	}

	var stdout bytes.Buffer

	logOpts := docker.LogsOptions{
		Container:    container.ID,
		OutputStream: &stdout,
		RawTerminal:  true,
		Follow:       true,
		Stdout:       true,
		Stderr:       true,
		Since:        0,
		Timestamps:   false,
		Tail:         "all",
	}
	log.WithField("=>", opts.Name).Debug("Getting logs from container")
	if err = util.DockerClient.Logs(logOpts); err != nil {
		log.Warn("Can't get logs")
		return nil, util.DockerError(err)
	}

	// Return the logs as a byte slice, if possible.
	if exitCode == 0 {
		return stdout.Bytes(), nil
	} else if exitCode == 1 {
		return stdout.Bytes(), fmt.Errorf("Compiler error.")
	} else {
		return nil, nil
	}
}
