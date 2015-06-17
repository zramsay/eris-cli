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
  if verbose {
    fmt.Println("Starting Service: " + srv.Name)
  }

  srv.Volumes = fixDirs(srv.Volumes)
  startContainer(srv)

  if verbose {
    fmt.Println(srv.Name + " Started")
  }
}

func DockerStop(srv *util.Service, verbose bool) {
  // don't limit this to verbose because it takes a few seconds
  fmt.Println("Stopping Service: " + srv.Name + ". This may take a few seconds.")

  var timeout uint = 10
  stopContainer(srv, timeout)

  if verbose {
    fmt.Println(srv.Name + " Stopped")
  }
}

func startContainer(srv *util.Service) {
  opts := configureContainer(srv)

  dockerContainer, err := util.DockerClient.CreateContainer(opts)
  if err != nil {
    // TODO: better error handling
    fmt.Println("Failed to create container ->\n  %v", err)
    os.Exit(1)
  }

  err = util.DockerClient.StartContainer(dockerContainer.ID, opts.HostConfig)
  if err != nil {
    // TODO: better error handling
    fmt.Println("Failed to start container ->\n  %v", err)
    os.Exit(1)
  }
}

func stopContainer(srv *util.Service, timeout uint) {
  dockerAPIContainer := srvToContainer(srv)

  err := util.DockerClient.StopContainer(dockerAPIContainer.ID, timeout)
  if err != nil {
    // TODO: better error handling
    fmt.Println("Failed to stop container ->\n  %v", err)
  }
}

func srvToContainer(srv *util.Service) *docker.APIContainers {
  var container *docker.APIContainers
  r := regexp.MustCompile(`\/eris_(?:service|chain)_(.+)_\S+?_\d`)

  contns, _ := util.DockerClient.ListContainers(docker.ListContainersOptions{All: false})
  for _, con := range contns {
    for _, c := range con.Names {
      match := r.FindAllStringSubmatch(c, 1)
      if len(match) != 0 {
        container = &con
      }
    }
  }

  return container
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
      fmt.Println(arg[n])
      continue
    }

    if strings.Contains(a, "$pwd") {
      arg[n] = strings.Replace(a, "$pwd", dir, 1)
    }
  }

  return arg
}