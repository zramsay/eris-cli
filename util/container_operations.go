package util

import (
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

//TODO add bool (backlog)
func AutoMagic(cNum int, typ string) (cnum int) {
	contns := ErisContainersByType(typ, true)

	contnums := make([]int, len(contns))
	for i, c := range contns {
		contnums[i] = c.Number
	}

	g := 0
	for _, n := range contnums {
		if n >= g {
			g = n
		}
	}
	return g + 1
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
