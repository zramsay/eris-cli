package perform

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path"
	"regexp"
	"runtime"
	"strconv"
	"strings"

	def "github.com/eris-ltd/eris-cli/definitions"
	"github.com/eris-ltd/eris-cli/util"

	dirs "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common"
	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/fsouza/go-dockerclient"
)

// Build against Docker cli...
//   Client version: 1.6.2
//   Client API version: 1.18
// Verified against ...
//   Client version: 1.6.2
//   Client API version: 1.18
func DockerCreateDataContainer(srvName string, containerNumber int) error {
	logger.Infoln("Starting Data Container for Service: " + srvName)

	srv := def.Service{}
	ops := def.ServiceOperation{}
	srvName = util.NameAndNumber(srvName, containerNumber)
	ops.DataContainerName = fmt.Sprintf("eris_data_%s", srvName)
	optsData, err := configureDataContainer(&srv, &ops, nil)
	if err != nil {
		return err
	}

	cont, err := createContainer(optsData)
	if err != nil {
		return err
	}

	logger.Infoln("Data Container ID: " + cont.ID)
	return nil
}

// create a container with volumes-from the srvName data container
// and either attach interactively or execute a command
// container should be destroyed on exit
func DockerRunVolumesFromContainer(volumesFrom string, interactive bool, args []string) error {
	opts := configureVolumesFromContainer(volumesFrom, interactive, args)
	cont, err := createContainer(opts)
	if err != nil {
		return err
	}
	id_main := cont.ID

	// trap signals so we can drop out of the container
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)
	go func() {
		<-c
		logger.Infof("Caught signal. Stopping container %s\n", id_main)
		if err = stopContainer(id_main, 5); err != nil {
			logger.Errorf("Error stopping container: %v\n", err)
		}
	}()

	defer func() {
		logger.Infof("Removing container %s\n", id_main)
		if err2 := removeContainer(id_main); err2 != nil {
			err = fmt.Errorf("Tragic! Error removing data container after executing (%v): %v", err, err2)
		}
	}()

	logger.Infoln("Exec Container ID: " + id_main)

	// start the container (either interactive or one off command)
	if err := startContainer(id_main, &opts); err != nil {
		return err
	}

	if interactive {
		if err := attachContainer(id_main); err != nil {
			return err
		}
	} else {
		if err := logsContainer(id_main, true); err != nil {
			return err
		}
	}

	logger.Infof("Waiting for %s to exit so we can remove the container\n", id_main)
	if err := waitContainer(id_main); err != nil {
		return err
	}

	return nil
}

func DockerRun(srv *def.Service, ops *def.ServiceOperation) error {
	var id_main, id_data string
	var optsData docker.CreateContainerOptions
	var dataCont docker.APIContainers
	var dataContCreated *docker.Container

	logger.Infoln("Starting Service: " + srv.Name)

	// copy service config into docker client config
	optsServ, err := configureContainer(srv, ops)
	if err != nil {
		return err
	}

	// fix volume paths
	srv.Volumes, err = fixDirs(srv.Volumes)
	if err != nil {
		return err
	}

	// setup data container
	if ops.DataContainer {
		logger.Infoln("You've asked me to manage the data containers. I shall do so good human.")
		optsData, err = configureDataContainer(srv, ops, &optsServ)
		if err != nil {
			return err
		}
	}

	// check existence || create the container
	if servCont, exists := ContainerExists(ops); exists {
		logger.Infoln("Service Container already exists, am not creating.")
		if ops.DataContainer {
			if dataCont, exists = parseContainers(ops.DataContainerName, true); exists {
				logger.Infoln("Data Container already exists, am not creating.")
				id_data = dataCont.ID
			} else {
				logger.Infoln("Data Container does not exist, creating.")
				dataContCreated, err := createContainer(optsData)
				if err != nil {
					return err
				}
				id_data = dataContCreated.ID
			}
		}

		id_main = servCont.ID

	} else {
		logger.Infoln("Service Container does not exist, creating.")

		if ops.DataContainer {
			if dataCont, exists = parseContainers(ops.DataContainerName, true); exists {
				logger.Infoln("Data Container already exists, am not creating.")
				id_data = dataCont.ID
			} else {
				logger.Infoln("Data Container does not exist, creating.")
				dataContCreated, err = createContainer(optsData)
				if err != nil {
					return err
				}
				id_data = dataContCreated.ID
			}
		}

		servContCreated, err := createContainer(optsServ)
		if err != nil {
			return err
		}
		id_main = servContCreated.ID
	}

	// start the container
	logger.Infoln("Starting Service Container ID: " + id_main)
	if ops.DataContainer {
		logger.Infoln("with DataContainer ID: " + id_data)
	}

	if err := startContainer(id_main, &optsServ); err != nil {
		return err
	}

	// XXX: setting Remove causes us to block here!
	if ops.Remove {

		// dump the logs (TODO: options about this)
		doneLogs := make(chan struct{}, 1)
		go func() {
			logger.Debugln("following logs")
			if err := logsContainer(id_main, true); err != nil {
				logger.Errorf("Unable to follow logs for %s\n", id_main)
			}
			logger.Debugln("done following logs")
			doneLogs <- struct{}{}
		}()

		logger.Infof("Waiting for %s to exit so we can remove the container\n", id_main)
		if err := waitContainer(id_main); err != nil {
			return err
		}

		logger.Debugln("waiting for logs to finish")
		// let the logs finish
		<-doneLogs

		logger.Infof("Removing container %s\n", id_main)
		if err := removeContainer(id_main); err != nil {
			return err
		}

	} else {
		logger.Infoln(srv.Name + " Started.")
	}

	return nil
}

func DockerExec(srv *def.Service, ops *def.ServiceOperation, cmd []string, interactive bool) error {
	logger.Infoln("Starting Execution: " + srv.Name)

	// check existence || create the container
	servCont, exists := ContainerExists(ops)
	if exists {
		logger.Infoln("Service Container exists.")
	} else {
		return fmt.Errorf("Cannot exec a service which is not created. Please start the service: %s.\n", srv.Name)
	}

	if !interactive {
		// Create the execution
		logger.Infoln("Creating Service Execution:", strings.Join(cmd, " "))

		exec, err := createExec(servCont.ID, cmd, srv)
		if err != nil {
			return err
		}

		return startExec(exec.ID)
	} else {
		logger.Infoln("Attaching to Container", srv.Name)
		return attachContainer(servCont.ID)
	}
}

func DockerRebuild(srv *def.Service, ops *def.ServiceOperation, skipPull bool) error {
	var id string
	var wasRunning bool = false

	logger.Infoln("Rebuilding Service: " + srv.Name)

	if service, exists := ContainerExists(ops); exists {
		logger.Infoln("Service exists, commensing with rebuild.")

		if _, running := ContainerRunning(ops); running {
			logger.Infoln("Service is running. Stopping now.")
			wasRunning = true
			err := DockerStop(srv, ops)
			if err != nil {
				return err
			}
		}

		logger.Infoln("Removing old container.")
		err := removeContainer(service.ID)
		if err != nil {
			return err
		}

	} else {

		logger.Infoln("Service did not previously exist. Nothing to rebuild.")
		return nil
	}

	if !skipPull {
		err := DockerPull(srv, ops)
		if err != nil {
			return err
		}
	}

	opts, err := configureContainer(srv, ops)
	if err != nil {
		return err
	}
	srv.Volumes, err = fixDirs(srv.Volumes)
	if err != nil {
		return err
	}

	logger.Infoln("Recreating container.")
	cont, err := createContainer(opts)
	if err != nil {
		return err
	}
	id = cont.ID

	if wasRunning {
		logger.Infoln("Restarting Container with new ID: " + id)
		err := startContainer(id, &opts)
		if err != nil {
			return err
		}
	}

	logger.Infoln(srv.Name + " Rebuilt.")

	return nil
}

func DockerPull(srv *def.Service, ops *def.ServiceOperation) error {
	logger.Infoln("Pulling an image for the service.")

	var wasRunning bool = false

	if service, exists := ContainerExists(ops); exists {
		logger.Infoln("Found Service ID: " + service.ID)
		if _, running := ContainerRunning(ops); running {
			wasRunning = true
			err := DockerStop(srv, ops)
			if err != nil {
				return err
			}
		}
		err := removeContainer(service.ID)
		if err != nil {
			return err
		}
	}

	if logger.Level > 0 {
		err := pullImage(srv.Image, logger.Writer)
		if err != nil {
			return err
		}
	} else {
		err := pullImage(srv.Image, bytes.NewBuffer([]byte{}))
		if err != nil {
			return err
		}
	}

	if wasRunning {
		err := DockerRun(srv, ops)
		if err != nil {
			return err
		}
	}

	return nil
}

func DockerLogs(srv *def.Service, ops *def.ServiceOperation, follow bool) error {
	if service, exists := ContainerExists(ops); exists {
		logger.Infoln("Logging Service ID: " + service.ID)
		err := logsContainer(service.ID, follow)
		if err != nil {
			return err
		}
	} else {
		logger.Infoln("Service does not exist. Cannot display logs.")
	}

	return nil
}

func DockerInspect(srv *def.Service, ops *def.ServiceOperation, field string) error {
	if service, exists := ContainerExists(ops); exists {
		logger.Infoln("Inspecting Service ID: " + service.ID)
		err := inspectContainer(service.ID, field)
		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf("Service container does not exist. Cannot inspect.")
	}
	return nil
}

func DockerStop(srv *def.Service, ops *def.ServiceOperation) error {
	// don't limit this to verbose because it takes a few seconds
	logger.Println("Stopping: " + srv.Name + ". This may take a few seconds.")

	var timeout uint = 10
	dockerAPIContainer, running := ContainerExists(ops)

	if running {
		err := stopContainer(dockerAPIContainer.ID, timeout)
		if err != nil {
			return err
		}
	}

	logger.Infoln(srv.Name + " Stopped.")

	return nil
}

func DockerRename(srv *def.Service, ops *def.ServiceOperation, oldName, newName string) error {
	if service, exists := ContainerExists(ops); exists {
		logger.Infoln("Renaming Service ID: " + service.ID)
		newName = strings.Replace(service.Names[0], oldName, newName, 1)
		err := renameContainer(service.ID, newName)
		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf("Service container does not exist. Cannot rename.")
	}

	return nil
}

func DockerRemove(srv *def.Service, ops *def.ServiceOperation) error {
	if service, exists := ContainerExists(ops); exists {
		logger.Infoln("Removing Service ID: " + service.ID)
		err := removeContainer(service.ID)
		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf("Service container does not exist. Cannot remove.")
	}

	return nil
}

func ContainerExists(ops *def.ServiceOperation) (docker.APIContainers, bool) {
	return parseContainers(ops.SrvContainerName, true)
}

func ContainerRunning(ops *def.ServiceOperation) (docker.APIContainers, bool) {
	return parseContainers(ops.SrvContainerName, false)
}

// ----------------------------------------------------------------------------
// ---------------------    Images Core    ------------------------------------
// ----------------------------------------------------------------------------
func pullImage(name string, writer io.Writer) error {
	var tag string = "latest"
	var reg string = ""

	nameSplit := strings.Split(name, ":")
	if len(nameSplit) == 2 {
		tag = nameSplit[1]
	}
	if len(nameSplit) == 3 {
		tag = nameSplit[2]
	}

	repoSplit := strings.Split(nameSplit[0], "/")
	if len(repoSplit) > 2 {
		reg = repoSplit[0]
	}

	opts := docker.PullImageOptions{
		Repository:   name,
		Registry:     reg,
		Tag:          tag,
		OutputStream: writer,
	}

	auth := docker.AuthConfiguration{}

	err := util.DockerClient.PullImage(opts, auth)
	if err != nil {
		return err
	}

	return nil
}

// ----------------------------------------------------------------------------
// ---------------------    Container Core ------------------------------------
// ----------------------------------------------------------------------------
func parseContainers(name string, all bool) (docker.APIContainers, bool) {
	name = "/" + name
	containers := listContainers(all)
	if len(containers) != 0 {
		for _, container := range containers {
			if container.Names[0] == name {
				return container, true
			}
		}
	}
	return docker.APIContainers{}, false
}

func listContainers(all bool) []docker.APIContainers {
	var container []docker.APIContainers
	r := regexp.MustCompile(`\/eris_(?:service|chain|data)_(.+)_\d`)

	contns, _ := util.DockerClient.ListContainers(docker.ListContainersOptions{All: all})
	for _, con := range contns {
		for _, c := range con.Names {
			match := r.FindAllStringSubmatch(c, 1)
			if len(match) != 0 {
				container = append(container, con)
			}
		}
	}

	return container
}

func createContainer(opts docker.CreateContainerOptions) (*docker.Container, error) {
	dockerContainer, err := util.DockerClient.CreateContainer(opts)
	if err != nil {
		// TODO: better error handling
		if fmt.Sprintf("%v", err) == "no such image" {
			fmt.Printf("Pulling image from repository. This could take a second.\n")
			pullImage(opts.Config.Image, nil)
			dockerContainer, err = util.DockerClient.CreateContainer(opts)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}
	return dockerContainer, nil
}

func startContainer(id string, opts *docker.CreateContainerOptions) error {
	return util.DockerClient.StartContainer(id, opts.HostConfig)
}

func attachContainer(id string) error {
	opts := docker.AttachToContainerOptions{
		Container:    id,
		InputStream:  os.Stdin,
		OutputStream: os.Stdout,
		ErrorStream:  os.Stderr,
		Logs:         true,
		Stream:       true,
		Stdin:        true,
		Stdout:       true,
		Stderr:       true,
		RawTerminal:  true,
	}

	return util.DockerClient.AttachToContainer(opts)
}

func waitContainer(id string) error {
	exitCode, err := util.DockerClient.WaitContainer(id)
	if exitCode != 0 {
		err1 := fmt.Errorf("Container %s exited with status %d")
		if err != nil {
			err = fmt.Errorf("%s. Error: %v", err1.Error(), err)
		}
	}
	return err
}

func logsContainer(id string, tail bool) error {
	opts := docker.LogsOptions{
		Container:    id,
		Follow:       false,
		Stdout:       true,
		Stderr:       true,
		Timestamps:   false,
		OutputStream: os.Stdout,
		ErrorStream:  os.Stderr,
		RawTerminal:  true, // Usually true when the container contains a TTY.
	}

	if tail {
		opts.Tail = "all"
		opts.Follow = true
	}

	if err := util.DockerClient.Logs(opts); err != nil {
		return err
	}

	return nil
}

func inspectContainer(id, field string) error {
	cont, err := util.DockerClient.InspectContainer(id)
	if err != nil {
		return err
	}
	PrintInspectionReport(cont, field)

	return nil
}

func stopContainer(id string, timeout uint) error {
	err := util.DockerClient.StopContainer(id, timeout)
	if err != nil {
		return err
	}
	return nil
}

func renameContainer(id, newName string) error {
	opts := docker.RenameContainerOptions{
		ID:   id,
		Name: newName,
	}

	err := util.DockerClient.RenameContainer(opts)
	if err != nil {
		return err
	}

	return nil
}

func removeContainer(id string) error {
	opts := docker.RemoveContainerOptions{
		ID:            id,
		RemoveVolumes: false,
		Force:         false,
	}

	err := util.DockerClient.RemoveContainer(opts)
	if err != nil {
		return err
	}

	return nil
}

func configureContainer(srv *def.Service, ops *def.ServiceOperation) (docker.CreateContainerOptions, error) {
	if ops.ContainerNumber == 0 {
		ops.ContainerNumber = 1
	}

	opts := docker.CreateContainerOptions{
		Name: ops.SrvContainerName,
		Config: &docker.Config{
			Hostname:        srv.HostName,
			Domainname:      srv.DomainName,
			User:            srv.User,
			Memory:          srv.MemLimit,
			CPUShares:       srv.CPUShares,
			AttachStdin:     false,
			AttachStdout:    false,
			AttachStderr:    false,
			Tty:             false,
			OpenStdin:       false,
			Env:             srv.Environment,
			Labels:          srv.Labels,
			Cmd:             strings.Fields(srv.Command),
			Entrypoint:      strings.Fields(srv.EntryPoint),
			Image:           srv.Image,
			WorkingDir:      srv.WorkDir,
			NetworkDisabled: false,
		},
		HostConfig: &docker.HostConfig{
			Binds:           srv.Volumes,
			Links:           srv.Links,
			PublishAllPorts: false,
			Privileged:      ops.Privileged,
			ReadonlyRootfs:  false,
			DNS:             srv.DNS,
			DNSSearch:       srv.DNSSearch,
			VolumesFrom:     srv.VolumesFrom,
			CapAdd:          srv.CapAdd,
			CapDrop:         srv.CapDrop,
			RestartPolicy:   docker.NeverRestart(),
			NetworkMode:     "bridge",
		},
	}

	if ops.Attach {
		opts.Config.AttachStdin = true
		opts.Config.AttachStdout = true
		opts.Config.AttachStderr = true
		opts.Config.Tty = true
		opts.Config.OpenStdin = true
	}

	if ops.Restart == "always" {
		opts.HostConfig.RestartPolicy = docker.AlwaysRestart()
	} else if strings.Contains(ops.Restart, "max") {
		times, err := strconv.Atoi(strings.Split(ops.Restart, ":")[1])
		if err != nil {
			return docker.CreateContainerOptions{}, err
		}
		opts.HostConfig.RestartPolicy = docker.RestartOnFailure(times)
	}

	opts.Config.ExposedPorts = make(map[docker.Port]struct{})
	opts.HostConfig.PortBindings = make(map[docker.Port][]docker.PortBinding)
	opts.Config.Volumes = make(map[string]struct{})

	for _, port := range srv.Ports {
		pS := strings.Split(port, ":")

		pR := pS[len(pS)-1]
		if len(strings.Split(pR, "/")) == 1 {
			pR = pR + "/tcp"
		}
		pC := docker.Port(fmt.Sprintf("%s", pR))

		if len(pS) > 1 {
			pH := docker.PortBinding{
				HostPort: pS[len(pS)-2],
			}

			if len(pS) == 3 {
				// ipv4
				pH.HostIP = pS[0]
			} else if len(pS) > 3 {
				// ipv6
				pH.HostIP = strings.Join(pS[:len(pS)-2], ":")
			}

			opts.Config.ExposedPorts[pC] = struct{}{}
			opts.HostConfig.PortBindings[pC] = []docker.PortBinding{pH}
		} else {
			opts.Config.ExposedPorts[pC] = struct{}{}
		}
	}

	for _, vol := range srv.Volumes {
		opts.Config.Volumes[strings.Split(vol, ":")[1]] = struct{}{}
	}

	for _, depServ := range srv.ServiceDeps {
		name := depServ
		depServ = nameToContainerName("service", depServ, ops.ContainerNumber)
		newLink := depServ + ":" + name
		opts.HostConfig.Links = append(opts.HostConfig.Links, newLink)
	}

	return opts, nil
}

func configureVolumesFromContainer(volumesFrom string, interactive bool, args []string) docker.CreateContainerOptions {
	opts := docker.CreateContainerOptions{
		Name: "eris_exec_" + volumesFrom,
		Config: &docker.Config{
			Image:           "eris/base",
			User:            "root",
			AttachStdout:    true,
			AttachStderr:    true,
			Tty:             true,
			StdinOnce:       true,
			NetworkDisabled: true,
		},
		HostConfig: &docker.HostConfig{
			VolumesFrom: []string{volumesFrom},
		},
	}
	if interactive {
		opts.Config.AttachStdin = true
		opts.Config.OpenStdin = true
		opts.Config.Cmd = []string{"/bin/bash"}
	} else {
		opts.Config.Cmd = args
	}

	return opts
}

func configureDataContainer(srv *def.Service, ops *def.ServiceOperation, mainContOpts *docker.CreateContainerOptions) (docker.CreateContainerOptions, error) {
	opts := docker.CreateContainerOptions{
		Name: ops.DataContainerName,
		Config: &docker.Config{
			Image:           "eris/data",
			User:            srv.User,
			AttachStdin:     false,
			AttachStdout:    false,
			AttachStderr:    false,
			Tty:             false,
			OpenStdin:       false,
			NetworkDisabled: true,
		},
		HostConfig: &docker.HostConfig{},
	}

	if mainContOpts != nil {
		mainContOpts.HostConfig.VolumesFrom = append(mainContOpts.HostConfig.VolumesFrom, ops.DataContainerName)
	}

	return opts, nil
}

// ----------------------------------------------------------------------------
// ---------------------    Exec Core -----------------------------------------
// ----------------------------------------------------------------------------
func createExec(container string, cmd []string, srv *def.Service) (*docker.Exec, error) {
	opts := docker.CreateExecOptions{
		AttachStdin:  false,
		AttachStdout: true,
		AttachStderr: true,
		Tty:          true,
		Cmd:          cmd,
		Container:    container,
		User:         srv.User,
	}

	if srv.User == "" {
		opts.User = "eris"
	}

	return util.DockerClient.CreateExec(opts)
}

func startExec(id string) error {
	opts := docker.StartExecOptions{
		Detach:       false,
		Tty:          true,
		InputStream:  os.Stdin,
		OutputStream: os.Stdout,
		ErrorStream:  os.Stderr,
		RawTerminal:  true,
	}

	return util.DockerClient.StartExec(id, opts)
}

// ----------------------------------------------------------------------------
// ---------------------    Util Funcs ----------------------------------------
// ----------------------------------------------------------------------------

// $(pwd) doesn't execute properly in golangs subshells; replace it
// use $eris as a shortcut
func fixDirs(arg []string) ([]string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return []string{}, err
	}

	for n, a := range arg {
		if strings.Contains(a, "$eris") {
			tmp := strings.Split(a, ":")[0]
			keep := strings.Replace(a, tmp+":", "", 1)
			if runtime.GOOS == "windows" {
				winTmp := strings.Split(tmp, "/")
				tmp = path.Join(winTmp...)
			}
			tmp = strings.Replace(tmp, "$eris", dirs.ErisRoot, 1)
			arg[n] = strings.Join([]string{tmp, keep}, ":")
			continue
		}

		if strings.Contains(a, "$pwd") {
			arg[n] = strings.Replace(a, "$pwd", dir, 1)
		}
	}

	return arg, nil
}

func nameToContainerName(contType, name string, num int) string {
	return "eris_" + contType + "_" + name + "_" + fmt.Sprintf("%v", num)
}
