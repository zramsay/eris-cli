package util

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/fsouza/go-dockerclient"
)

// ------------------------------------------------------------------------
// Container Name Functions

type ContainerName struct {
	FullName    string
	DockersName string
	ShortName   string
	Number      int
	Type        string
	ContainerID string
}

func ContainersName(typ, name string, number int) string {
	return ContainerAssemble(typ, name, number).FullName
}

func ContainersNumber(containerName string) int {
	return ContainerDisassemble(containerName).Number
}

func ContainersType(containerName string) string {
	return ContainerDisassemble(containerName).Type
}

func ContainersShortName(containerName string) string {
	return ContainerDisassemble(containerName).ShortName
}

func ContainerAssemble(typ, name string, number int) *ContainerName {
	full := fmt.Sprintf("eris_%s_%s_%d", typ, name, number)

	return &ContainerName{
		FullName:    full,
		DockersName: "/" + full,
		ShortName:   name,
		Type:        typ,
		Number:      number,
	}
}

func ContainerDisassemble(containerName string) *ContainerName {
	pop := strings.Split(containerName, "_")

	if len(pop) < 4 {
		logger.Debugln("The marmots cannot disassemble container name", containerName)
		return &ContainerName{}
	}

	if !(pop[0] == "eris" || pop[0] == "/eris") {
		logger.Debugln("The marmots cannot disassemble container name", containerName)
		return &ContainerName{}
	}

	typ := pop[1]
	srt := strings.Join(pop[2:len(pop)-1], "_")
	num, err := strconv.Atoi(pop[len(pop)-1])
	if err != nil {
		logger.Debugln("The marmots cannot disassemble container name", containerName)
		return &ContainerName{}
	}

	// remove the leading slash docker adds
	containerName = strings.Replace(containerName, "/", "", 1)

	return &ContainerName{
		FullName:    containerName,
		DockersName: "/" + containerName,
		Type:        typ,
		Number:      num,
		ShortName:   srt,
	}
}

func ServiceContainersName(name string, number int) string {
	return ContainersName("service", name, number)
}

func ChainContainersName(name string, number int) string {
	return ContainersName("chain", name, number)
}

func DataContainersName(name string, number int) string {
	return ContainersName("data", name, number)
}

func ServiceToDataContainer(serviceContainerName string) string {
	return strings.Replace(serviceContainerName, "service", "data", 1)
}

func ChainToDataContainer(chainContainerName string) string {
	return strings.Replace(chainContainerName, "chain", "data", 1)
}

func DataContainerToService(dataContainerName string) string {
	return strings.Replace(dataContainerName, "data", "service", 1)
}

func DataContainerToChain(dataContainerName string) string {
	return strings.Replace(dataContainerName, "data", "chain", 1)
}

// ------------------------------------------------------------------------
// Container Find and Assemble Functions

func ErisContainersByType(typ string, running bool) []*ContainerName {
	containers := []*ContainerName{}
	r := erisRegExp(typ)      // eris containers
	q := erisRegExpLinks(typ) // skip past these -- they're containers docker makes especially to handle links

	contns, err := DockerClient.ListContainers(docker.ListContainersOptions{All: running})

	if len(contns) == 0 || err != nil {
		logger.Infoln("There are no containers.")
		if err != nil {
			logger.Debugf("Marmot error duing DockerClient.ListContainers: %v\n", err)
		}
		return containers
	}

	for _, con := range contns {
		for _, c := range con.Names {
			if q.MatchString(c) {
				continue
			}
			if r.MatchString(c) {
				c = strings.Replace(c, "/", "", 1) // Docker's leading slash
				// logger.Debugf("Found Eris Container =>\t\t%s\n", c)
				cont := ContainerDisassemble(c)
				cont.ContainerID = con.ID
				containers = append(containers, cont)
			}
		}
	}

	return containers
}

func HowManyContainers(name, typ string, running bool) int {
	res := 0
	conts := ErisContainersByType(typ, running)
	if len(conts) == 0 {
		return res
	}

	for _, c := range conts {
		if c.ShortName == name {
			res++
		}
	}
	return res
}

func HowManyContainersExisting(name, typ string) int {
	return HowManyContainers(name, typ, true)
}

func HowManyContainersRunning(name, typ string) int {
	return HowManyContainers(name, typ, false)
}

func ServiceContainers(running bool) []*ContainerName {
	return ErisContainersByType("service", running)
}

func ServiceContainerNames(running bool) []string {
	// logger.Debugf("Populating current containrs =>\tall? %t\n", running)
	a := ServiceContainers(running)
	b := []string{}
	for _, c := range a {
		b = append(b, c.ShortName)
	}
	// logger.Debugf("Containers I found =>\t\t%v\n", b)
	return b
}

func ServiceContainerFullNames(running bool) []string {
	a := ServiceContainers(running)
	b := []string{}
	for _, c := range a {
		b = append(b, c.FullName)
	}
	return b
}

func ChainContainers(running bool) []*ContainerName {
	return ErisContainersByType("chain", running)
}

func ChainContainerNames(running bool) []string {
	a := ChainContainers(running)
	b := []string{}
	for _, c := range a {
		b = append(b, c.ShortName)
	}
	return b
}

func ChainContainerFullNames(running bool) []string {
	a := ChainContainers(running)
	b := []string{}
	for _, c := range a {
		b = append(b, c.FullName)
	}
	return b
}

func DataContainers() []*ContainerName {
	return ErisContainersByType("data", true)
}

func DataContainerNames() []string {
	a := DataContainers()
	b := []string{}
	for _, c := range a {
		b = append(b, strings.Replace(c.ShortName, "_", " ", -1))
	}
	return b
}

func DataContainerFullNames() []string {
	a := DataContainers()
	b := []string{}
	for _, c := range a {
		b = append(b, c.FullName)
	}
	return b
}

func FindServiceContainer(srvName string, number int, running bool) *ContainerName {
	for _, srv := range ServiceContainers(running) {
		if srv.ShortName == srvName {
			if srv.Number == number {
				logger.Debugf("Found Service container =>\t%s:%s and %d:%d\n", srv.ShortName, srvName, srv.Number, number)
				return srv
			}
		}
	}
	logger.Infof("Could not find service container =>\t%s:%d\n", srvName, number)
	return nil
}

// TODO: populate the ContainerID during this portion of the general sequence
func IsServiceContainer(name string, number int, running bool) bool {
	if FindServiceContainer(name, number, running) == nil {
		return false
	}
	return true
}

func FindChainContainer(name string, number int, running bool) *ContainerName {
	for _, srv := range ChainContainers(running) {
		if srv.ShortName == name {
			if srv.Number == number {
				logger.Debugf("Found Chain container =>\t%s:%s and %d:%d\n", srv.ShortName, name, srv.Number, number)
				return srv
			}
		}
	}
	logger.Infof("Could not find chain container =>\t%s:%d\n", name, number)
	return nil
}

func IsChainContainer(name string, number int, running bool) bool {
	if FindChainContainer(name, number, running) == nil {
		return false
	}
	return true
}

func FindDataContainer(name string, number int) *ContainerName {
	for _, srv := range DataContainers() {
		if srv.ShortName == name {
			if srv.Number == number {
				logger.Debugf("Found Data container =>\t\t%s:%s and %d:%d\n", srv.ShortName, name, srv.Number, number)
				return srv
			}
		}
	}
	logger.Infof("Could not find data container =>\t%s:%d\n", name, number)
	return nil
}

func IsDataContainer(name string, number int) bool {
	if FindDataContainer(name, number) == nil {
		return false
	}
	return true
}

func erisRegExp(typ string) *regexp.Regexp {
	return regexp.MustCompile(fmt.Sprintf(`\/eris_%s_(.+?)_(\d+)`, typ))
}

// docker has this weird thing where it returns links as individual
// container (as in there is the container of two linked services and
// the linkage between them is actually its own containers). this explains
// the leading hash on containers. the q regexp is to filer out these
// links from the return list as they are irrelevant to the information
// desired by this function. and frankly they give false answers to
// IsServiceRunning and ls,ps,known functions.
func erisRegExpLinks(typ string) *regexp.Regexp {
	return regexp.MustCompile(fmt.Sprintf(`\A\/eris_%s_(.+?)_\d+/(.+?)\z`, typ))
}

// here temporarily
func NameAndNumber(name string, num int) string {
	return fmt.Sprintf("%s_%d", name, num)
}
