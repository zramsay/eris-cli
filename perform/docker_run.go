package perform

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path"
	"regexp"
	"runtime"
	"strconv"
	"strings"

	"github.com/eris-ltd/eris-cli/util"

	dirs "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common"
	def "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common/definitions"
	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/fsouza/go-dockerclient"
)

// Build against Docker cli...
//   Client version: 1.6.2
//   Client API version: 1.18
// Verified against ...
//   Client version: 1.6.2
//   Client API version: 1.18
func DockerCreateDataContainer(srvName string, verbose bool, w io.Writer) error {
	if verbose {
		w.Write([]byte("Starting Service: " + srvName))
	}

	srv := def.Service{}
	ops := def.ServiceOperation{}
	containerNumber := 1
	ops.DataContainerName = "eris_data_" + srvName + "_" + strconv.Itoa(containerNumber)
	optsData, err := configureDataContainer(&srv, &ops, &docker.CreateContainerOptions{})
	if err != nil {
		return err
	}

	cont, err := createContainer(optsData)
	if err != nil {
		return err
	}

	if verbose {
		w.Write([]byte("Data Container ID: " + cont.ID))
	}
	return nil
}

func DockerRun(srv *def.Service, ops *def.ServiceOperation, verbose bool, w io.Writer) error {
	var id_main string
	var id_data string
	var optsData docker.CreateContainerOptions
	var dataCont docker.APIContainers
	var dataContCreated *docker.Container

	if verbose {
		w.Write([]byte("Starting Service: " + srv.Name))
	}

	// configure
	optsServ, err := configureContainer(srv, ops)
	if err != nil {
		return err
	}

	srv.Volumes, err = fixDirs(srv.Volumes)
	if err != nil {
		return err
	}

	if ops.DataContainer {
		if verbose {
			w.Write([]byte("You've asked me to manage the data containers. I shall do so good human."))
		}
		optsData, err = configureDataContainer(srv, ops, &optsServ)
		if err != nil {
			return err
		}
	}

	// check existence || create the container
	if servCont, exists := ContainerExists(ops); exists {
		if verbose {
			w.Write([]byte("Service Container already exists, am not creating."))
		}
		if ops.DataContainer {
			if dataCont, exists = parseContainers(ops.DataContainerName, true); exists {
				if verbose {
					w.Write([]byte("Data Container already exists, am not creating."))
				}
				id_data = dataCont.ID
			} else {
				if verbose {
					w.Write([]byte("Data Container does not exist, creating."))
				}
				dataContCreated, err := createContainer(optsData)
				if err != nil {
					return err
				}
				id_data = dataContCreated.ID
			}
		}

		id_main = servCont.ID

	} else {
		if verbose {
			w.Write([]byte("Service Container does not exist, creating."))
		}

		if ops.DataContainer {
			if dataCont, exists = parseContainers(ops.DataContainerName, true); exists {
				if verbose {
					w.Write([]byte("Data Container already exists, am not creating."))
				}
				id_data = dataCont.ID
			} else {
				if verbose {
					w.Write([]byte("Data Container does not exist, creating."))
				}
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
	if verbose {
		w.Write([]byte("Starting ServiceContainer ID: " + id_main))
		if ops.DataContainer {
			w.Write([]byte("with DataContainer ID: " + id_data))
		}
	}

	err = startContainer(id_main, &optsServ)
	if err != nil {
		return err
	}

	if verbose {
		w.Write([]byte(srv.Name + " Started."))
	}
	return nil
}

func DockerRebuild(srv *def.Service, ops *def.ServiceOperation, pull, verbose bool, w io.Writer) error {
	var id string
	var wasRunning bool = false

	if verbose {
		w.Write([]byte("Rebuilding Service: " + srv.Name))
	}

	if service, exists := ContainerExists(ops); exists {
		if verbose {
			w.Write([]byte("Service exists, commensing with rebuild."))
		}

		if _, running := ContainerRunning(ops); running {
			if verbose {
				w.Write([]byte("Service is running. Stopping now."))
			}
			wasRunning = true
			err := DockerStop(srv, ops, verbose, w)
			if err != nil {
				return err
			}
		}

		if verbose {
			w.Write([]byte("Removing old container."))
		}
		err := removeContainer(service.ID)
		if err != nil {
			return err
		}

	} else {

		if verbose {
			w.Write([]byte("Service did not previously exist. Nothing to rebuild."))
		}
		return nil
	}

	if pull {
		err := DockerPull(srv, ops, verbose, w)
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

	if verbose {
		w.Write([]byte("Recreating container."))
	}
	cont, err := createContainer(opts)
	if err != nil {
		return err
	}
	id = cont.ID

	if wasRunning {
		if verbose {
			w.Write([]byte("Restarting Container with new ID: " + id))
		}
		err := startContainer(id, &opts)
		if err != nil {
			return err
		}
	}

	if verbose {
		w.Write([]byte(srv.Name + " Rebuilt."))
	}

	return nil
}

func DockerPull(srv *def.Service, ops *def.ServiceOperation, verbose bool, w io.Writer) error {
	if verbose {
		w.Write([]byte("Pulling an image for the service."))
	}

	var wasRunning bool = false

	if service, exists := ContainerExists(ops); exists {
		if verbose {
			w.Write([]byte("Service ID: " + service.ID))
		}
		if _, running := ContainerRunning(ops); running {
			wasRunning = true
			err := DockerStop(srv, ops, verbose, w)
			if err != nil {
				return err
			}
		}
		err := removeContainer(service.ID)
		if err != nil {
			return err
		}
	}

	if verbose {
		err := pullImage(srv.Image, w)
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
		err := DockerRun(srv, ops, verbose, w)
		if err != nil {
			return err
		}
	}

	return nil
}

func DockerLogs(srv *def.Service, ops *def.ServiceOperation, verbose bool, w io.Writer) error {
	if verbose {
		w.Write([]byte("Getting service's logs."))
	}

	if service, exists := ContainerExists(ops); exists {
		if verbose {
			w.Write([]byte("Service ID: " + service.ID))
		}
		err := logsContainer(service.ID)
		if err != nil {
			return err
		}
	} else {
		if verbose {
			w.Write([]byte("Service does not exist. Cannot display logs."))
		}
	}

	return nil
}

func DockerInspect(srv *def.Service, ops *def.ServiceOperation, field string, verbose bool, w io.Writer) error {
	if verbose {
		w.Write([]byte("Inspecting service"))
	}

	if service, exists := ContainerExists(ops); exists {
		if verbose {
			w.Write([]byte("Service ID: " + service.ID))
		}
		err := inspectContainer(service.ID, field)
		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf("Service container does not exist. Cannot inspect.")
	}
	return nil
}

func DockerStop(srv *def.Service, ops *def.ServiceOperation, verbose bool, w io.Writer) error {
	// don't limit this to verbose because it takes a few seconds
	w.Write([]byte("Stopping: " + srv.Name + ". This may take a few seconds."))

	var timeout uint = 10
	dockerAPIContainer, running := ContainerExists(ops)

	if running {
		err := stopContainer(dockerAPIContainer.ID, timeout)
		if err != nil {
			return err
		}
	}

	if verbose {
		w.Write([]byte(srv.Name + " Stopped."))
	}

	return nil
}

func DockerRename(srv *def.Service, ops *def.ServiceOperation, oldName, newName string, verbose bool, w io.Writer) error {
	if service, exists := ContainerExists(ops); exists {
		if verbose {
			w.Write([]byte("Service ID: " + service.ID))
		}
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

func DockerRemove(srv *def.Service, ops *def.ServiceOperation, verbose bool, w io.Writer) error {
	if service, exists := ContainerExists(ops); exists {
		if verbose {
			w.Write([]byte("Service ID: " + service.ID))
		}
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
	err := util.DockerClient.StartContainer(id, opts.HostConfig)
	if err != nil {
		return err
	}
	return nil
}

func logsContainer(id string) error {
	opts := docker.LogsOptions{
		Container:    id,
		Follow:       false,
		Stdout:       true,
		Stderr:       true,
		Timestamps:   false,
		OutputStream: os.Stdout,
		ErrorStream:  os.Stderr,
		// Tail:         "",
		// RawTerminal:  true, // Usually true when the container contains a TTY.
	}

	err := util.DockerClient.Logs(opts)
	if err != nil {
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

	return opts, nil
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
