package perform

import (
  "fmt"
  "os"
  "path"
  "regexp"
  "runtime"
  "strings"
  "strconv"

  "github.com/eris-ltd/eris-cli/util"

  "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/fsouza/go-dockerclient"
)

// Build against Docker cli...
//   Client version: 1.6.2
//   Client API version: 1.18
// Verified against ...
//   Client version: 1.6.2
//   Client API version: 1.18
func DockerRun(srv *util.Service, verbose bool) {
  var id_main string
  var id_data string
  var optsServ docker.CreateContainerOptions
  var optsData docker.CreateContainerOptions
  var dataCont docker.APIContainers
  var dataContCreated *docker.Container

  if verbose {
    fmt.Println("Starting Service: " + srv.Name)
  }

  // configure
  optsServ = configureContainer(srv)
  srv.Volumes = fixDirs(srv.Volumes)
  if srv.DataContainer {
    fmt.Println("You've asked me to manage the data containers. I shall do so good human.")
    optsData = configureDataContainer(srv, &optsServ)
  }

  // check existence || create the container
  if servCont, exists := ContainerExists(srv); exists {
    if verbose {
      fmt.Println("Service Container already exists, am not creating.")
    }
    if srv.DataContainer {
      if dataCont, exists = parseContainers(srv.DataContainerName, true); exists {
        if verbose {
          fmt.Println("Data Container already exists, am not creating.")
        }
        id_data = dataCont.ID
      } else {
        if verbose {
          fmt.Println("Data Container does not exist, creating.")
        }
        dataContCreated := createContainer(optsData)
        id_data = dataContCreated.ID
      }
    }

    id_main = servCont.ID

  } else {
    if verbose {
      fmt.Println("Service Container does not exist, creating.")
    }

    if srv.DataContainer {
      if dataCont, exists = parseContainers(srv.DataContainerName, true); exists {
        if verbose {
          fmt.Println("Data Container already exists, am not creating.")
        }
        id_data = dataCont.ID
      } else {
        if verbose {
          fmt.Println("Data Container does not exist, creating.")
        }
        dataContCreated = createContainer(optsData)
        id_data = dataContCreated.ID
      }
    }

    servContCreated := createContainer(optsServ)
    id_main = servContCreated.ID
  }

  // start the container
  if verbose {
    fmt.Println("Starting ServiceContainer ID: " + id_main)
    if srv.DataContainer {
      fmt.Println("with DataContainer ID: " + id_data)
    }
  }
  startContainer(id_main, &optsServ)

  if verbose {
    fmt.Println(srv.Name + " Started.")
  }
}

func DockerRebuild(srv *util.Service, verbose bool) {
  var id string
  var wasRunning bool = false

  if verbose {
    fmt.Println("Rebuilding Service: " + util.FullNameToShort(srv.Name))
  }

  if service, exists := ContainerExists(srv); exists {
    if verbose {
      fmt.Println("Service exists, commensing with rebuild.")
    }

    if _, running := ContainerRunning(srv); running {
      if verbose {
        fmt.Println("Service is running. Stopping now.")
      }
      wasRunning = true
      DockerStop(srv, verbose)
    }

    if verbose {
      fmt.Println("Removing old container.")
    }
    removeContainer(service.ID)

  } else {

    if verbose {
      fmt.Println("Service did not previously exist. Nothing to rebuild.")
    }
    return
  }

  opts := configureContainer(srv)
  srv.Volumes = fixDirs(srv.Volumes)

  if verbose {
    fmt.Println("Recreating container.")
  }
  cont := createContainer(opts)
  id = cont.ID

  if wasRunning {
    if verbose {
      fmt.Println("Restarting Container with new ID: " + id)
    }
    startContainer(id, &opts)
  }

  if verbose {
    fmt.Println(util.FullNameToShort(srv.Name) + " Rebuilt.")
  }
}

func DockerLogs(srv *util.Service, verbose bool) {
  if verbose {
    fmt.Println("Getting service's logs.")
  }

  if service, exists := ContainerExists(srv); exists {
    if verbose {
      fmt.Println("Service ID: " + service.ID)
    }
    logsContainer(service.ID)
  } else {
    if verbose {
      fmt.Println("Service does not exist. Cannot display logs.")
    }
  }
}

func DockerStop(srv *util.Service, verbose bool) {
  // don't limit this to verbose because it takes a few seconds
  fmt.Println("Stopping: " + util.FullNameToShort(srv.Name) + ". This may take a few seconds.")

  var timeout uint = 10
  dockerAPIContainer, running := ContainerExists(srv)

  if running {
    stopContainer(dockerAPIContainer.ID, timeout)
  }

  if verbose {
    fmt.Println(util.FullNameToShort(srv.Name) + " Stopped.")
  }
}

func ContainerExists(srv *util.Service) (docker.APIContainers, bool) {
  return parseContainers(srv.Name, true)
}

func ContainerRunning(srv *util.Service) (docker.APIContainers, bool) {
  return parseContainers(srv.Name, false)
}

func parseContainers(name string, all bool) (docker.APIContainers, bool) {
  name = "/" + name
  containers := listContainers(all)
  if len(containers) != 0 {
    for _, container := range containers {
      if container.Names[0] == name {
        return container, true
      }
    }
  }
  return docker.APIContainers{}, false
}

func listContainers(all bool) ([]docker.APIContainers) {
  var container []docker.APIContainers
  r := regexp.MustCompile(`\/eris_(?:service|chain|data)_(.+)_\d`)

  contns, _ := util.DockerClient.ListContainers(docker.ListContainersOptions{All: all})
  for _, con := range contns {
    for _, c := range con.Names {
      match := r.FindAllStringSubmatch(c, 1)
      if len(match) != 0 {
        container = append(container, con)
      }
    }
  }

  return container
}

func createContainer(opts docker.CreateContainerOptions) *docker.Container {
  dockerContainer, err := util.DockerClient.CreateContainer(opts)
  if err != nil {
    // TODO: better error handling
    fmt.Println("Failed to create container ->\n  %v", err)
    os.Exit(1)
  }
  return dockerContainer
}

func startContainer(id string, opts *docker.CreateContainerOptions) {
  err := util.DockerClient.StartContainer(id, opts.HostConfig)
  if err != nil {
    // TODO: better error handling
    fmt.Println("Failed to start container ->\n  %v", err)
    os.Exit(1)
  }
}

func logsContainer(id string) {
  opts := docker.LogsOptions {
    Container:    id,
    Follow:       false,
    Stdout:       true,
    Stderr:       true,
    Timestamps:   false,
    OutputStream: os.Stdout,
    ErrorStream:  os.Stderr,
    // Tail:         "",
    // RawTerminal:  true, // Usually true when the container contains a TTY.
  }

  err := util.DockerClient.Logs(opts)
  if err != nil {
    // TODO: better error handling
    fmt.Println("Failed to get container logs ->\n  %v", err)
    os.Exit(1)
  }
}

func stopContainer(id string, timeout uint) {
  err := util.DockerClient.StopContainer(id, timeout)
  if err != nil {
    // TODO: better error handling
    fmt.Println("Failed to stop container ->\n  %v", err)
  }
}

func removeContainer(id string) {
  opts := docker.RemoveContainerOptions {
    ID:            id,
    RemoveVolumes: false,
    Force:         false,
  }
  err := util.DockerClient.RemoveContainer(opts)
  if err != nil {
    // TODO: better error handling
    fmt.Println("Failed to remove container ->\n  %v", err)
  }
}

func configureContainer(srv *util.Service) docker.CreateContainerOptions {
  opts := docker.CreateContainerOptions{
    Name: srv.Name,
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
      Labels:          srv.Labels,
      Cmd:             strings.Fields(srv.Command),
      Entrypoint:      strings.Fields(srv.EntryPoint),
      Image:           srv.Image,
      WorkingDir:      srv.WorkDir,
      NetworkDisabled: false,
    },
    HostConfig: &docker.HostConfig{
      Binds:           srv.Volumes,
      Links:           srv.Links,
      PublishAllPorts: false,
      Privileged:      srv.Privileged,
      ReadonlyRootfs:  false,
      DNS:             srv.DNS,
      DNSSearch:       srv.DNSSearch,
      VolumesFrom:     srv.VolumesFrom,
      CapAdd:          srv.CapAdd,
      CapDrop:         srv.CapDrop,
      RestartPolicy:   docker.NeverRestart(),
      NetworkMode:     "bridge",
    },
  }

  if srv.Attach {
    opts.Config.AttachStdin = true
    opts.Config.AttachStdout = true
    opts.Config.AttachStderr = true
    opts.Config.Tty = true
    opts.Config.OpenStdin = true
  }

  if srv.Restart == "always" {
    opts.HostConfig.RestartPolicy = docker.AlwaysRestart()
  } else if strings.Contains(srv.Restart, "max") {
    times, err := strconv.Atoi(strings.Split(srv.Restart, ":")[1])
    if err != nil {
      // TODO: better error handling
      fmt.Println(err)
      os.Exit(1)
    }
    opts.HostConfig.RestartPolicy = docker.RestartOnFailure(times)
  }

  opts.Config.ExposedPorts = make(map[docker.Port]struct{})
  opts.HostConfig.PortBindings = make(map[docker.Port][]docker.PortBinding)
  opts.Config.Volumes = make(map[string]struct{})

  for _, port := range srv.Ports {
    pS := strings.Split(port, ":")

    pR := pS[len(pS)-1]
    if len(strings.Split(pR, "/")) == 1 {
      pR = pR + "/tcp"
    }
    pC := docker.Port(fmt.Sprintf("%s", pR))

    if len(pS) > 1 {
      pH := docker.PortBinding {
        HostPort: pS[len(pS)-2],
      }

      if len(pS) == 3 {
        // ipv4
        pH.HostIP = pS[0]
      } else if len(pS) > 3 {
        // ipv6
        pH.HostIP = strings.Join(pS[:len(pS)-2], ":")
      }

      opts.Config.ExposedPorts[pC] = struct{}{}
      opts.HostConfig.PortBindings[pC] = []docker.PortBinding{pH}
    } else {
      opts.Config.ExposedPorts[pC] = struct{}{}
    }
  }

  for _, vol := range srv.Volumes {
    opts.Config.Volumes[strings.Split(vol, ":")[1]] = struct{}{}
  }

  return opts
}

func configureDataContainer(srv *util.Service, mainContOpts *docker.CreateContainerOptions) docker.CreateContainerOptions {
  opts := docker.CreateContainerOptions{
    Name: srv.DataContainerName,
    Config: &docker.Config{
      Image:           "eris/data",
      User:            srv.User,
      AttachStdin:     false,
      AttachStdout:    false,
      AttachStderr:    false,
      Tty:             false,
      OpenStdin:       false,
      NetworkDisabled: true,
    },
    HostConfig: &docker.HostConfig{
    },
  }

  mainContOpts.HostConfig.VolumesFrom = append(mainContOpts.HostConfig.VolumesFrom, srv.DataContainerName)

  return opts
}

// $(pwd) doesn't execute properly in golangs subshells; replace it
// use $eris as a shortcut
func fixDirs(arg []string) ([]string) {
  dir, err := os.Getwd()
  if err != nil {
    // TODO: error handling
    fmt.Println(err)
    os.Exit(1)
  }

  for n, a := range arg {
    if strings.Contains(a, "$eris") {
      tmp := strings.Split(a, ":")[0]
      keep := strings.Replace(a, tmp + ":", "", 1)
      if runtime.GOOS == "windows" {
        winTmp := strings.Split(tmp, "/")
        tmp = path.Join(winTmp...)
      }
      tmp = strings.Replace(tmp, "$eris", util.ErisRoot, 1)
      arg[n] = strings.Join([]string{tmp, keep}, ":")
      continue
    }

    if strings.Contains(a, "$pwd") {
      arg[n] = strings.Replace(a, "$pwd", dir, 1)
    }
  }

  return arg
}