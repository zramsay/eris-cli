package util

import (
	"os"
	"path"
	"testing"

	"github.com/monax/monax/config"
	"github.com/monax/monax/definitions"
	"github.com/monax/monax/version"

	docker "github.com/fsouza/go-dockerclient"
)

func init() {
	var err error
	config.Global, err = config.New(os.Stdout, os.Stderr)
	if err != nil {
		os.Exit(1)
	}

	DockerConnect(false, "monax")

	// Pull the necessary image.
	PullImage(path.Join(version.DefaultRegistry, version.ImageKeys), os.Stdout)
}

func TestUniqueName(t *testing.T) {
	pass1 := UniqueName("keys")
	pass2 := UniqueName("keys")

	if pass1 == pass2 {
		t.Fatalf("returned names %v and %v are equal", pass1, pass2)
	}
}

func TestContainerNameSameType(t *testing.T) {
	defer invalidateCache()

	pass1 := ContainerName(definitions.TypeData, "a")
	pass2 := ContainerName(definitions.TypeData, "a")

	if pass1 != pass2 {
		t.Fatalf("returned names %v and %v should differ", pass1, pass2)
	}
}

func TestContainerNameDifferentType(t *testing.T) {
	defer invalidateCache()

	pass1 := ContainerName(definitions.TypeData, "a")
	pass2 := ContainerName(definitions.TypeData, "a")
	pass3 := ContainerName(definitions.TypeService, "b")
	pass4 := ContainerName(definitions.TypeData, "a")

	if pass1 != pass2 {
		t.Fatalf("1. returned names %v and %v should be equal", pass1, pass2)
	}

	if pass2 == pass3 {
		t.Fatalf("2. returned names of different types %v and %v should differ", pass2, pass3)
	}

	if pass2 != pass4 {
		t.Fatalf("3. returned names %v and %v should be equal", pass2, pass4)
	}

}

func TestContainerDetailsSimple(t *testing.T) {
	defer invalidateCache()
	defer RemoveAllMonaxContainers()

	const name = "a"

	if err := create(definitions.TypeService, name); err != nil {
		t.Fatalf("expecting container to be created, got %v", err)
	}

	details := ContainerDetails(ContainerName(definitions.TypeService, name))

	if details.ShortName != name {
		t.Fatalf("expecting short name %q to be returned, got %v", name, details.ShortName)
	}
	if details.Type != definitions.TypeService {
		t.Fatalf("expecting container type %q to be returned, got %v", definitions.TypeService, details.Type)
	}
}

func TestContainerDetailsRunning(t *testing.T) {
	defer invalidateCache()
	defer RemoveAllMonaxContainers()

	const name = "b"

	if err := create(definitions.TypeService, name); err != nil {
		t.Fatalf("expecting container to be created, got %v", err)
	}

	if err := start(ContainerName(definitions.TypeService, name)); err != nil {
		t.Fatalf("expecting container to start, got %v", err)
	}

	details := ContainerDetails(ContainerName(definitions.TypeService, name))

	if details.ShortName != name {
		t.Fatalf("expecting short name %q to be returned, got %v", name, details.ShortName)
	}
	if details.Type != definitions.TypeService {
		t.Fatalf("expecting container type %q to be returned, got %v", definitions.TypeService, details.Type)
	}
}

func TestContainerDetailsEmpty(t *testing.T) {
	defer invalidateCache()

	const name = "a"

	details := ContainerDetails(ContainerName(definitions.TypeData, name))

	if details.ShortName != "" {
		t.Fatalf("1. expecting non-existent container doesn't exist, got %v", details.ShortName)
	}
	if details.Type != "" {
		t.Fatalf("2. expecting non-existent container doesn't exist, got %v", details.Type)
	}
}

func TestContainerLookupNotFound(t *testing.T) {
	defer invalidateCache()

	const name = "a"

	if containerCache.initialized {
		t.Fatalf("expecting the container cache not to be initialized")
	}

	pass1, err := Lookup(definitions.TypeData, name)
	if err == nil {
		t.Fatalf("didn't expect to find the container name in the first pass, got %v", pass1)
	}

	if !containerCache.initialized {
		t.Fatalf("expecting the container cache to be initialized after the first pass")
	}

	pass2, err := Lookup(definitions.TypeData, name)
	if err == nil {
		t.Fatalf("didn't expect to find the container name in the second pass, got %v", pass2)
	}
}

func TestContainerLookupWithContainerName(t *testing.T) {
	defer invalidateCache()

	const name = "a"

	pass1, err := Lookup(definitions.TypeData, name)
	if err == nil {
		t.Fatalf("didn't expect to find the container name in the first pass, got %v", pass1)
	}

	if !containerCache.initialized {
		t.Fatalf("expecting the container cache to be initialized after the first pass")
	}

	ContainerName(definitions.TypeData, name)

	pass2, err := Lookup(definitions.TypeData, name)
	if err != nil {
		t.Fatalf("expecting the container name %v to be in the cache after the second pass, got %v", pass2, err)
	}
}

func TestContainerLookupWithoutContainerName(t *testing.T) {
	defer invalidateCache()
	defer RemoveAllMonaxContainers()

	const name = "a"

	if containerCache.initialized {
		t.Fatalf("expecting the container cache not to be initialized")
	}

	create(definitions.TypeData, name)

	pass1, err := Lookup(definitions.TypeData, name)
	if err != nil {
		t.Fatalf("expecting the container name %v to be in the cache, got %v", pass1, err)
	}

	if !containerCache.initialized {
		t.Fatalf("expecting the container cache to be initialized after the first pass")
	}
}

func TestMonaxContainersExisting(t *testing.T) {
	defer invalidateCache()
	defer RemoveAllMonaxContainers()

	const name = "a"

	create(definitions.TypeData, name)
	create(definitions.TypeService, name)
	start(ContainerName(definitions.TypeService, name))

	containers := MonaxContainers(func(name string, details *Details) bool { return true }, false)

	if len(containers) != 2 {
		t.Fatalf("expecting to find 2 existing containers")
	}

	if !containerCache.initialized {
		t.Fatalf("expecting the container cache to be initialized")
	}
}

func TestMonaxContainersRunning(t *testing.T) {
	defer invalidateCache()
	defer RemoveAllMonaxContainers()

	const name = "a"

	create(definitions.TypeData, name)
	create(definitions.TypeService, name)
	start(ContainerName(definitions.TypeService, name))

	containers := MonaxContainers(func(name string, details *Details) bool { return true }, true)

	if len(containers) != 1 {
		t.Fatalf("expecting to find 1 running containers")
	}

	if !containerCache.initialized {
		t.Fatalf("expecting the container cache to be initialized")
	}
}

func TestMonaxContainersByType(t *testing.T) {
	defer invalidateCache()
	defer RemoveAllMonaxContainers()

	create(definitions.TypeData, "a")
	create(definitions.TypeService, "a")
	create(definitions.TypeService, "b")
	create(definitions.TypeService, "c")
	create(definitions.TypeChain, "a")
	create(definitions.TypeChain, "b")
	start(ContainerName(definitions.TypeService, "b"))
	start(ContainerName(definitions.TypeService, "c"))
	start(ContainerName(definitions.TypeChain, "b"))

	if data1 := MonaxContainersByType(definitions.TypeData, false); len(data1) != 1 {
		t.Fatalf("expecting to find 1 data container, got %v", len(data1))
	}

	if data2 := MonaxContainersByType(definitions.TypeData, true); len(data2) != 0 {
		t.Fatalf("expecting to find 0 running data container, got %v", len(data2))
	}

	if service1 := MonaxContainersByType(definitions.TypeService, false); len(service1) != 3 {
		t.Fatalf("expecting to find 3 service containers, got %v", len(service1))
	}

	if service2 := MonaxContainersByType(definitions.TypeService, true); len(service2) != 2 {
		t.Fatalf("expecting to find 2 running service containers, got %v", len(service2))
	}

	if chain1 := MonaxContainersByType(definitions.TypeChain, false); len(chain1) != 2 {
		t.Fatalf("expecting to find 2 chain containers, got %v", len(chain1))
	}

	if chain2 := MonaxContainersByType(definitions.TypeChain, true); len(chain2) != 1 {
		t.Fatalf("expecting to find 1 running chain containers, got %v", len(chain2))
	}

}

func TestFindContainer(t *testing.T) {
	defer invalidateCache()
	defer RemoveAllMonaxContainers()

	const name = "a"

	create(definitions.TypeData, name)
	create(definitions.TypeService, name)
	start(ContainerName(definitions.TypeService, name))

	if !FindContainer(ContainerName(definitions.TypeData, name), false) {
		t.Fatalf("expecting to find data container existing")
	}

	if FindContainer(ContainerName(definitions.TypeData, name), true) {
		t.Fatalf("expecting to find data container not running")
	}

	if !FindContainer(ContainerName(definitions.TypeService, name), false) {
		t.Fatalf("expecting to find service container existing")
	}

	if !FindContainer(ContainerName(definitions.TypeService, name), true) {
		t.Fatalf("expecting to find service container running")
	}
}

func invalidateCache() {
	containerCache = cache{
		c: make(map[key]string),
	}
}

func create(t, name string) error {
	labels := make(map[string]string)
	labels[definitions.LabelMonax] = "true"
	labels[definitions.LabelShortName] = name
	labels[definitions.LabelType] = t

	keysImage := path.Join(version.DefaultRegistry, version.ImageKeys)
	opts := docker.CreateContainerOptions{
		Name: ContainerName(t, name),
		Config: &docker.Config{
			Image:  keysImage,
			Labels: labels,
		},
	}

	_, err := DockerClient.CreateContainer(opts)
	if err != nil {
		return DockerError(err)
	}
	return nil
}

func start(name string) error {
	return DockerError(DockerClient.StartContainer(name, nil))
}
