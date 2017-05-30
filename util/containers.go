package util

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/monax/monax/config"
	"github.com/monax/monax/definitions"
	"github.com/monax/monax/log"
	"github.com/monax/monax/version"

	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/docker/docker/pkg/term"
	docker "github.com/fsouza/go-dockerclient"
	"github.com/pborman/uuid"
)

var (
	ErrImagePullTimeout = errors.New("image pull timed out")
)

// Details stores useful container information like its type, short name,
// labels, and Docker inspect output.
type Details struct {
	Type      string
	ShortName string
	FullName  string

	Labels map[string]string
	Info   *docker.Container
}

type cache struct {
	c           map[key]string
	initialized bool
}

type key struct {
	ShortName string
	Type      string
}

var (
	// Cached container names.
	containerCache cache = cache{
		c: make(map[key]string),
	}

	ErrNameNotFound = errors.New("container name not found")
)

// UniqueName() returns a unique container name, prefixed with the short
// entity name, e.g. `keys-6ba7b811-9dad-11d1-80b4-00c04fd430c8`
//
// [pv]: might be a good idea to truncate this long name to ~20 characters
// without much danger of bumping into collisions, e.g. `keys-6ba7b811-9dad-11d1`
// which arguably looks better in logs and `docker ps` output.
func UniqueName(name string) string {
	return name + "-" + uuid.NewRandom().String()
}

// ContainerName returns a long container name by a given container type
// and a short name.
func ContainerName(t, name string) string {
	lookup, err := Lookup(t, name)
	if err != nil {
		containerName := UniqueName(name)

		// Save the container's name in the cache (so that when the
		// ContainerName() is called the second time, the name would
		// be found in the cache).
		containerCache.c[key{Type: t, ShortName: name}] = containerName

		return containerName
	}
	return lookup
}

// Lookup tries the container cache if the container name has been
// generated before for a give type and short name.
func Lookup(t, name string) (string, error) {
	if !containerCache.initialized {
		initializeCache()
	}

	if lookup, ok := containerCache.c[key{Type: t, ShortName: name}]; ok {
		return lookup, nil
	}

	return "", ErrNameNotFound
}

func initializeCache() {
	containers, err := DockerClient.ListContainers(docker.ListContainersOptions{All: true})
	if err != nil {
		return
	}

	for _, c := range containers {
		// A container belongs to Monax if it has the "MONAX" label.
		if _, ok := c.Labels[definitions.LabelMonax]; !ok {
			continue
		}

		// Container doesn't have a name.
		if len(c.Names) == 0 {
			continue
		}

		// Cache names.
		containerCache.c[key{
			ShortName: c.Labels[definitions.LabelShortName],
			Type:      c.Labels[definitions.LabelType],
		}] = strings.TrimLeft(c.Names[0], "/")
	}

	containerCache.initialized = true
}

// ContainerDetails uses Docker inspect API call to retrieve useful
// information about the container. The Docker information is enriched
// with Monax container short name and type, as well as with Monax labels.
func ContainerDetails(name string) *Details {
	info, err := DockerClient.InspectContainer(name)
	if err != nil {
		return &Details{}
	}

	labels := info.Config.Labels

	return &Details{
		FullName:  name,
		Type:      labels[definitions.LabelType],
		ShortName: labels[definitions.LabelShortName],
		Labels:    labels,
		Info:      info,
	}
}

// ServiceContainerName returns a full container name for a given short service name.
func ServiceContainerName(name string) string {
	return ContainerName(definitions.TypeService, name)
}

// ChainContainerName returns a full container name for a given short service name.
func ChainContainerName(name string) string {
	return ContainerName(definitions.TypeChain, name)
}

// DataContainerName returns a full container name for a given short data name.
func DataContainerName(name string) string {
	return ContainerName(definitions.TypeData, name)
}

// MonaxContainers returns a list of full container names matching the filter
// criteria filter, applied to container names and details.
func MonaxContainers(filter func(name string, details *Details) bool, running bool) []string {
	log.WithField("running", running).Info("Discovering Monax containers")

	var monaxContainers []string

	containers, err := DockerClient.ListContainers(docker.ListContainersOptions{All: !running})
	if err != nil {
		return monaxContainers
	}

	for _, c := range containers {
		// Container doesn't have a name.
		if len(c.Names) == 0 {
			continue
		}

		// A container belongs to Monax if it has the "MONAX" label.
		if _, ok := c.Labels[definitions.LabelMonax]; !ok {
			continue
		}

		name := strings.TrimLeft(c.Names[0], "/")
		details := ContainerDetails(name)

		// Cache names.
		containerCache.c[key{
			ShortName: details.Labels[definitions.LabelShortName],
			Type:      details.Labels[definitions.LabelType],
		}] = name

		// Apply filter.
		if !filter(name, details) {
			continue
		}

		monaxContainers = append(monaxContainers, name)
	}

	// Initialized cache means that it contains information
	// about all containers, not just the running ones.
	if !running {
		containerCache.initialized = true
	}
	return monaxContainers
}

// MonaxContainersByType generates a list of container details for a given
// container type and status: running (running=true) or existing (running=false).
func MonaxContainersByType(t string, running bool) []*Details {
	log.WithFields(log.Fields{
		"running": running,
		"type":    t,
	}).Info("Discovering Monax containers")

	var monaxContainers []*Details

	containers, err := DockerClient.ListContainers(docker.ListContainersOptions{All: !running})
	if err != nil {
		return monaxContainers
	}

	for _, c := range containers {
		// Container doesn't have a name.
		if len(c.Names) == 0 {
			continue
		}

		// A container belongs to Monax if it has the "MONAX" label.
		if _, ok := c.Labels[definitions.LabelMonax]; !ok {
			continue
		}

		if c.Labels[definitions.LabelType] != t {
			continue
		}

		name := strings.TrimLeft(c.Names[0], "/")
		details := ContainerDetails(name)

		// Cache names.
		containerCache.c[key{
			ShortName: details.Labels[definitions.LabelShortName],
			Type:      details.Labels[definitions.LabelType],
		}] = name

		monaxContainers = append(monaxContainers, details)
	}

	return monaxContainers
}

// IsService returns true if the service container specified by its short name
// runs (running=true) or exists.
func IsService(name string, running bool) bool {
	return State(definitions.TypeService, name, running)
}

// IsChain returns true if the chain container specified by its short name
// runs (running=true) or exists.
func IsChain(name string, running bool) bool {
	return State(definitions.TypeChain, name, running)
}

// IsChain returns true if the data container specified by its short name
// exists.
func IsData(name string) bool {
	return State(definitions.TypeData, name, false)
}

// State returns true if the container of type t is running (running=true)
// or exists (running = false).
func State(t, name string, running bool) bool {
	containers, err := DockerClient.ListContainers(docker.ListContainersOptions{All: !running})
	if err != nil {
		return false
	}

	for _, c := range containers {
		if name != c.Labels[definitions.LabelShortName] {
			continue
		}
		if t != c.Labels[definitions.LabelType] {
			continue
		}
		return true
	}

	return false
}

// Exists returns true if the container of type t exists.
func Exists(t, name string) bool {
	return State(t, name, false)
}

// Running returns true if the container of type t is running.
func Running(t, name string) bool {
	return State(t, name, true)
}

// FindContainer returns true if the given container specified by
// its long name runs (running=true) or exists.
func FindContainer(name string, running bool) bool {
	containers, err := DockerClient.ListContainers(docker.ListContainersOptions{All: !running})
	if err != nil {
		return false
	}

	for _, c := range containers {
		// Container doesn't have a name.
		if len(c.Names) == 0 {
			continue
		}

		if name == strings.TrimLeft(c.Names[0], "/") {
			return true
		}
	}

	return false
}

// Labels returns a map with container labels, based on the container
// short name and ops settings.
//
//  ops.SrvContainerName  - container name
//  ops.ContainerType     - container type
//
func Labels(name string, ops *definitions.Operation) map[string]string {
	labels := ops.Labels
	if labels == nil {
		labels = make(map[string]string)
	}

	labels[definitions.LabelMonax] = "true"
	labels[definitions.LabelShortName] = name
	labels[definitions.LabelType] = ops.ContainerType

	if user, _, err := config.GitConfigUser(); err == nil {
		labels[definitions.LabelUser] = user
	}

	return labels
}

// SetLabel returns a labels map with additional label name and value.
func SetLabel(labels map[string]string, name, value string) map[string]string {
	if labels == nil {
		labels = make(map[string]string)
	}

	labels[name] = value

	return labels
}

// PullImage pulls an image with or without echo
// to the writer.
func PullImage(image string, writer io.Writer) error {
	var tag string = "latest"

	nameSplit := strings.Split(image, ":")
	if len(nameSplit) == 2 {
		tag = nameSplit[1]
	}
	if len(nameSplit) == 3 {
		tag = nameSplit[2]
	}
	image = nameSplit[0]

	auth := docker.AuthConfiguration{}

	r, w := io.Pipe()
	opts := docker.PullImageOptions{
		Repository:    image,
		Registry:      version.DefaultRegistry,
		Tag:           tag,
		OutputStream:  w,
		RawJSONStream: true,
	}

	if os.Getenv("MONAX_PULL_APPROVE") == "true" {
		opts.OutputStream = ioutil.Discard
	}

	timeoutDuration, err := time.ParseDuration(config.Global.ImagesPullTimeout)
	if err != nil {
		return fmt.Errorf(`Cannot read the ImagesPullTimeout=%q value in monax.toml. Aborting`, config.Global.ImagesPullTimeout)
	}

	ch := make(chan error)
	timeout := make(chan error)
	go func() {
		defer w.Close()
		defer close(ch)

		if err := DockerClient.PullImage(opts, auth); err != nil {
			opts.Repository = image
			opts.Registry = version.BackupRegistry
			if err := DockerClient.PullImage(opts, auth); err != nil {
				ch <- DockerError(err)
			}
		}
	}()
	go func() {
		defer w.Close()
		defer close(timeout)

		<-time.After(timeoutDuration)
		log.Warn("image pull timed out (%v)", timeoutDuration)
		timeout <- ErrImagePullTimeout
	}()
	go jsonmessage.DisplayJSONMessagesStream(r, os.Stdout, os.Stdout.Fd(), term.IsTerminal(os.Stdout.Fd()), nil)
	select {
	case err := <-ch:
		if err != nil {
			return err
		}
	case err := <-timeout:
		return err
	}

	// Spacer.
	log.Warn()

	return nil
}
