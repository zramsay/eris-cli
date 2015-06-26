package definitions

type ServiceDefinition struct {
	Service    *Service          `json:"service" yaml:"service" toml:"service"`
	Manager    Manager           `json:"manager,omitempty" yaml:"manager,omitempty" toml:"manager,omitempty"`
	Maintainer *Maintainer       `json:"maintainer,omitempty" yaml:"maintainer,omitempty" toml:"maintainer,omitempty"`
	Location   *Location         `json:"location,omitempty" yaml:"location,omitempty" toml:"location,omitempty"`
	Machine    *Machine          `json:"machine,omitempty" yaml:"machine,omitempty" toml:"machine,omitempty"`
	Operations *ServiceOperation `json:"operations" yaml:"operations" toml:"operations"`
}

// Service has the same structure used by docker-compose.yml. Complete and up
//   to date with the docker compose specification as of 04.06.15.
//   https://docs.docker.com/compose/yml
type Service struct {
	// TODO: harmonize with services_definition_spec.md
	Name        string            `json:"name,omitempty" yaml:"name,omitempty" toml:"name,omitempty"`
	Image       string            `json:"image,omitempty" yaml:"image,omitempty" toml:"image,omitempty"`
	AutoData    bool              `json:"data_container" yaml:"data_container" toml:"data_container`
	Command     string            `json:"command,omitempty" yaml:"command,omitempty" toml:"command,omitempty"`
	ServiceDeps []string          `mapstructure:"services" json:"services,omitempty", yaml:"services,omitempty" toml:"services,omitempty"`
	Labels      map[string]string `json:"labels,omitempty" yaml:"labels,omitempty" toml:"labels,omitempty"`
	Links       []string          `json:"links,omitempty" yaml:"links,omitempty" toml:"links,omitempty"`
	Ports       []string          `json:"ports,omitempty" yaml:"ports,omitempty" toml:"ports,omitempty"`
	Expose      []string          `json:"expose,omitempty" yaml:"expose,omitempty" toml:"expose,omitempty"`
	Volumes     []string          `json:"volumes,omitempty" yaml:"volumes,omitempty" toml:"volumes,omitempty"`
	VolumesFrom []string          `mapstructure:"volumes_from" json:"volumes_from,omitempty" yaml:"volumes_from,omitempty" toml:"volumes_from,omitempty"`
	Environment []string          `json:"environment,omitempty" yaml:"environment,omitempty" toml:"environment,omitempty"`
	EnvFile     []string          `mapstructure:"env_file" json:"env_file,omitempty" yaml:"env_file,omitempty" toml:"env_file,omitempty"`
	Net         string            `json:"net,omitempty" yaml:"net,omitempty" toml:"net,omitempty"`
	PID         string            `json:"pid,omitempty" yaml:"pid,omitempty" toml:"pid,omitempty"`
	CapAdd      []string          `mapstructure:"cap_add" json:"cap_add,omitempty" yaml:"cap_add,omitempty" toml:"cap_add,omitempty"`
	CapDrop     []string          `mapstructure:"cap_drop" json:"cap_drop,omitempty" yaml:"cap_drop,omitempty" toml:"cap_drop,omitempty"`
	DNS         []string          `json:"dns,omitempty" yaml:"dns,omitempty" toml:"dns,omitempty"`
	DNSSearch   []string          `mapstructure:"dns_search" json:"dns_search,omitempty" yaml:"dns_search,omitempty" toml:"dns_search,omitempty"`
	CPUShares   int64             `mapstructure:"cpu_shares" json:"cpu_shares,omitempty,omitzero" yaml:"cpu_shares,omitempty" toml:"cpu_shares,omitempty,omitzero"`
	WorkDir     string            `mapstructure:"work_dir" json:"work_dir,omitempty" yaml:"work_dir,omitempty" toml:"work_dir,omitempty"`
	EntryPoint  string            `mapstructure:"entry_point" json:"entry_point,omitempty" yaml:"entry_point,omitempty" toml:"entry_point,omitempty"`
	HostName    string            `mapstructure:"host_name" json:"host_name,omitempty" yaml:"host_name,omitempty" toml:"host_name,omitempty"`
	DomainName  string            `mapstructure:"domain_name" json:"domain_name,omitempty" yaml:"domain_name,omitempty" toml:"domain_name,omitempty"`
	User        string            `json:"user,omitempty" yaml:"user,omitempty" toml:"user,omitempty"`
	MemLimit    int64             `mapstructure:"mem_limit" json:"memory,omitempty,omitzero" yaml:"memory,omitempty" toml:"memory,omitempty,omitzero"`
}

type ServiceOperation struct {
	// Filled in dynamically prerun
	DataContainer     bool   `json:",omitempty" yaml:",omitempty" toml:",omitempty"`
	SrvContainerName  string `json:",omitempty" yaml:",omitempty" toml:",omitempty"`
	SrvContainerID    string `json:",omitempty" yaml:",omitempty" toml:",omitempty"`
	DataContainerName string `json:",omitempty" yaml:",omitempty" toml:",omitempty"`
	DataContainerID   string `json:",omitempty" yaml:",omitempty" toml:",omitempty"`
  ContainerNumber   int    `json:",omitempty,omitzero" yaml:",omitempty" toml:",omitempty,omitzero"`
	Restart           string `json:",omitempty" yaml:",omitempty" toml:",omitempty"`
	Remove            bool   `json:",omitempty" yaml:",omitempty" toml:",omitempty"`
	Privileged        bool   `json:",omitempty" yaml:",omitempty" toml:",omitempty"`
	Attach            bool   `json:",omitempty" yaml:",omitempty" toml:",omitempty"`
	AppName           string `json:",omitempty" yaml:",omitempty" toml:",omitempty"`
	DockerHostConn    string `json:",omitempty" yaml:",omitempty" toml:",omitempty"`
}

type Manager map[string]string
