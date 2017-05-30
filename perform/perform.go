package perform

import (
	"archive/tar"
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"unicode"

	"github.com/monax/monax/config"
	"github.com/monax/monax/definitions"
	"github.com/monax/monax/log"
	"github.com/monax/monax/util"
	"github.com/monax/monax/version"

	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/docker/docker/pkg/term"
	docker "github.com/fsouza/go-dockerclient"
)

var (
	ErrContainerExists = errors.New("container exists")
)

// DockerCreateData creates a blank data container. It returns ErrContainerExists
// if such a container exists or other Docker errors.
//
//  ops.DataContainerName  - data container name to be created
//  ops.ContainerType      - container type
//  ops.Labels             - container creation time labels (use LoadDataDefinition)
//
func DockerCreateData(ops *definitions.Operation) error {
	log.WithField("=>", ops.DataContainerName).Info("Creating data container")

	if exists := ContainerExists(ops.DataContainerName); exists {
		log.Info("Data container exists. Not creating")
		return ErrContainerExists
	}

	optsData, err := configureDataContainer(definitions.BlankService(), ops, nil)
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
//  ops.Labels            - container creation time labels (use LoadDataDefinition)
//  ops.Args              - if specified, run these args in a container
//
func DockerRunData(ops *definitions.Operation, service *definitions.Service) (result []byte, err error) {
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
	reader, ok := config.Global.Writer.(*bytes.Buffer)
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
//  ops.Labels       - container creation time labels (use LoadDataDefinition)
//
// See parameter description for DockerRunData.
func DockerExecData(ops *definitions.Operation, service *definitions.Service) (buf *bytes.Buffer, err error) {
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
	stdout, stderr := config.Global.Writer, config.Global.ErrorWriter
	defer func() {
		config.Global.Writer, config.Global.ErrorWriter = stdout, stderr
	}()

	buf = new(bytes.Buffer)
	config.Global.Writer = buf
	config.Global.ErrorWriter = buf

	// Start the container.
	log.WithField("=>", opts.Name).Info("Executing interactive data container")
	if err = startInteractiveContainer(opts, ops.Terminal); err != nil {
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
//  ops.ContainerType     - container type
//  ops.Labels            - container creation time labels
//                          (use LoadServiceDefinition or LoadChainDefinition)
// Container parameters:
//
//  ops.PublishAllPorts   - if true, publish exposed ports to random ports
//  ops.CapAdd            - add linux capabilities (similar to `docker run --cap-add=[]`)
//  ops.CapDrop           - add linux capabilities (similar to `docker run --cap-drop=[]`)
//  ops.Privileged        - if true, give extended privileges
//
func DockerRunService(srv *definitions.Service, ops *definitions.Operation) error {
	log.WithField("=>", ops.SrvContainerName).Info("Running container")

	running := ContainerRunning(ops.SrvContainerName)
	if running {
		log.WithField("=>", ops.SrvContainerName).Info("Container already running. Skipping")
		return nil
	}

	optsServ := configureServiceContainer(srv, ops)

	// Setup data container.
	log.WithField("autodata", srv.AutoData).Info("Manage data containers?")
	if srv.AutoData {
		optsData, err := configureDataContainer(srv, ops, &optsServ)
		if err != nil {
			return err
		}

		if exists := util.FindContainer(ops.DataContainerName, false); exists {
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
	if exists := ContainerExists(ops.SrvContainerName); exists {
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
func DockerExecService(srv *definitions.Service, ops *definitions.Operation) (buf *bytes.Buffer, err error) {
	log.WithField("=>", ops.SrvContainerName).Info("Executing container")

	optsServ := configureInteractiveContainer(srv, ops)

	// Setup data container.
	log.WithField("autodata", srv.AutoData).Info("Manage data containers?")

	if srv.AutoData {
		optsData, err := configureDataContainer(srv, ops, &optsServ)
		if err != nil {
			return nil, err
		}

		if exists := util.FindContainer(ops.DataContainerName, false); exists {
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
	stdout, stderr := config.Global.Writer, config.Global.ErrorWriter
	defer func() {
		config.Global.Writer, config.Global.ErrorWriter = stdout, stderr
	}()

	buf = new(bytes.Buffer)
	config.Global.Writer = buf
	config.Global.ErrorWriter = buf

	// Start the container.
	log.WithFields(log.Fields{
		"=>":              optsServ.Name,
		"data container":  ops.DataContainerName,
		"entrypoint":      optsServ.Config.Entrypoint,
		"workdir":         optsServ.Config.WorkingDir,
		"cmd":             optsServ.Config.Cmd,
		"ports published": optsServ.HostConfig.PublishAllPorts,
		"environment":     optsServ.Config.Env,
		"image":           optsServ.Config.Image,
		"user":            optsServ.Config.User,
		"vols":            optsServ.HostConfig.Binds,
	}).Info("Executing interactive container")

	err = startInteractiveContainer(optsServ, ops.Terminal)

	return buf, err
}

// DockerRebuild recreates the container based on the srv settings template.
// If pullImage is true, it updates the Docker image before recreating
// the container. Timeout is a number of seconds to wait before killing the
// container process ungracefully.
//
//  ops.SrvContainerName  - service or a chain container name to rebuild
//  ops.ContainerType     - container type
//  ops.Labels            - container creation time labels
//
// Also see container parameters for DockerRunService.
func DockerRebuild(srv *definitions.Service, ops *definitions.Operation, pullImage bool, timeout uint) error {
	var wasRunning bool = false

	log.WithField("=>", srv.Name).Info("Rebuilding container")

	if exists := ContainerExists(ops.SrvContainerName); exists {
		if running := ContainerRunning(ops.SrvContainerName); running {
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

	log.WithField("=>", ops.SrvContainerName).Info("Recreating container")
	_, err := createContainer(opts)
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
//  ops.ContainerType     - container type
//  ops.Labels            - container creation time labels
//
// Also see container parameters for DockerRunService.
func DockerPull(srv *definitions.Service, ops *definitions.Operation) error {
	log.WithFields(log.Fields{
		"=>":    srv.Name,
		"image": srv.Image,
	}).Info("Pulling container image for")

	var wasRunning bool

	if exists := ContainerExists(ops.SrvContainerName); exists {
		if running := ContainerRunning(ops.SrvContainerName); running {
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
		if err := util.PullImage(srv.Image, os.Stdout); err != nil {
			return err
		}
	} else {
		if err := util.PullImage(srv.Image, ioutil.Discard); err != nil {
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
func DockerLogs(srv *definitions.Service, ops *definitions.Operation, follow bool, tail string) error {
	log.WithFields(log.Fields{
		"=>":     ops.SrvContainerName,
		"follow": follow,
		"tail":   tail,
	}).Info("Getting logs")
	return logsContainer(ops.SrvContainerName, follow, tail)
}

// DockerInspect displays container ops.SrvContainerName data on the terminal.
// field can be a field name of one of `docker inspect` output or it can be
// either "line" to display a short info line or "all" to display everything. I
// DockerInspect returns Docker errors on exit in not successful.
func DockerInspect(srv *definitions.Service, ops *definitions.Operation, field string) error {
	log.WithField("=>", ops.SrvContainerName).Info("Inspecting")
	return inspectContainer(ops.SrvContainerName, field)
}

// DockerStop stops a running ops.SrvContainerName container unforcedly.
// timeout is a number of seconds to wait before killing the container process
// ungracefully.
// It returns Docker errors on exit if not successful. DockerStop doesn't return
// an error if the container isn't running.
func DockerStop(srv *definitions.Service, ops *definitions.Operation, timeout uint) error {
	log.WithFields(log.Fields{
		"=>":      ops.SrvContainerName,
		"timeout": timeout,
	}).Info("Stopping container")

	running := ContainerRunning(ops.SrvContainerName)
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

// DockerRemove removes the ops.SrvContainerName container.
// If withData is true, the associated data container is also removed.
// If volumes is true, the associated volumes are removed for both containers.
// DockerRemove returns Docker errors on exit if not successful.
func DockerRemove(srv *definitions.Service, ops *definitions.Operation, withData, volumes, force bool) error {
	if exists := ContainerExists(ops.SrvContainerName); exists {
		log.WithField("=>", ops.SrvContainerName).Info("Removing container")
		if err := removeContainer(ops.SrvContainerName, volumes, force); err != nil {
			return err
		}
		if withData {
			if exists := ContainerExists(ops.DataContainerName); exists {
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

// DockerRemoveImage removes the image specified by the name. Image will be
// force removed if force = true. DockerRemoveImage returns Docker errors
// on exit if not successful.
func DockerRemoveImage(name string, force bool) error {
	removeOpts := docker.RemoveImageOptions{
		Force: force,
	}
	return util.DockerError(util.DockerClient.RemoveImageExtended(name, removeOpts))
}

// DockerBuild builds an image with image name and dockerfile text passed
// as parameters. It behaves the same way as the command `docker build -t <image> .`
// where the dockerfile text is in the Dockerfile within the same directory.
// DockerBuild returns Docker errors on exit if not successful.
func DockerBuild(image, dockerfile string) error {
	// Below has been adapted from https://godoc.org/github.com/fsouza/go-dockerclient#Client.BuildImage
	inputbuf := bytes.NewBuffer(nil)
	tr := tar.NewWriter(inputbuf)
	tr.WriteHeader(&tar.Header{Name: "Dockerfile", Size: int64(len([]byte(dockerfile)))})
	tr.Write([]byte(dockerfile))
	tr.Close()

	r, w := io.Pipe()
	imgOpts := docker.BuildImageOptions{
		Name:                image,
		RmTmpContainer:      true,
		ForceRmTmpContainer: true,
		InputStream:         inputbuf,
		OutputStream:        w,
		RawJSONStream:       true,
	}

	ch := make(chan error, 1)
	go func() {
		defer w.Close()
		defer close(ch)

		if err := util.DockerClient.BuildImage(imgOpts); err != nil {
			ch <- err
		}
	}()
	jsonmessage.DisplayJSONMessagesStream(r, os.Stdout, os.Stdout.Fd(), term.IsTerminal(os.Stdout.Fd()), nil)
	if err, ok := <-ch; ok {
		return util.DockerError(err)
	}

	ok, err := checkImageExists(image)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("Image does not exist. Something went wrong. Exiting")
	}

	return nil
}

// ContainerExists returns true if the container specified
// by a long name exists, false otherwise.
func ContainerExists(name string) bool {
	return util.FindContainer(name, false)
}

// ContainerRunning returns true if the container specified
// by a long name is running, false otherwise.
func ContainerRunning(name string) bool {
	return util.FindContainer(name, true)
}

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

	r, w := io.Pipe()
	opts := docker.PullImageOptions{
		Repository:    name,
		Registry:      reg,
		Tag:           tag,
		OutputStream:  w,
		RawJSONStream: true,
	}

	if os.Getenv("MONAX_PULL_APPROVE") == "true" {
		opts.OutputStream = ioutil.Discard
	}

	auth := docker.AuthConfiguration{}

	ch := make(chan error, 1)
	go func() {
		defer w.Close()
		defer close(ch)

		if err := util.DockerClient.PullImage(opts, auth); err != nil {
			ch <- util.DockerError(err)
		}
	}()
	jsonmessage.DisplayJSONMessagesStream(r, writer, os.Stdout.Fd(), term.IsTerminal(os.Stdout.Fd()), nil)
	if err, ok := <-ch; ok {
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
			if os.Getenv("MONAX_PULL_APPROVE") != "true" {
				log.WithField("image", opts.Config.Image).Warn("The Docker image not found locally")
				if util.QueryYesOrNo("Would you like the marmots to pull it from the repository?") == util.Yes {
					log.Debug("User assented to pull")
				} else {
					log.Debug("User refused to pull")
					return nil, fmt.Errorf("Cannot start a container based on an image you will not let me pull")
				}
			} else {
				log.WithField("image", opts.Config.Image).Warn("The Docker image is not found locally")
				log.Warn("The marmots are approved to pull it from the repository on your behalf")
				log.Warn("This could take a few minutes")
			}
			if err := pullImage(opts.Config.Image, os.Stdout); err != nil {
				return nil, util.DockerError(err)
			}
			dockerContainer, err = util.DockerClient.CreateContainer(opts)
			if err != nil {
				return nil, util.DockerError(err)
			}
		} else {
			return nil, util.DockerError(err)
		}
	}
	return dockerContainer, nil
}

func startContainer(opts docker.CreateContainerOptions) error {
	// Setting HostConfig in 'POST /containers/.../start' API call
	// is deprecated since Docker v1.10.0.
	opts.HostConfig = nil

	return util.DockerError(util.DockerClient.StartContainer(opts.Name, opts.HostConfig))
}

func startInteractiveContainer(opts docker.CreateContainerOptions, terminal bool) error {
	// Trap signals so we can drop out of the container.
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		log.WithField("=>", opts.Name).Info("Caught signal. Stopping container")
		if err := stopContainer(opts.Name, 5); err != nil {
			log.Errorf("Error stopping container: %v", err)
		}
	}()

	attached := make(chan struct{})
	cw, err := attachContainer(opts.Name, terminal, attached)
	if err != nil {
		return util.DockerError(err)
	}

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
	if terminal && term.IsTerminal(os.Stdin.Fd()) {
		savedState, err := term.SetRawTerminal(os.Stdin.Fd())
		if err != nil {
			log.Info("Cannot set the terminal into raw mode")
		} else {
			defer term.RestoreTerminal(os.Stdin.Fd(), savedState)
		}
	}

	if err := waitContainer(opts.Name); err != nil {
		return err
	}

	cw.Wait()
	cw.Close()

	return nil
}

func attachContainer(id string, terminal bool, attached chan struct{}) (docker.CloseWaiter, error) {
	opts := docker.AttachToContainerOptions{
		Container:    id,
		OutputStream: io.MultiWriter(config.Global.Writer, config.Global.InteractiveWriter),
		ErrorStream:  io.MultiWriter(config.Global.ErrorWriter, config.Global.InteractiveErrorWriter),
		Logs:         false,
		Stream:       true,
		Stdout:       true,
		Stderr:       true,
		RawTerminal:  true,
		Success:      attached,
	}

	if terminal {
		// Use a proxy pipe between os.Stdin and an attached container, so that
		// when the reader end of the pipe is closed, os.Stdin is still open.
		reader, writer := io.Pipe()
		go func() {
			io.Copy(writer, os.Stdin)
		}()

		opts.InputStream = reader
		opts.Stdin = true
	}

	return util.DockerClient.AttachToContainerNonBlocking(opts)
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

	if config.Global != nil {
		writer = config.Global.Writer
		eWriter = config.Global.ErrorWriter
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
	}

	if err := util.DockerClient.Logs(opts); err != nil {
		return util.DockerError(err)
	}
	return nil
}

func inspectContainer(id, field string) error {
	cont, err := util.DockerClient.InspectContainer(id)
	if err != nil {
		return util.DockerError(err)
	}
	util.PrintInspectionReport(cont, field)

	return nil
}

func stopContainer(id string, timeout uint) error {
	err := util.DockerClient.StopContainer(id, timeout)
	if err != nil {
		return util.DockerError(err)
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
		return util.DockerError(err)
	}

	return nil
}

func configureInteractiveContainer(srv *definitions.Service, ops *definitions.Operation) docker.CreateContainerOptions {
	opts := configureServiceContainer(srv, ops)

	opts.Name = util.UniqueName("interactive")
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
		bind := filepath.Join(ops.Volume) + ":" +
			filepath.Join(config.MonaxContainerRoot, filepath.Base(ops.Volume))

		if opts.HostConfig.Binds == nil {
			opts.HostConfig.Binds = []string{bind}
		} else {
			opts.HostConfig.Binds = append(opts.HostConfig.Binds, bind)
		}
	}

	// We expect to link to the main service container.
	opts.HostConfig.Links = srv.Links

	// Ignore the restart policy of a container.
	opts.HostConfig.RestartPolicy = docker.NeverRestart()

	return opts
}

func configureServiceContainer(srv *definitions.Service, ops *definitions.Operation) docker.CreateContainerOptions {
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
		ports := util.MapPorts(srv.Ports, strings.FieldsFunc(ops.Ports, func(c rune) bool {
			return unicode.IsSpace(c) || c == ','
		}))

		for _, entry := range srv.Ports {
			ip, _, exposed := util.PortComponents(entry)
			published := ports[exposed]

			opts.Config.ExposedPorts[docker.Port(exposed)] = struct{}{}

			opts.HostConfig.PortBindings[docker.Port(exposed)] = []docker.PortBinding{{
				HostPort: published,
				HostIP:   ip,
			}}
		}
	}

	for _, vol := range srv.Volumes {
		if !strings.Contains(vol, ":") {
			continue
		}
		opts.Config.Volumes[strings.Split(vol, ":")[1]] = struct{}{}
	}

	return opts
}

func configureVolumesFromContainer(ops *definitions.Operation, service *definitions.Service) docker.CreateContainerOptions {
	// Set the defaults.
	opts := docker.CreateContainerOptions{
		Name: util.UniqueName("interactive"),
		Config: &docker.Config{
			Image:           path.Join(version.DefaultRegistry, version.ImageData),
			User:            "root",
			WorkingDir:      config.MonaxContainerRoot,
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

func configureDataContainer(srv *definitions.Service, ops *definitions.Operation, mainContOpts *docker.CreateContainerOptions) (docker.CreateContainerOptions, error) {
	// by default data containers will rely on the image used by
	//   the base service. sometimes, tho, especially for testing
	//   that base image will not be present. in such cases use
	//   the base monax data container.
	if srv.Image == "" {
		srv.Image = path.Join(version.DefaultRegistry, version.ImageData)
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
		labels = util.SetLabel(labels, definitions.LabelType, definitions.TypeData)

		// Set the data container service label pointing to the service.
		labels = util.SetLabel(labels, definitions.LabelService, mainContOpts.Name)
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

func checkImageExists(image string) (bool, error) {
	fail := false

	opts := docker.ListImagesOptions{
		Filter: image,
	}

	anImage, err := util.DockerClient.ListImages(opts)
	if err != nil {
		return fail, util.DockerError(err)
	}
	if len(anImage) == 1 {
		return true, nil
	}

	return fail, nil
}
