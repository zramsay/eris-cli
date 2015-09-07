package util

import (
	"strings"

	def "github.com/eris-ltd/eris-cli/definitions"
)

// need to be alot smarter with this
func OverWriteOperations(opsBase, opsOver *def.Operation) {
	opsBase.SrvContainerName = OverWriteString(opsBase.SrvContainerName, opsOver.SrvContainerName)
	opsBase.SrvContainerID = OverWriteString(opsBase.SrvContainerID, opsOver.SrvContainerID)
	opsBase.DataContainerName = OverWriteString(opsBase.DataContainerName, opsOver.DataContainerName)
	opsBase.DataContainerID = OverWriteString(opsBase.DataContainerID, opsOver.DataContainerID)
	opsBase.ContainerNumber = OverWriteInt(opsBase.ContainerNumber, opsOver.ContainerNumber)
	opsBase.Restart = OverWriteString(opsBase.Restart, opsOver.Restart)
	opsBase.Remove = OverWriteBool(opsBase.Remove, opsOver.Remove)
	opsBase.Privileged = OverWriteBool(opsBase.Privileged, opsOver.Privileged)
	opsBase.Attach = OverWriteBool(opsBase.Attach, opsOver.Attach)
	opsBase.AppName = OverWriteString(opsBase.AppName, opsOver.AppName)
	opsBase.DockerHostConn = OverWriteString(opsBase.DockerHostConn, opsOver.DockerHostConn)
	opsBase.Labels = MergeMap(opsBase.Labels, opsOver.Labels)
	opsBase.PublishAllPorts = OverWriteBool(opsBase.PublishAllPorts, opsOver.PublishAllPorts)
}

// AutoMagic will return the highest container number which would represent the most recent
// container to work on unless newCont == true in which case it would return the highest
// container number plus one.
func AutoMagic(cNum int, typ string, newCont bool) int {
	logger.Debugf("Automagic (base) =>\t\t%s:%d\n", typ, cNum)
	contns := ErisContainersByType(typ, true)

	contnums := make([]int, len(contns))
	for i, c := range contns {
		contnums[i] = c.Number
	}

	// get highest container number
	g := 0
	for _, n := range contnums {
		if n >= g {
			g = n
		}
	}

	// ensure outcomes appropriate
	result := g
	if newCont {
		result = g + 1
	}
	if result == 0 {
		result = 1
	}

	logger.Debugf("Automagic (result) =>\t\t%s:%d\n", typ, result)
	return result
}

func OverWriteBool(trumpEr, toOver bool) bool {
	if trumpEr {
		return trumpEr
	}
	return toOver
}

func OverWriteString(trumpEr, toOver string) string {
	if trumpEr != "" {
		return trumpEr
	}
	return toOver
}

func OverWriteInt(trumpEr, toOver int) int {
	if trumpEr != 0 {
		return trumpEr
	}
	return toOver
}

func OverWriteInt64(trumpEr, toOver int64) int64 {
	if trumpEr != 0 {
		return trumpEr
	}
	return toOver
}

func OverWriteSlice(trumpEr, toOver []string) []string {
	if len(trumpEr) != 0 {
		return trumpEr
	}
	return toOver
}

func MergeSlice(mapOne, mapTwo []string) []string {
	for _, v := range mapOne {
		mapTwo = append(mapTwo, v)
	}
	return mapTwo
}

func MergeMap(mapOne, mapTwo map[string]string) map[string]string {
	for k, v := range mapOne {
		mapTwo[k] = v
	}
	return mapTwo
}

// Parse dependencies for internalName (ie. in /etc/hosts), whether to link, and whether to mount volumes-from
// eg. `tinydns:tiny:l` to link to the tinydns container as tiny
func ParseDependency(nameAndOpts string) (name, internalName string, link, mount bool) {
	spl := strings.Split(nameAndOpts, ":")
	name = spl[0]

	// defaults
	link, mount = true, true
	internalName = name

	if len(spl) > 1 {
		if spl[1] != "" {
			internalName = spl[1]
		}
	}
	if len(spl) > 2 {
		switch spl[2] {
		case "l": // link
			link, mount = true, false
		case "m", "v": // mount/volumes-from
			link, mount = false, true
		case "_": // nothing!
			link, mount = false, false
		}
	}
	return
}
