package util

import (
	"testing"

	docker "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/fsouza/go-dockerclient"
	def "github.com/eris-ltd/eris-cli/definitions"
)

func init() {
	DockerConnect(false, "eris")
}

func TestUniqueName(t *testing.T) {
	pass1 := UniqueName()
	pass2 := UniqueName()

	if pass1 == pass2 {
		t.Fatalf("returned names %v and %v are equal", pass1, pass2)
	}
}

func TestContainerNameSameType(t *testing.T) {
	defer invalidateCache()

	pass1 := ContainerName(def.TypeData, "a")
	pass2 := ContainerName(def.TypeData, "a")

	if pass1 != pass2 {
		t.Fatalf("returned names %v and %v should differ", pass1, pass2)
	}
}

func TestContainerNameDifferentType(t *testing.T) {
	defer invalidateCache()

	pass1 := ContainerName(def.TypeData, "a")
	pass2 := ContainerName(def.TypeData, "a")
	pass3 := ContainerName(def.TypeService, "b")
	pass4 := ContainerName(def.TypeData, "a")

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
	defer RemoveAllErisContainers()

	const name = "a"

	if err := create(def.TypeService, name); err != nil {
		t.Fatalf("expecting container to be created, got %v", err)
	}

	details := ContainerDetails(ContainerName(def.TypeService, name))

	if details.ShortName != name {
		t.Fatalf("expecting short name %q to be returned, got %v", name, details.ShortName)
	}
	if details.Type != def.TypeService {
		t.Fatalf("expecting container type %q to be returned, got %v", def.TypeService, details.Type)
	}
}

func TestContainerDetailsRunning(t *testing.T) {
	defer invalidateCache()
	defer RemoveAllErisContainers()

	const name = "b"

	if err := create(def.TypeService, name); err != nil {
		t.Fatalf("expecting container to be created, got %v", err)
	}

	if err := start(ContainerName(def.TypeService, name)); err != nil {
		t.Fatalf("expecting container to start, got %v", err)
	}

	details := ContainerDetails(ContainerName(def.TypeService, name))

	if details.ShortName != name {
		t.Fatalf("expecting short name %q to be returned, got %v", name, details.ShortName)
	}
	if details.Type != def.TypeService {
		t.Fatalf("expecting container type %q to be returned, got %v", def.TypeService, details.Type)
	}
}

func TestContainerDetailsEmpty(t *testing.T) {
	defer invalidateCache()

	const name = "a"

	details := ContainerDetails(ContainerName(def.TypeData, name))

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

	pass1, err := Lookup(def.TypeData, name)
	if err == nil {
		t.Fatalf("didn't expect to find the container name in the first pass, got %v", pass1)
	}

	if !containerCache.initialized {
		t.Fatalf("expecting the container cache to be initialized after the first pass")
	}

	pass2, err := Lookup(def.TypeData, name)
	if err == nil {
		t.Fatalf("didn't expect to find the container name in the second pass, got %v", pass2)
	}
}

func TestContainerLookupWithContainerName(t *testing.T) {
	defer invalidateCache()

	const name = "a"

	pass1, err := Lookup(def.TypeData, name)
	if err == nil {
		t.Fatalf("didn't expect to find the container name in the first pass, got %v", pass1)
	}

	if !containerCache.initialized {
		t.Fatalf("expecting the container cache to be initialized after the first pass")
	}

	ContainerName(def.TypeData, name)

	pass2, err := Lookup(def.TypeData, name)
	if err != nil {
		t.Fatalf("expecting the container name %v to be in the cache after the second pass, got %v", pass2, err)
	}
}

func TestContainerLookupWithoutContainerName(t *testing.T) {
	defer invalidateCache()
	defer RemoveAllErisContainers()

	const name = "a"

	if containerCache.initialized {
		t.Fatalf("expecting the container cache not to be initialized")
	}

	create(def.TypeData, name)

	pass1, err := Lookup(def.TypeData, name)
	if err != nil {
		t.Fatalf("expecting the container name %v to be in the cache, got %v", pass1, err)
	}

	if !containerCache.initialized {
		t.Fatalf("expecting the container cache to be initialized after the first pass")
	}
}

func TestErisContainersExisting(t *testing.T) {
	defer invalidateCache()
	defer RemoveAllErisContainers()

	const name = "a"

	create(def.TypeData, name)
	create(def.TypeService, name)
	start(ContainerName(def.TypeService, name))

	containers := ErisContainers(func(name string, details *Details) bool { return true }, false)

	if len(containers) != 2 {
		t.Fatalf("expecting to find 2 existing containers")
	}

	if !containerCache.initialized {
		t.Fatalf("expecting the container cache to be initialized")
	}
}

func TestErisContainersRunning(t *testing.T) {
	defer invalidateCache()
	defer RemoveAllErisContainers()

	const name = "a"

	create(def.TypeData, name)
	create(def.TypeService, name)
	start(ContainerName(def.TypeService, name))

	containers := ErisContainers(func(name string, details *Details) bool { return true }, true)

	if len(containers) != 1 {
		t.Fatalf("expecting to find 1 running containers")
	}

	if !containerCache.initialized {
		t.Fatalf("expecting the container cache to be initialized")
	}
}

func TestErisContainersByType(t *testing.T) {
	defer invalidateCache()
	defer RemoveAllErisContainers()

	create(def.TypeData, "a")
	create(def.TypeService, "a")
	create(def.TypeService, "b")
	create(def.TypeService, "c")
	create(def.TypeChain, "a")
	create(def.TypeChain, "b")
	start(ContainerName(def.TypeService, "b"))
	start(ContainerName(def.TypeService, "c"))
	start(ContainerName(def.TypeChain, "b"))

	if data1 := ErisContainersByType(def.TypeData, false); len(data1) != 1 {
		t.Fatalf("expecting to find 1 data container, got %v", len(data1))
	}

	if data2 := ErisContainersByType(def.TypeData, true); len(data2) != 0 {
		t.Fatalf("expecting to find 0 running data container, got %v", len(data2))
	}

	if service1 := ErisContainersByType(def.TypeService, false); len(service1) != 3 {
		t.Fatalf("expecting to find 3 service containers, got %v", len(service1))
	}

	if service2 := ErisContainersByType(def.TypeService, true); len(service2) != 2 {
		t.Fatalf("expecting to find 2 running service containers, got %v", len(service2))
	}

	if chain1 := ErisContainersByType(def.TypeChain, false); len(chain1) != 2 {
		t.Fatalf("expecting to find 2 chain containers, got %v", len(chain1))
	}

	if chain2 := ErisContainersByType(def.TypeChain, true); len(chain2) != 1 {
		t.Fatalf("expecting to find 1 running chain containers, got %v", len(chain2))
	}

}

func TestFindContainer(t *testing.T) {
	defer invalidateCache()
	defer RemoveAllErisContainers()

	const name = "a"

	create(def.TypeData, name)
	create(def.TypeService, name)
	start(ContainerName(def.TypeService, name))

	if FindContainer(ContainerName(def.TypeData, name), false) == false {
		t.Fatalf("expecting to find data container existing")
	}

	if FindContainer(ContainerName(def.TypeData, name), true) == true {
		t.Fatalf("expecting to find data container not running")
	}

	if FindContainer(ContainerName(def.TypeService, name), false) == false {
		t.Fatalf("expecting to find service container existing")
	}

	if FindContainer(ContainerName(def.TypeService, name), true) == false {
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
	labels[def.LabelEris] = "true"
	labels[def.LabelShortName] = name
	labels[def.LabelType] = t

	opts := docker.CreateContainerOptions{
		Name: ContainerName(t, name),
		Config: &docker.Config{
			Image:  "quay.io/eris/keys",
			Labels: labels,
		},
	}

	_, err := DockerClient.CreateContainer(opts)
	if err != nil {
		return err
	}
	return nil
}

func start(name string) error {
	return DockerClient.StartContainer(name, nil)
}

func stop(name string) error {
	return DockerClient.StopContainer(name, 10)
}

func remove(name string) error {
	return DockerClient.RemoveContainer(docker.RemoveContainerOptions{
		ID:            name,
		RemoveVolumes: true,
		Force:         true,
	})
}
