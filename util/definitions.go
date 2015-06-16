package util

// Service has the same structure used by docker-compose.yml. Complete and up
//   to date with the docker compose specification as of 04.06.15.
//   https://docs.docker.com/compose/yml
type Service struct {
  // TODO: harmonize with services_definition_spec.md

  // Relatively static, stored in service definition files
  Name                 string            `json:"name" yaml:"name" toml:"name"`
  Image                string            `json:"image" yaml:"image" toml:"image"`
  Command              string            `json:"command" yaml:"command" toml:"command"`
  Labels               map[string]string `json:"labels" yaml:"labels" toml:"labels"`
  Links                []string          `json:"links" yaml:"links" toml:"links"`
  Ports                []string          `json:"ports" yaml:"ports" toml:"ports"`
  Expose               []string          `json:"expose" yaml:"expose" toml:"expose"`
  Volumes              []string          `json:"volumes" yaml:"volumes" toml:"volumes"`
  VolumesFrom          []string          `json:"volumes_from" yaml:"volumes_from" toml:"volumes_from"`
  Environment          []string          `json:"environment" yaml:"environment" toml:"environment"`
  EnvFile              []string          `json:"env_file" yaml:"env_file" toml:"env_file"`
  Net                  string            `json:"net" yaml:"net" toml:"net"`
  PID                  string            `json:"pid" yaml:"pid" toml:"pid"`
  CapAdd               []string          `json:"cap_add" yaml:"cap_add" toml:"cap_add"`
  CapDrop              []string          `json:"cap_drop" yaml:"cap_drop" toml:"cap_drop"`
  DNS                  []string          `json:"dns" yaml:"dns" toml:"dns"`
  DNSSearch            []string          `json:"dns_search" yaml:"dns_search" toml:"dns_search"`
  CPUShares            int64             `json:"cpu_shares" yaml:"cpu_shares" toml:"cpu_shares"`
  WorkDir              string            `json:"work_dir" yaml:"work_dir" toml:"work_dir"`
  EntryPoint           string            `json:"entry_point" yaml:"entry_point" toml:"entry_point"`
  HostName             string            `json:"host_name" yaml:"host_name" toml:"host_name"`
  DomainName           string            `json:"domain_name" yaml:"domain_name" toml:"domain_name"`
  User                 string            `json:"user" yaml:"user" toml:"user"`
  MemLimit             int64             `json:"memory" yaml:"memory" toml:"memory"`

  // Filled in dynamically prerun
  Restart          string
  Privileged       bool
  Attach           bool
  AppName          string
  DockerHostConn   string

  // Used internally
  lastUpdated    string
  containerID    string
}

type Chain struct {
  // TODO: harmonize with chains_definition_spec.md
  Name      string `json:"name" yaml:"name" toml:"name"`
  Type      string `json:"type" yaml:"type" toml:"type"`
  Directory string `json:"directory" yaml:"directory" toml:"directory"`
  Service   *Service
}

type Action struct {
  // TODO: harmonize with actions_definition_spec.md
  Name        string            `json:"name" yaml:"name" toml:"name"`
  Services    []string          `json:"services" yaml:"services" toml:"services"`
  Chains      []string          `json:"chains" yaml:"chains" toml:"chains"`
  Steps       []string          `json:"steps" yaml:"steps" toml:"steps"`
  Environment map[string]string `json:"environment" yaml:"environment" toml:"environment"`

  // Used internally
  lastRan    string
}