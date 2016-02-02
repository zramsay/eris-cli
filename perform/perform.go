package perform

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/eris-ltd/eris-cli/config"
	def "github.com/eris-ltd/eris-cli/definitions"
	"github.com/eris-ltd/eris-cli/util"
	ver "github.com/eris-ltd/eris-cli/version"

	log "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/Sirupsen/logrus"
	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/docker/docker/pkg/term"
	dirs "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common/go/common"
	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/fsouza/go-dockerclient"
)

var (
	ErrContainerExists = errors.New("container exists")
)

// DockerCreateData creates a blank data container. It returns ErrContainerExists
// if such a container exists or other Docker errors.
//
//  ops.DataContainerName  - data container name to be created
//  ops.ContainerType      - container type
//  ops.ContainerNumber    - container number
//  ops.Labels             - container creation time labels (use LoadDataDefinition)
//
func DockerCreateData(ops *def.Operation) error {
	log.WithField("=>", ops.DataContainerName).Info("Creating data container")

	if _, exists := ContainerExists(ops); exists {
		log.Info("Data container exists. Not creating")
		return ErrContainerExists
	}

	optsData, err := configureDataContainer(def.BlankService(), ops, nil)
	if err != nil {
		return err
	}

	_, err = createContainer(optsData)
	if err != nil {
		return err
	}

	log.WithField("=>", optsData.Name).Info("Data container created")

	return nil
}

// DockerRunData runs a data container with the volumes-from option set.
// The container is destroyed on exit; the container log is returned as a byte stream.
// If service parameter is specified, the command inherits the service settings
// from that container.
//
//  ops.DataContainerName - container name to be mount with `--volumes-from=[]` option.
//  ops.ContainerType     - container type
//  ops.ContainerNumber   - container number
//  ops.Labels            - container creation time labels (use LoadDataDefinition)
//  ops.Args              - if specified, run these args in a container
//
func DockerRunData(ops *def.Operation, service *def.Service) (result []byte, err error) {
	log.WithFields(log.Fields{
		"=>":   ops.DataContainerName,
		"args": ops.Args,
	}).Info("Running data container")

	opts := configureVolumesFromContainer(ops, service)
	log.WithField("image", opts.Config.Image).Info("Data container configured")

	_, err = createContainer(opts)
	if err != nil {
		return nil, err
	}

	// Clean up the container.
	defer func() {
		log.WithField("=>", opts.Name).Info("Removing data container")
		if err2 := removeContainer(opts.Name, true, false); err2 != nil {
			if os.Getenv("CIRCLE_BRANCH") == "" {
				err = fmt.Errorf("Tragic! Error removing data container after executing (%v): %v", err, err2)
			}
		}
		log.WithField("=>", opts.Name).Info("Container removed")
	}()

	// Start the container.
	log.WithField("=>", opts.Name).Info("Starting data container")
	if err = startContainer(opts); err != nil {
		return nil, err
	}

	log.WithField("=>", opts.Name).Info("Waiting for data container to exit")
	if err := waitContainer(opts.Name); err != nil {
		return nil, err
	}

	log.WithField("=>", opts.Name).Info("Getting logs from container")
	if err = logsContainer(opts.Name, true, "all"); err != nil {
		return nil, err
	}

	// Return the logs as a byte slice, if possible.
	reader, ok := config.GlobalConfig.Writer.(*bytes.Buffer)
	if ok {
		return reader.Bytes(), nil
	}
	return nil, nil
}

// DockerExecData runs a data container with volumes-from field set interactively.
// It returns the container output or error on exit.
//
//  ops.Args         - command line parameters
//  ops.Interactive  - if true, set Entrypoint to ops.Args,
//                     if false, set Cmd to ops.Args
//
// See parameter description for DockerRunData.
func DockerExecData(ops *def.Operation, service *def.Service) (buf *bytes.Buffer, err error) {
	log.WithFields(log.Fields{
		"=>":   ops.DataContainerName,
		"args": ops.Args,
	}).Info("Executing data container")

	opts := configureVolumesFromContainer(ops, service)
	log.WithField("image", opts.Config.Image).Info("Data container configured")

	_, err = createContainer(opts)
	if err != nil {
		return nil, err
	}

	// Clean up the container.
	defer func() {
		log.WithField("=>", opts.Name).Info("Removing data container")
		if err2 := removeContainer(opts.Name, true, false); err2 != nil {
			if os.Getenv("CIRCLE_BRANCH") == "" {
				err = fmt.Errorf("Tragic! Error removing data container after executing (%v): %v", err, err2)
			}
		}
		log.WithField("=>", opts.Name).Info("Data container removed")
	}()

	// Save writer values for later and restore on exit.
	stdout, stderr := config.GlobalConfig.Writer, config.GlobalConfig.ErrorWriter
	defer func() {
		config.GlobalConfig.Writer, config.GlobalConfig.ErrorWriter = stdout, stderr
	}()

	buf = new(bytes.Buffer)
	config.GlobalConfig.Writer = buf
	config.GlobalConfig.ErrorWriter = buf

	// Start the container.
	log.WithField("=>", opts.Name).Info("Executing interactive data container")
	if err = startInteractiveContainer(opts); err != nil {
		return nil, err
	}

	return buf, err
}

// DockerRunService creates and runs a chain or a service container with the srv
// settings template. It also creates dependent data containers if srv.AutoData
// is true. DockerRunService returns Docker errors if not successful.
//
//  srv.AutoData          - if true, create or use existing data container
//  srv.Restart           - container restart policy ("always", "max:<#attempts>"
//                          or never if unspecified)
//
//  ops.SrvContainerName  - service or a chain container name
//  ops.DataContainerName - dependent data container name
//  ops.ContainerNumber   - container number
//  ops.ContainerType     - container type
//  ops.Labels            - container creation time labels
//                          (use LoadServiceDefinition or LoadChainDefinition)
// Container parameters:
//
//  ops.Remove            - remove container on exit (similar to `docker run --rm`)
//  ops.PublishAllPorts   - if true, publish exposed ports to random ports
//  ops.CapAdd            - add linux capabilities (similar to `docker run --cap-add=[]`)
//  ops.CapDrop           - add linux capabilities (similar to `docker run --cap-drop=[]`)
//  ops.Privileged        - if true, give extended privileges
//
func DockerRunService(srv *def.Service, ops *def.Operation) error {
	log.WithField("=>", ops.SrvContainerName).Info("Running container")

	_, running := ContainerRunning(ops)
	if running {
		log.WithField("=>", ops.SrvContainerName).Info("Container already started. Skipping")
		return nil
	}

	optsServ := configureServiceContainer(srv, ops)

	// Fix volume paths.
	var err error
	srv.Volumes, err = util.FixDirs(srv.Volumes)
	if err != nil {
		return err
	}

	// Setup data container.
	log.WithField("autodata", srv.AutoData).Info("Manage data containers?")
	if srv.AutoData {
		optsData, err := configureDataContainer(srv, ops, &optsServ)
		if err != nil {
			return err
		}

		if _, exists := util.ParseContainers(ops.DataContainerName, true); exists {
			log.Info("Data container already exists. Not creating")
		} else {
			log.Info("Data container does not exist. Creating")
			_, err := createContainer(optsData)
			if err != nil {
				return err
			}
		}
	}

	// Check existence || create the container.
	if _, exists := ContainerExists(ops); exists {
		log.Debug("Container already exists. Not creating")
	} else {
		log.WithField("image", srv.Image).Debug("Container does not exist. Creating")

		_, err := createContainer(optsServ)
		if err != nil {
			return err
		}
	}

	// Start the container.
	log.WithFields(log.Fields{
		"=>":              optsServ.Name,
		"data container":  ops.DataContainerName,
		"entrypoint":      optsServ.Config.Entrypoint,
		"cmd":             optsServ.Config.Cmd,
		"published ports": optsServ.HostConfig.PublishAllPorts,
		"environment":     optsServ.Config.Env,
		"image":           optsServ.Config.Image,
	}).Info("Starting container")
	if err := startContainer(optsServ); err != nil {
		return err
	}

	if ops.Remove {
		log.WithField("=>", optsServ.Name).Info("Removing container")
		if err := removeContainer(optsServ.Name, false, false); err != nil {
			return err
		}
	}

	log.WithField("=>", optsServ.Name).Info("Container started")

	return nil
}

// DockerExecService creates and runs a chain or a service container interactively.
//
//  ops.Args         - command line parameters
//  ops.Interactive  - if true, set Entrypoint to ops.Args,
//                     if false, set Cmd to ops.Args
//
// See parameter description for DockerRunService.
func DockerExecService(srv *def.Service, ops *def.Operation) (buf *bytes.Buffer, err error) {
	log.WithField("=>", ops.SrvContainerName).Info("Executing container")

	optsServ := configureInteractiveContainer(srv, ops)

	// Fix volume paths.
	srv.Volumes, err = util.FixDirs(srv.Volumes)
	if err != nil {
		return nil, err
	}

	// Setup data container.
	log.WithField("autodata", srv.AutoData).Info("Manage data containers?")

	if srv.AutoData {
		optsData, err := configureDataContainer(srv, ops, &optsServ)
		if err != nil {
			return nil, err
		}

		if _, exists := util.ParseContainers(ops.DataContainerName, true); exists {
			log.Info("Data container already exists, am not creating")
		} else {
			log.Info("Data container does not exist. Creating")

			_, err := createContainer(optsData)
			if err != nil {
				return nil, err
			}
		}
	}

	log.WithField("image", srv.Image).Debug("Container does not exist. Creating")
	_, err = createContainer(optsServ)
	if err != nil {
		return nil, err
	}

	defer func() {
		log.WithField("=>", optsServ.Name).Info("Removing container")
		if err := removeContainer(optsServ.Name, false, false); err != nil {
			log.WithField("=>", optsServ.Name).Error("Tragic! Error removing data container after executing")
			log.Error(err)
		}
		log.WithField("=>", optsServ.Name).Info("Container removed")
	}()

	// Save writer values for later and restore on exit.
	stdout, stderr := config.GlobalConfig.Writer, config.GlobalConfig.ErrorWriter
	defer func() {
		config.GlobalConfig.Writer, config.GlobalConfig.ErrorWriter = stdout, stderr
	}()

	buf = new(bytes.Buffer)
	config.GlobalConfig.Writer = buf
	config.GlobalConfig.ErrorWriter = buf

	// Start the container.
	log.WithFields(log.Fields{
		"=>":              optsServ.Name,
		"data container":  ops.DataContainerName,
		"entrypoint":      optsServ.Config.Entrypoint,
		"cmd":             optsServ.Config.Cmd,
		"ports published": optsServ.HostConfig.PublishAllPorts,
		"environment":     optsServ.Config.Env,
		"image":           optsServ.Config.Image,
		"user":            optsServ.Config.User,
		"vols":            optsServ.HostConfig.Binds,
	}).Info("Executing interactive container")
	if err := startInteractiveContainer(optsServ); err != nil {
		return nil, err
	}

	return buf, nil
}

// DockerRebuild recreates the container based on the srv settings template.
// If pullImage is true, it updates the Docker image before recreating
// the container. Timeout is a number of seconds to wait before killing the
// container process ungracefully.
//
//  ops.SrvContainerName  - service or a chain container name to rebuild
//  ops.ContainerNumber   - container number
//  ops.ContainerType     - container type
//  ops.Labels            - container creation time labels
//
// Also see container parameters for DockerRunService.
func DockerRebuild(srv *def.Service, ops *def.Operation, pullImage bool, timeout uint) error {
	var wasRunning bool = false

	log.WithField("=>", srv.Name).Info("Rebuilding container")

	if _, exists := ContainerExists(ops); exists {
		if _, running := ContainerRunning(ops); running {
			wasRunning = true
			err := DockerStop(srv, ops, timeout)
			if err != nil {
				return err
			}
		}

		log.WithField("=>", ops.SrvContainerName).Info("Removing old container")
		err := removeContainer(ops.SrvContainerName, true, false)
		if err != nil {
			return err
		}

	} else {
		log.Info("Container did not previously exist. Nothing to rebuild")
		return nil
	}

	if pullImage {
		log.WithField("image", srv.Image).Info("Pulling image")
		err := DockerPull(srv, ops)
		if err != nil {
			return err
		}
	}

	opts := configureServiceContainer(srv, ops)
	var err error
	srv.Volumes, err = util.FixDirs(srv.Volumes)
	if err != nil {
		return err
	}

	log.WithField("=>", ops.SrvContainerName).Info("Recreating container")
	_, err = createContainer(opts)
	if err != nil {
		return err
	}

	if wasRunning {
		log.WithField("=>", opts.Name).Info("Restarting container")
		err := startContainer(opts)
		if err != nil {
			return err
		}
	}

	log.WithField("=>", ops.SrvContainerName).Info("Container rebuilt")

	return nil
}

// DockerPull pulls the image for the container specified in srv.Image.
// DockerPull returns Docker errors on exit if not successful.
//
//  ops.SrvContainerName  - service or a chain container name
//  ops.ContainerNumber   - container number
//  ops.ContainerType     - container type
//  ops.Labels            - container creation time labels
//
// Also see container parameters for DockerRunService.
func DockerPull(srv *def.Service, ops *def.Operation) error {
	log.WithFields(log.Fields{
		"=>":    srv.Name,
		"image": srv.Image,
	}).Info("Pulling container image for")

	var wasRunning bool = false

	if _, exists := ContainerExists(ops); exists {
		if _, running := ContainerRunning(ops); running {
			wasRunning = true
			if err := DockerStop(srv, ops, 10); err != nil {
				return err
			}
		}
		if err := removeContainer(ops.SrvContainerName, false, false); err != nil {
			return err
		}
	}

	if log.GetLevel() > 0 {
		if err := pullImage(srv.Image, os.Stdout); err != nil {
			return err
		}
	} else {
		if err := pullImage(srv.Image, bytes.NewBuffer([]byte{})); err != nil {
			return err
		}
	}

	if wasRunning {
		if err := DockerRunService(srv, ops); err != nil {
			return err
		}
	}

	return nil
}

// DockerLogs displays tail number of lines of container ops.SrvContainerName
// output. If follow is true, it behaves like `tail -f`. It returns Docker
// errors on exit if not successful.
func DockerLogs(srv *def.Service, ops *def.Operation, follow bool, tail string) error {
	if _, exists := ContainerExists(ops); exists {
		log.WithFields(log.Fields{
			"=>":     ops.SrvContainerName,
			"follow": follow,
			"tail":   tail,
		}).Info("Getting logs")
		if err := logsContainer(ops.SrvContainerName, follow, tail); err != nil {
			return err
		}
	} else {
		log.Info("Container does not exist. Cannot display logs")
	}

	return nil
}

// DockerInspect displays container ops.SrvContainerName data on the terminal.
// field can be a field name of one of `docker inspect` output or it can be
// either "line" to display a short info line or "all" to display everything. I
// DockerInspect returns Docker errors on exit in not successful.
func DockerInspect(srv *def.Service, ops *def.Operation, field string) error {
	if _, exists := ContainerExists(ops); exists {
		log.WithField("=>", ops.SrvContainerName).Info("Inspecting")
		err := inspectContainer(ops.SrvContainerName, field)
		if err != nil {
			return err
		}
	} else {
		log.Info("Container does not exist. Cannot inspect")
	}
	return nil
}

// DockerStop stops a running ops.SrvContainerName container unforcedly.
// timeout is a number of seconds to wait before killing the container process
// ungracefully.
// It returns Docker errors on exit if not successful. DockerStop doesn't return
// an error if the container isn't running.
func DockerStop(srv *def.Service, ops *def.Operation, timeout uint) error {
	// don't limit this to verbose because it takes a few seconds
	// [zr] unless force sets timeout to 0 (for, eg. stdout)
	if timeout != 0 {
		log.WithField("=>", srv.Name).Warn("Stopping (may take a few seconds)")
	}

	log.WithFields(log.Fields{
		"=>":      ops.SrvContainerName,
		"timeout": timeout,
	}).Info("Stopping container")

	_, running := ContainerExists(ops)
	if running {
		log.WithField("=>", ops.SrvContainerName).Debug("Container found running")

		err := stopContainer(ops.SrvContainerName, timeout)
		if err != nil {
			return err
		}
	} else {
		log.WithField("=>", ops.SrvContainerName).Debug("Container found not running")
	}

	log.WithField("=>", ops.SrvContainerName).Info("Container stopped")

	return nil
}

// DockerRename renames the container by removing and recreating it. The container
// is also restarted if it was running before rename. The container ops.SrvContainerName
// is renamed to a new name, constructed using a short given newName.
// DockerRename returns Docker errors on exit or ErrContainerExists
// if the container with the new (long) name exists.
//
//  ops.SrvContainerName  - container name
//  ops.ContainerNumber   - container number
//  ops.ContainerType     - container type
//  ops.Labels            - container creation time labels
//
func DockerRename(ops *def.Operation, newName string) error {
	longNewName := util.ContainersName(ops.ContainerType, newName, ops.ContainerNumber)

	log.WithFields(log.Fields{
		"from": ops.SrvContainerName,
		"to":   longNewName,
	}).Info("Renaming container")

	log.WithField("=>", ops.SrvContainerName).Debug("Checking container exists")
	container, err := util.DockerClient.InspectContainer(ops.SrvContainerName)
	if err != nil {
		return err
	}

	log.WithField("=>", longNewName).Debug("Checking new container exists")
	_, err = util.DockerClient.InspectContainer(longNewName)
	if err == nil {
		return ErrContainerExists
	}

	// Mark if the container's running to restart it later.
	_, wasRunning := ContainerRunning(ops)
	if wasRunning {
		log.Debug("Stopping old container")
		if err := util.DockerClient.StopContainer(container.ID, 5); err != nil {
			log.Debug("Container not stopped")
		}
	}

	log.Debug("Removing container")
	removeOpts := docker.RemoveContainerOptions{
		ID:            container.ID,
		RemoveVolumes: true,
		Force:         true,
	}
	if err := util.DockerClient.RemoveContainer(removeOpts); err != nil {
		return err
	}

	log.Debug("Creating new container")
	createOpts := docker.CreateContainerOptions{
		Name:       longNewName,
		Config:     container.Config,
		HostConfig: container.HostConfig,
	}

	// If VolumesFrom contains links to non-existent containers, remove them.
	var newVolumesFrom []string
	for _, name := range createOpts.HostConfig.VolumesFrom {
		_, err = util.DockerClient.InspectContainer(name)
		if err != nil {
			continue
		}

		name = strings.TrimSuffix(name, ":ro")
		name = strings.TrimSuffix(name, ":rw")

		newVolumesFrom = append(newVolumesFrom, name)
	}
	createOpts.HostConfig.VolumesFrom = newVolumesFrom

	// Rename labels.
	createOpts.Config.Labels = util.Labels(newName, ops)

	newContainer, err := util.DockerClient.CreateContainer(createOpts)
	if err != nil {
		log.Debug("Container not created")
		return err
	}

	// Was running before remove.
	if wasRunning {
		err := util.DockerClient.StartContainer(newContainer.ID, createOpts.HostConfig)
		if err != nil {
			log.Debug("Container not restarted")
		}
	}

	log.WithField("=>", longNewName).Info("Container renamed to")

	return nil
}

// DockerRemove removes the ops.SrvContainerName container.
// If withData is true, the associated data container is also removed.
// If volumes is true, the associated volumes are removed for both containers.
// DockerRemove returns Docker errors on exit if not successful.
func DockerRemove(srv *def.Service, ops *def.Operation, withData, volumes, force bool) error {
	if _, exists := ContainerExists(ops); exists {
		log.WithField("=>", ops.SrvContainerName).Info("Removing container")
		if err := removeContainer(ops.SrvContainerName, volumes, force); err != nil {
			return err
		}
		if withData {
			if _, ext := DataContainerExists(ops); ext {
				log.WithField("=>", ops.DataContainerName).Info("Removing dependent data container")
				if err := removeContainer(ops.DataContainerName, volumes, force); err != nil {
					return err
				}
			}
		}
	} else {
		log.Info("Container does not exist. Cannot remove")
	}

	return nil
}

// ContainerExists returns APIContainers containers list and true
// if the container ops.SrvContainerName exists, otherwise false.
func ContainerExists(ops *def.Operation) (docker.APIContainers, bool) {
	return util.ParseContainers(ops.SrvContainerName, true)
}

// ContainerExists returns APIContainers containers list and true
// if the container ops.SrvContainerName exists and is running,
// otherwise false.
func ContainerRunning(ops *def.Operation) (docker.APIContainers, bool) {
	return util.ParseContainers(ops.SrvContainerName, false)
}

// ContainerExists returns APIContainers containers list and true
// if the container ops.DataContainerName exists and running,
// otherwise false.
func DataContainerExists(ops *def.Operation) (docker.APIContainers, bool) {
	return util.ParseContainers(ops.DataContainerName, true)
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
	name = nameSplit[0]

	repoSplit := strings.Split(nameSplit[0], "/")
	if len(repoSplit) > 2 {
		reg = repoSplit[0]
	}

	opts := docker.PullImageOptions{
		Repository:   name,
		Registry:     reg,
		Tag:          tag,
		OutputStream: os.Stdout,
	}

	if os.Getenv("ERIS_PULL_APPROVE") == "true" {
		opts.OutputStream = nil
	}

	auth := docker.AuthConfiguration{}

	if err := util.DockerClient.PullImage(opts, auth); err != nil {
		return err
	}

	return nil
}

// ----------------------------------------------------------------------------
// ---------------------    Container Core ------------------------------------
// ----------------------------------------------------------------------------
func createContainer(opts docker.CreateContainerOptions) (*docker.Container, error) {
	dockerContainer, err := util.DockerClient.CreateContainer(opts)
	if err != nil {
		if err == docker.ErrNoSuchImage {
			if os.Getenv("ERIS_PULL_APPROVE") != "true" {
				var input string
				log.WithField("image", opts.Config.Image).Warn("The docker image not found locally")
				fmt.Print("Would you like the marmots to pull it from the repository? (y/n): ")
				fmt.Scanln(&input)

				if input == "Y" || input == "y" || input == "YES" || input == "Yes" || input == "yes" {
					log.Debug("User assented to pull")
				} else {
					log.Debug("User refused to pull")
					return nil, fmt.Errorf("Cannot start a container based on an image you will not let me pull.\n")
				}
			} else {
				log.WithField("image", opts.Config.Image).Warn("The Docker image is not found locally")
				log.Warn("The marmots are approved to pull it from the repository on your behalf")
				log.Warn("This could take a few minutes")
			}
			if err := pullImage(opts.Config.Image, nil); err != nil {
				return nil, err
			}
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

func startContainer(opts docker.CreateContainerOptions) error {
	return util.DockerClient.StartContainer(opts.Name, opts.HostConfig)
}

func startInteractiveContainer(opts docker.CreateContainerOptions) error {
	// Trap signals so we can drop out of the container.
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)
	go func() {
		<-c
		log.WithField("=>", opts.Name).Info("Caught signal. Stopping container")
		if err := stopContainer(opts.Name, 5); err != nil {
			log.Errorf("Error stopping container: %v", err)
		}
	}()

	attached := make(chan struct{})
	go func(chan struct{}) {
		attachContainer(opts.Name, attached)
	}(attached)

	// Wait for a console prompt to appear.
	_, ok := <-attached
	if ok {
		attached <- struct{}{}
	}

	if err := startContainer(opts); err != nil {
		return err
	}

	log.WithField("=>", opts.Name).Info("Waiting for container to exit")

	// Set terminal into raw mode, and restore upon container exit.
	savedState, err := term.SetRawTerminal(os.Stdin.Fd())
	if err != nil {
		log.Info("Cannot set the terminal into raw mode")
	} else {
		defer term.RestoreTerminal(os.Stdin.Fd(), savedState)
	}

	if err := waitContainer(opts.Name); err != nil {
		return err
	}

	return nil
}

func attachContainer(id string, attached chan struct{}) error {
	// Use a proxy pipe between os.Stdin and an attached container, so that
	// when the reader end of the pipe is closed, os.Stdin is still open.
	reader, writer := io.Pipe()
	go func() {
		io.Copy(writer, os.Stdin)
	}()

	opts := docker.AttachToContainerOptions{
		Container:    id,
		InputStream:  reader,
		OutputStream: io.MultiWriter(config.GlobalConfig.Writer, config.GlobalConfig.InteractiveWriter),
		ErrorStream:  io.MultiWriter(config.GlobalConfig.ErrorWriter, config.GlobalConfig.InteractiveErrorWriter),
		Logs:         false,
		Stream:       true,
		Stdin:        true,
		Stdout:       true,
		Stderr:       true,
		RawTerminal:  true,
		Success:      attached,
	}

	return util.DockerClient.AttachToContainer(opts)
}

func waitContainer(id string) error {
	exitCode, err := util.DockerClient.WaitContainer(id)
	if exitCode != 0 {
		err1 := fmt.Errorf("Container %s exited with status %d", id, exitCode)
		if err != nil {
			err = fmt.Errorf("%s. Error: %v", err1.Error(), err)
		} else {
			err = err1
		}
	}
	return err
}

func logsContainer(id string, follow bool, tail string) error {
	var writer io.Writer
	var eWriter io.Writer

	if config.GlobalConfig != nil {
		writer = config.GlobalConfig.Writer
		eWriter = config.GlobalConfig.ErrorWriter
	} else {
		writer = os.Stdout
		eWriter = os.Stderr
	}

	opts := docker.LogsOptions{
		Container:    id,
		OutputStream: writer,
		ErrorStream:  eWriter,
		Follow:       follow,
		Stdout:       true,
		Stderr:       true,
		Since:        0,
		Timestamps:   false,
		Tail:         tail,

		RawTerminal: true, // Usually true when the container contains a TTY.
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
	util.PrintInspectionReport(cont, field)

	return nil
}

func stopContainer(id string, timeout uint) error {
	err := util.DockerClient.StopContainer(id, timeout)
	if err != nil {
		return err
	}
	return nil
}

func removeContainer(id string, volumes, force bool) error {
	opts := docker.RemoveContainerOptions{
		ID:            id,
		RemoveVolumes: volumes,
		Force:         force,
	}

	err := util.DockerClient.RemoveContainer(opts)
	if err != nil {
		return err
	}

	return nil
}

func configureInteractiveContainer(srv *def.Service, ops *def.Operation) docker.CreateContainerOptions {
	opts := configureServiceContainer(srv, ops)

	opts.Name = "eris_interactive_" + opts.Name
	if srv.User == "" {
		opts.Config.User = "root"
	} else {
		opts.Config.User = srv.User
	}
	opts.Config.OpenStdin = true
	opts.Config.Tty = true
	opts.Config.AttachStdout = true
	opts.Config.AttachStderr = true
	opts.Config.AttachStdin = true

	if ops.Interactive {
		// if there are args, we overwrite the entrypoint
		// else we just start an interactive shell
		if len(ops.Args) > 0 {
			opts.Config.Entrypoint = ops.Args
		} else {
			opts.Config.Entrypoint = []string{"/bin/bash"}
		}
	} else {
		// use the image's own entrypoint
		opts.Config.Cmd = ops.Args
	}

	// Mount a volume.
	if ops.Volume != "" {
		bind := filepath.Join(dirs.ErisRoot, ops.Volume) + ":" +
			filepath.Join(dirs.ErisContainerRoot, ops.Volume)

		if opts.HostConfig.Binds == nil {
			opts.HostConfig.Binds = []string{bind}
		} else {
			opts.HostConfig.Binds = append(opts.HostConfig.Binds, bind)
		}
	}

	// we expect to link to the main service container
	opts.HostConfig.Links = srv.Links

	return opts
}

func configureServiceContainer(srv *def.Service, ops *def.Operation) docker.CreateContainerOptions {
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
			Tty:             true,
			OpenStdin:       false,
			Env:             srv.Environment,
			Labels:          ops.Labels,
			Image:           srv.Image,
			NetworkDisabled: false,
		},
		HostConfig: &docker.HostConfig{
			Binds:           srv.Volumes,
			Links:           srv.Links,
			PublishAllPorts: ops.PublishAllPorts,
			Privileged:      ops.Privileged,
			ReadonlyRootfs:  false,
			DNS:             srv.DNS,
			DNSSearch:       srv.DNSSearch,
			VolumesFrom:     srv.VolumesFrom,
			CapAdd:          ops.CapAdd,
			CapDrop:         ops.CapDrop,
			RestartPolicy:   docker.NeverRestart(), //default. overide below
			NetworkMode:     "bridge",
		},
	}

	// some fields may be set in the dockerfile and we only want to overwrite if they are present in the service def
	if srv.EntryPoint != "" {
		opts.Config.Entrypoint = strings.Fields(srv.EntryPoint)
	}
	if srv.Command != "" {
		opts.Config.Cmd = strings.Fields(srv.Command)
	}
	if srv.WorkDir != "" {
		opts.Config.WorkingDir = srv.WorkDir
	}

	//[zr] used to be ops.Restart
	if srv.Restart == "always" {
		opts.HostConfig.RestartPolicy = docker.AlwaysRestart()
	} else if strings.Contains(srv.Restart, "max") {
		times, err := strconv.Atoi(strings.Split(srv.Restart, ":")[1])
		if err != nil {
			return docker.CreateContainerOptions{}
		}
		opts.HostConfig.RestartPolicy = docker.RestartOnFailure(times)
	}

	opts.Config.ExposedPorts = make(map[docker.Port]struct{})
	opts.HostConfig.PortBindings = make(map[docker.Port][]docker.PortBinding)
	opts.Config.Volumes = make(map[string]struct{})

	// Don't fill in port bindings if randomizing the ports.
	if !ops.PublishAllPorts {
		for _, port := range srv.Ports {
			pS := strings.Split(port, ":")
			pC := docker.Port(util.PortAndProtocol(pS[len(pS)-1]))

			opts.Config.ExposedPorts[pC] = struct{}{}
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

				opts.HostConfig.PortBindings[pC] = []docker.PortBinding{pH}
			} else {
				pH := docker.PortBinding{
					HostPort: pS[0],
				}
				opts.HostConfig.PortBindings[pC] = []docker.PortBinding{pH}
			}
		}
	}

	for _, vol := range srv.Volumes {
		opts.Config.Volumes[strings.Split(vol, ":")[1]] = struct{}{}
	}

	return opts
}

func configureVolumesFromContainer(ops *def.Operation, service *def.Service) docker.CreateContainerOptions {
	// Set the defaults.
	opts := docker.CreateContainerOptions{
		Name: "eris_exec_" + ops.DataContainerName,
		Config: &docker.Config{
			Image:           path.Join(ver.ERIS_REG_DEF, ver.ERIS_IMG_BASE),
			User:            "root",
			WorkingDir:      dirs.ErisContainerRoot,
			AttachStdout:    true,
			AttachStderr:    true,
			AttachStdin:     true,
			Tty:             true,
			NetworkDisabled: false,
			Labels:          ops.Labels,
		},
		HostConfig: &docker.HostConfig{
			VolumesFrom: []string{ops.DataContainerName},
		},
	}

	opts.Config.OpenStdin = true
	if ops.Interactive {
		opts.Config.Cmd = []string{"/bin/bash"}
	} else {
		opts.Config.Cmd = ops.Args
	}

	// Overwrite some things.
	if service != nil {
		opts.Config.NetworkDisabled = false
		opts.Config.Image = service.Image
		opts.Config.User = service.User
		opts.Config.Env = service.Environment
		opts.HostConfig.Links = service.Links
		opts.Config.Entrypoint = strings.Fields(service.EntryPoint)
	}

	log.WithFields(log.Fields{
		"cmd":        opts.Config.Cmd,
		"entrypoint": opts.Config.Entrypoint,
	}).Debug("Data container configured")

	return opts
}

func configureDataContainer(srv *def.Service, ops *def.Operation, mainContOpts *docker.CreateContainerOptions) (docker.CreateContainerOptions, error) {
	// by default data containers will rely on the image used by
	//   the base service. sometimes, tho, especially for testing
	//   that base image will not be present. in such cases use
	//   the base eris data container.
	if srv.Image == "" {
		srv.Image = path.Join(ver.ERIS_REG_DEF, ver.ERIS_IMG_DATA)
	}

	// Manipulate labels locally.
	labels := make(map[string]string)
	for k, v := range ops.Labels {
		labels[k] = v
	}

	// If connected to a service.
	if mainContOpts != nil {
		// Set the service container's VolumesFrom pointing to the data container.
		mainContOpts.HostConfig.VolumesFrom = append(mainContOpts.HostConfig.VolumesFrom, ops.DataContainerName)

		// Operations are inherited from the service container.
		labels = util.SetLabel(labels, def.LabelType, def.TypeData)

		// Set the data container service label pointing to the service.
		labels = util.SetLabel(labels, def.LabelService, mainContOpts.Name)
	}

	opts := docker.CreateContainerOptions{
		Name: ops.DataContainerName,
		Config: &docker.Config{
			Image:        srv.Image,
			User:         srv.User,
			AttachStdin:  false,
			AttachStdout: false,
			AttachStderr: false,
			Tty:          false,
			OpenStdin:    false,
			Labels:       labels,

			// Data containers do not need to talk to the outside world.
			NetworkDisabled: true,

			// Just gracefully exit. Data containers just need to "exist" not run.
			Entrypoint: []string{"true"},
			Cmd:        []string{},
		},
		HostConfig: &docker.HostConfig{},
	}

	return opts, nil
}
