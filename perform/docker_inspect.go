package perform

import (
	"fmt"
	"strings"

	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/fsouza/go-dockerclient"
)

func PrintInspectionReport(cont *docker.Container, field string) {
	fieldSplit := strings.Split(field, ".")
	maj := fieldSplit[0]
	min := ""
	if len(fieldSplit) != 1 {
		min = fieldSplit[1]
	}
	switch maj {
	case "id":
		fmt.Println(cont.ID)
	case "image":
		fmt.Println(cont.Image)
	case "created":
		fmt.Println(cont.Created)
	case "path":
		fmt.Println(cont.Path)
	case "args":
		fmt.Println(cont.Args)
	case "config":
		printInspectionReportConfig(cont, min)
	case "state":
		fmt.Println(cont.State)
	case "node":
		// printInspectionReportNode(cont, min) // doesn't currently work
	case "network_settings":
		printInspectionReportNetwork(cont, min)
	case "sys_init_path":
		fmt.Println(cont.SysInitPath)
	case "resolv_conf_path":
		fmt.Println(cont.ResolvConfPath)
	case "hostname_path":
		fmt.Println(cont.HostnamePath)
	case "hosts_path":
		fmt.Println(cont.HostsPath)
	case "name":
		fmt.Println(cont.Name)
	case "driver":
		fmt.Println(cont.Driver)
	case "volumes":
		fmt.Println(cont.Volumes)
	case "volumes_rw":
		fmt.Println(cont.VolumesRW)
	case "host_config":
		printInspectionReportHostConfig(cont, min)
	case "exec_ids":
		fmt.Println(cont.ExecIDs)
	case "app_armor_profile":
		fmt.Println(cont.AppArmorProfile)
	default:
		fmt.Printf("ID\t=>\t%v\n", cont.ID)
		fmt.Printf("Image\t=>\t%v\n", cont.Image)
		fmt.Printf("Created\t=>\t%v\n", cont.Created)
		fmt.Printf("Path\t=>\t%v\n", cont.Path)
		fmt.Printf("Args\t=>\t%v\n", cont.Args)
		printInspectionReportConfig(cont, min)
		fmt.Printf("State\t=>\t%v\n", cont.State)
		// printInspectionReportNode(cont, min) // doesn't currently work
		printInspectionReportNetwork(cont, min)
		fmt.Printf("SysInitPath\t=>\t%v\n", cont.SysInitPath)
		fmt.Printf("ResolvConfPath\t=>\t%v\n", cont.ResolvConfPath)
		fmt.Printf("HostnamePath\t=>\t%v\n", cont.HostnamePath)
		fmt.Printf("HostsPath\t=>\t%v\n", cont.HostsPath)
		fmt.Printf("Name\t=>\t%v\n", cont.Name)
		fmt.Printf("Driver\t=>\t%v\n", cont.Driver)
		fmt.Printf("Volumes\t=>\t%v\n", cont.Volumes)
		fmt.Printf("VolumesRW\t=>\t%v\n", cont.VolumesRW)
		printInspectionReportHostConfig(cont, min)
		fmt.Printf("ExecIDs\t\t=>\t%v\n", cont.ExecIDs)
		fmt.Printf("AppArmorProfile\t=>\t%v\n", cont.AppArmorProfile)
	}
}

func printInspectionReportConfig(cont *docker.Container, field string) {
	switch field {
	case "host_name":
		fmt.Println(cont.Config.Hostname)
	case "domain_name":
		fmt.Println(cont.Config.Domainname)
	case "user":
		fmt.Println(cont.Config.User)
	case "memory":
		fmt.Println(cont.Config.Memory)
	case "memory_swap":
		fmt.Println(cont.Config.MemorySwap)
	case "cpu_shares":
		fmt.Println(cont.Config.CPUShares)
	case "cpu_set":
		fmt.Println(cont.Config.CPUSet)
	case "attach_stdin":
		fmt.Println(cont.Config.AttachStdin)
	case "attach_stdout":
		fmt.Println(cont.Config.AttachStdout)
	case "attach_stderr":
		fmt.Println(cont.Config.AttachStderr)
	case "ports":
		fmt.Println(cont.Config.PortSpecs)
	case "expose":
		fmt.Println(cont.Config.ExposedPorts)
	case "tty":
		fmt.Println(cont.Config.Tty)
	case "open_stdin":
		fmt.Println(cont.Config.OpenStdin)
	case "stdin_once":
		fmt.Println(cont.Config.StdinOnce)
	case "environment":
		fmt.Println(cont.Config.Env)
	case "cmd":
		fmt.Println(cont.Config.Cmd)
	case "dns":
		fmt.Println(cont.Config.DNS)
	case "image":
		fmt.Println(cont.Config.Image)
	case "volumes":
		fmt.Println(cont.Config.Volumes)
	case "volumes_from":
		fmt.Println(cont.Config.VolumesFrom)
	case "work_dir":
		fmt.Println(cont.Config.WorkingDir)
	case "mac_address":
		fmt.Println(cont.Config.MacAddress)
	case "entry_point":
		fmt.Println(cont.Config.Entrypoint)
	case "network_disabled":
		fmt.Println(cont.Config.NetworkDisabled)
	case "security_opts":
		fmt.Println(cont.Config.SecurityOpts)
	case "on_build":
		fmt.Println(cont.Config.OnBuild)
	case "labels":
		fmt.Println(cont.Config.Labels)
	default:
		fmt.Println("Config\t=>\t")
		fmt.Printf("Config\t=>\tHostname\t=>\t%v\n", cont.Config.Hostname)
		fmt.Printf("Config\t=>\tDomainname\t=>\t%v\n", cont.Config.Domainname)
		fmt.Printf("Config\t=>\tUser\t\t=>\t%v\n", cont.Config.User)
		fmt.Printf("Config\t=>\tMemory\t\t=>\t%v\n", cont.Config.Memory)
		fmt.Printf("Config\t=>\tMemorySwap\t=>\t%v\n", cont.Config.MemorySwap)
		fmt.Printf("Config\t=>\tCPUShares\t=>\t%v\n", cont.Config.CPUShares)
		fmt.Printf("Config\t=>\tCPUSet\t\t=>\t%v\n", cont.Config.CPUSet)
		fmt.Printf("Config\t=>\tAttachStdin\t=>\t%v\n", cont.Config.AttachStdin)
		fmt.Printf("Config\t=>\tAttachStdout\t=>\t%v\n", cont.Config.AttachStdout)
		fmt.Printf("Config\t=>\tAttachStderr\t=>\t%v\n", cont.Config.AttachStderr)
		fmt.Printf("Config\t=>\tPortSpecs\t=>\t%v\n", cont.Config.PortSpecs)
		fmt.Printf("Config\t=>\tExposedPorts\t=>\t%#v\n", cont.Config.ExposedPorts)
		fmt.Printf("Config\t=>\tTty\t\t=>\t%v\n", cont.Config.Tty)
		fmt.Printf("Config\t=>\tOpenStdin\t=>\t%v\n", cont.Config.OpenStdin)
		fmt.Printf("Config\t=>\tStdinOnce\t=>\t%v\n", cont.Config.StdinOnce)
		fmt.Printf("Config\t=>\tEnv\t\t=>\t%v\n", cont.Config.Env)
		fmt.Printf("Config\t=>\tCmd\t\t=>\t%v\n", cont.Config.Cmd)
		fmt.Printf("Config\t=>\tDNS\t\t=>\t%v\n", cont.Config.DNS)
		fmt.Printf("Config\t=>\tImage\t\t=>\t%v\n", cont.Config.Image)
		fmt.Printf("Config\t=>\tVolumes\t\t=>\t%#v\n", cont.Config.Volumes)
		fmt.Printf("Config\t=>\tVolumesFrom\t=>\t%v\n", cont.Config.VolumesFrom)
		fmt.Printf("Config\t=>\tWorkingDir\t=>\t%v\n", cont.Config.WorkingDir)
		fmt.Printf("Config\t=>\tMacAddress\t=>\t%v\n", cont.Config.MacAddress)
		fmt.Printf("Config\t=>\tEntrypoint\t=>\t%v\n", cont.Config.Entrypoint)
		fmt.Printf("Config\t=>\tNetworkDisabled\t=>\t%v\n", cont.Config.NetworkDisabled)
		fmt.Printf("Config\t=>\tSecurityOpts\t=>\t%v\n", cont.Config.SecurityOpts)
		fmt.Printf("Config\t=>\tOnBuild\t\t=>\t%v\n", cont.Config.OnBuild)
		fmt.Printf("Config\t=>\tLabels\t\t=>\t%v\n", cont.Config.Labels)
	}
}

func printInspectionReportNode(cont *docker.Container, field string) {
	switch field {
	case "id":
		fmt.Println(cont.Node.ID)
	case "ip":
		fmt.Println(cont.Node.IP)
	case "addr":
		fmt.Println(cont.Node.Addr)
	case "name":
		fmt.Println(cont.Node.Name)
	case "cpus":
		fmt.Println(cont.Node.CPUs)
	case "memory":
		fmt.Println(cont.Node.Memory)
	case "labels":
		fmt.Println(cont.Node.Labels)
	default:
		fmt.Println("Node\t=>\t")
		fmt.Printf("Node\t=>\tID\t=>\t%v\n", cont.Node.ID)
		fmt.Printf("Node\t=>\tIP\t=>\t%v\n", cont.Node.IP)
		fmt.Printf("Node\t=>\tAddr\t=>\t%v\n", cont.Node.Addr)
		fmt.Printf("Node\t=>\tName\t=>\t%v\n", cont.Node.Name)
		fmt.Printf("Node\t=>\tCPUs\t=>\t%v\n", cont.Node.CPUs)
		fmt.Printf("Node\t=>\tMemory\t=>\t%v\n", cont.Node.Memory)
		fmt.Printf("Node\t=>\tLabels\t=>\t%v\n", cont.Node.Labels)
	}
}

func printInspectionReportNetwork(cont *docker.Container, field string) {
	switch field {
	case "ip_address":
		fmt.Println(cont.NetworkSettings.IPAddress)
	case "ip_prefix_len":
		fmt.Println(cont.NetworkSettings.IPPrefixLen)
	case "gateway":
		fmt.Println(cont.NetworkSettings.Gateway)
	case "bridge":
		fmt.Println(cont.NetworkSettings.Bridge)
	case "port_mapping":
		fmt.Println(cont.NetworkSettings.PortMapping)
	case "ports":
		fmt.Println(cont.NetworkSettings.Ports)
	default:
		fmt.Println("Network\t=>\t")
		fmt.Printf("Network\t=>\tIPAddress\t=>\t%v\n", cont.NetworkSettings.IPAddress)
		fmt.Printf("Network\t=>\tIPPrefixLen\t=>\t%v\n", cont.NetworkSettings.IPPrefixLen)
		fmt.Printf("Network\t=>\tGateway\t\t=>\t%v\n", cont.NetworkSettings.Gateway)
		fmt.Printf("Network\t=>\tBridge\t\t=>\t%v\n", cont.NetworkSettings.Bridge)
		fmt.Printf("Network\t=>\tPortMapping\t=>\t%v\n", cont.NetworkSettings.PortMapping)
		fmt.Printf("Network\t=>\tPorts\t\t=>\t%v\n", cont.NetworkSettings.Ports)
	}
}

func printInspectionReportHostConfig(cont *docker.Container, field string) {
	switch field {
	case "binds":
		fmt.Println(cont.HostConfig.Binds)
	case "cap_add":
		fmt.Println(cont.HostConfig.CapAdd)
	case "cap_drop":
		fmt.Println(cont.HostConfig.CapDrop)
	case "container_id_file":
		fmt.Println(cont.HostConfig.ContainerIDFile)
	case "lxc_conf":
		fmt.Println(cont.HostConfig.LxcConf)
	case "privileged":
		fmt.Println(cont.HostConfig.Privileged)
	case "port_bindings":
		fmt.Println(cont.HostConfig.PortBindings)
	case "links":
		fmt.Println(cont.HostConfig.Links)
	case "publish_all_ports":
		fmt.Println(cont.HostConfig.PublishAllPorts)
	case "dns":
		fmt.Println(cont.HostConfig.DNS)
	case "dns_search":
		fmt.Println(cont.HostConfig.DNSSearch)
	case "extra_hosts":
		fmt.Println(cont.HostConfig.ExtraHosts)
	case "volumes_from":
		fmt.Println(cont.HostConfig.VolumesFrom)
	case "network_mode":
		fmt.Println(cont.HostConfig.NetworkMode)
	case "ipc_mode":
		fmt.Println(cont.HostConfig.IpcMode)
	case "pid_mode":
		fmt.Println(cont.HostConfig.PidMode)
	case "restart_policy":
		fmt.Println(cont.HostConfig.RestartPolicy)
	case "devices":
		fmt.Println(cont.HostConfig.Devices)
	case "log_config":
		fmt.Println(cont.HostConfig.LogConfig)
	case "readonly_rootfs":
		fmt.Println(cont.HostConfig.ReadonlyRootfs)
	case "security_opts":
		fmt.Println(cont.HostConfig.SecurityOpt)
	case "cgroup_parent":
		fmt.Println(cont.HostConfig.CgroupParent)
	case "memory":
		fmt.Println(cont.HostConfig.Memory)
	case "memory_swap":
		fmt.Println(cont.HostConfig.MemorySwap)
	case "cpu_shares":
		fmt.Println(cont.HostConfig.CPUShares)
	case "cpu_set":
		fmt.Println(cont.HostConfig.CPUSet)
	case "cpu_quota":
		fmt.Println(cont.HostConfig.CPUQuota)
	case "cpu_period":
		fmt.Println(cont.HostConfig.CPUPeriod)
	default:
		fmt.Println("HostConfig\t=>\t")
		fmt.Printf("HostConfig\t=>\tBinds\t=>\t%v\n", cont.HostConfig.Binds)
		fmt.Printf("HostConfig\t=>\tCapAdd\t=>\t%v\n", cont.HostConfig.CapAdd)
		fmt.Printf("HostConfig\t=>\tCapDrop\t=>\t%v\n", cont.HostConfig.CapDrop)
		fmt.Printf("HostConfig\t=>\tContainerIDFile\t=>\t%v\n", cont.HostConfig.ContainerIDFile)
		fmt.Printf("HostConfig\t=>\tLxcConf\t=>\t%v\n", cont.HostConfig.LxcConf)
		fmt.Printf("HostConfig\t=>\tPrivileged\t=>\t%v\n", cont.HostConfig.Privileged)
		fmt.Printf("HostConfig\t=>\tPortBindings\t=>\t%v\n", cont.HostConfig.PortBindings)
		fmt.Printf("HostConfig\t=>\tLinks\t=>\t%v\n",cont.HostConfig.Links)
		fmt.Printf("HostConfig\t=>\tPublishAllPorts\t=>\t%v\n", cont.HostConfig.PublishAllPorts)
		fmt.Printf("HostConfig\t=>\tDNS\t=>\t%v\n", cont.HostConfig.DNS)
		fmt.Printf("HostConfig\t=>\tDNSSearch\t=>\t%v\n", cont.HostConfig.DNSSearch)
		fmt.Printf("HostConfig\t=>\tExtraHosts\t=>\t%v\n", cont.HostConfig.ExtraHosts)
		fmt.Printf("HostConfig\t=>\tVolumesFrom\t=>\t%v\n", cont.HostConfig.VolumesFrom)
		fmt.Printf("HostConfig\t=>\tIpcMode\t=>\t%v\n", cont.HostConfig.IpcMode)
		fmt.Printf("HostConfig\t=>\tPidMode\t=>\t%v\n", cont.HostConfig.PidMode)
		fmt.Printf("HostConfig\t=>\tRestartPolicy\t=>\t%v\n", cont.HostConfig.RestartPolicy)
		fmt.Printf("HostConfig\t=>\tDevices\t=>\t%v\n", cont.HostConfig.Devices)
		fmt.Printf("HostConfig\t=>\tLogConfig\t=>\t%v\n", cont.HostConfig.LogConfig)
		fmt.Printf("HostConfig\t=>\tReadonlyRootfs\t=>\t%v\n", cont.HostConfig.ReadonlyRootfs)
		fmt.Printf("HostConfig\t=>\tSecurityOpt\t=>\t%v\n", cont.HostConfig.SecurityOpt)
		fmt.Printf("HostConfig\t=>\tCgroupParent\t=>\t%v\n", cont.HostConfig.CgroupParent)
		fmt.Printf("HostConfig\t=>\tMemory\t=>\t%v\n", cont.HostConfig.Memory)
		fmt.Printf("HostConfig\t=>\tMemorySwap\t=>\t%v\n", cont.HostConfig.MemorySwap)
		fmt.Printf("HostConfig\t=>\tCPUShares\t=>\t%v\n", cont.HostConfig.CPUShares)
		fmt.Printf("HostConfig\t=>\tCPUSet\t=>\t%v\n", cont.HostConfig.CPUSet)
		fmt.Printf("HostConfig\t=>\tCPUQuota\t=>\t%v\n", cont.HostConfig.CPUQuota)
		fmt.Printf("HostConfig\t=>\tCPUPeriod\t=>\t%v\n", cont.HostConfig.CPUPeriod)
	}
}
