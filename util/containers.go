package util

import (
	"strings"

	"github.com/eris-ltd/eris-cli/config"
	def "github.com/eris-ltd/eris-cli/definitions"
	. "github.com/eris-ltd/eris-cli/errors"

	log "github.com/eris-ltd/eris-logger"
	docker "github.com/fsouza/go-dockerclient"
	"github.com/pborman/uuid"
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
)

// UniqueName() returns a unique container name, prefixed with the short
// entity name, e.g. `ipfs-6ba7b811-9dad-11d1-80b4-00c04fd430c8`
//
// [pv]: might be a good idea to truncate this long name to ~20 characters
// without much danger of bumping into collisions, e.g. `ipfs-6ba7b811-9dad-11d1`
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

	return "", ErrContainerNameNotFound
}

func initializeCache() {
	containers, err := DockerClient.ListContainers(docker.ListContainersOptions{All: true})
	if err != nil {
		return
	}

	for _, c := range containers {
		// A container belongs to Eris if it has the "ERIS" label.
		if _, ok := c.Labels[def.LabelEris]; !ok {
			continue
		}

		// Cache names.
		containerCache.c[key{
			ShortName: c.Labels[def.LabelShortName],
			Type:      c.Labels[def.LabelType],
		}] = strings.TrimLeft(c.Names[0], "/")
	}

	containerCache.initialized = true
}

// ContainerDetails uses Docker inspect API call to retrieve useful
// information about the container. The Docker information is enriched
// with Eris container short name and type, as well as with Eris labels.
func ContainerDetails(name string) *Details {
	info, err := DockerClient.InspectContainer(name)
	if err != nil {
		return &Details{}
	}

	labels := info.Config.Labels

	return &Details{
		FullName:  name,
		Type:      labels[def.LabelType],
		ShortName: labels[def.LabelShortName],
		Labels:    labels,
		Info:      info,
	}
}

// ServiceContainerName returns a full container name for a given short service name.
func ServiceContainerName(name string) string {
	return ContainerName(def.TypeService, name)
}

// ChainContainerName returns a full container name for a given short service name.
func ChainContainerName(name string) string {
	return ContainerName(def.TypeChain, name)
}

// DataContainerName returns a full container name for a given short data name.
func DataContainerName(name string) string {
	return ContainerName(def.TypeData, name)
}

// ErisContainers returns a list of full container names matching the filter
// criteria filter, applied to container names and details.
func ErisContainers(filter func(name string, details *Details) bool, running bool) []string {
	log.WithField("running", running).Info("Discovering Eris containers")

	var erisContainers []string

	containers, err := DockerClient.ListContainers(docker.ListContainersOptions{All: !running})
	if err != nil {
		return erisContainers
	}

	for _, c := range containers {
		name := strings.TrimLeft(c.Names[0], "/")

		// A container belongs to Eris if it has the "ERIS" label.
		if _, ok := c.Labels[def.LabelEris]; !ok {
			continue
		}

		details := ContainerDetails(name)

		// Cache names.
		containerCache.c[key{
			ShortName: details.Labels[def.LabelShortName],
			Type:      details.Labels[def.LabelType],
		}] = name

		// Apply filter.
		if !filter(name, details) {
			continue
		}

		erisContainers = append(erisContainers, name)
	}

	// Initialized cache means that it contains information
	// about all containers, not just the running ones.
	if running == false {
		containerCache.initialized = true
	}
	return erisContainers
}

// ErisContainersByType generates a list of container details for a given
// container type and status: running (running=true) or existing (running=false).
func ErisContainersByType(t string, running bool) []*Details {
	log.WithFields(log.Fields{
		"running": running,
		"type":    t,
	}).Info("Discovering Eris containers")

	var erisContainers []*Details

	containers, err := DockerClient.ListContainers(docker.ListContainersOptions{All: !running})
	if err != nil {
		return erisContainers
	}

	for _, c := range containers {
		name := strings.TrimLeft(c.Names[0], "/")

		// A container belongs to Eris if it has the "ERIS" label.
		if _, ok := c.Labels[def.LabelEris]; !ok {
			continue
		}

		if c.Labels[def.LabelType] != t {
			continue
		}

		details := ContainerDetails(name)

		// Cache names.
		containerCache.c[key{
			ShortName: details.Labels[def.LabelShortName],
			Type:      details.Labels[def.LabelType],
		}] = name

		erisContainers = append(erisContainers, details)
	}

	return erisContainers
}

// IsService returns true if the service container specified by its short name
// runs (running=true) or exists.
func IsService(name string, running bool) bool {
	return State(def.TypeService, name, running)
}

// IsChain returns true if the chain container specified by its short name
// runs (running=true) or exists.
func IsChain(name string, running bool) bool {
	return State(def.TypeChain, name, running)
}

// IsChain returns true if the data container specified by its short name
// exists.
func IsData(name string) bool {
	return State(def.TypeData, name, false)
}

// State returns true if the container of type t is running (running=true)
// or exists (running = false).
func State(t, name string, running bool) bool {
	containers, err := DockerClient.ListContainers(docker.ListContainersOptions{All: !running})
	if err != nil {
		return false
	}

	for _, c := range containers {
		if name != c.Labels[def.LabelShortName] {
			continue
		}
		if t != c.Labels[def.LabelType] {
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
func Labels(name string, ops *def.Operation) map[string]string {
	labels := ops.Labels
	if labels == nil {
		labels = make(map[string]string)
	}

	labels[def.LabelEris] = "true"
	labels[def.LabelShortName] = name
	labels[def.LabelType] = ops.ContainerType

	if user, _, err := config.GitConfigUser(); err == nil {
		labels[def.LabelUser] = user
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
