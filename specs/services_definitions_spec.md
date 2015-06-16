# Action Definition Files Specification

// Service has the same structure used by docker-compose.yml. Complete and up
//   to date with the docker compose specification as of 04.06.15.
//   https://docs.docker.com/compose/yml
type Service struct {
  // Relatively static, stored in service definition files
  Name                 string
  Image                string
  Build                string
  Command              string
  Links                []string
  Ports                []string
  Expose               []string
  Volumes              []string
  VolumesFrom          []string
  Environment          map[string]string
  EnvFile              []string
  Net                  string
  PID                  string
  DNS                  []string
  CapAdd               []string
  CapDrop              []string
  DNSSearch            []string
  CPUShares            string
  WorkingDir           string
  EntryPoint           string
  HostName             string
  User                 string
  MemLimit             string
  Restart              string

  // Filled in dynamically prerun
  Privileged           bool
  Attach               bool
  AppName              string
  DockerHostConnCmdArg string

  // Used internally
}

all path element should use POSIX file formatting. Underneath the hood, eris will transpose those to windows path formats if eris is being run on windows.
