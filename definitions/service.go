package definitions

// Service has the same structure used by docker-compose.yml. Complete and up
// to date with the docker compose specification as of 04.06.15.
// https://docs.docker.com/compose/yml
//
// Services and Service Definitions have been separated because the
type Service struct {
	Name        string   `json:"name" yaml:"name" toml:"name"`
	Image       string   `json:"image,omitempty" yaml:"image,omitempty" toml:"image,omitempty"`
	AutoData    bool     `json:"data_container" yaml:"data_container" toml:"data_container"`
	Command     string   `json:"command,omitempty" yaml:"command,omitempty" toml:"command,omitempty"`
	Links       []string `mapstructure:"links" json:"links,omitempty" yaml:"links,omitempty" toml:"links,omitempty"`
	Ports       []string `mapstructure:"ports" json:"ports,omitempty" yaml:"ports,omitempty" toml:"ports,omitempty"`
	Expose      []string `mapstructure:"expose" json:"expose,omitempty" yaml:"expose,omitempty" toml:"expose,omitempty"`
	Volumes     []string `mapstructure:"volumes" json:"volumes,omitempty" yaml:"volumes,omitempty" toml:"volumes,omitempty"`
	VolumesFrom []string `mapstructure:"volumes_from" json:"volumes_from,omitempty" yaml:"volumes_from,omitempty" toml:"volumes_from,omitempty"`
	Environment []string `json:"environment,omitempty" yaml:"environment,omitempty" toml:"environment,omitempty"`
	EnvFile     []string `mapstructure:"env_file" json:"env_file,omitempty" yaml:"env_file,omitempty" toml:"env_file,omitempty"`
	Net         string   `json:"net,omitempty" yaml:"net,omitempty" toml:"net,omitempty"`
	PID         string   `json:"pid,omitempty" yaml:"pid,omitempty" toml:"pid,omitempty"`
	DNS         []string `mapstructure:"dns" json:"dns,omitempty" yaml:"dns,omitempty" toml:"dns,omitempty"`
	DNSSearch   []string `mapstructure:"dns_search" json:"dns_search,omitempty" yaml:"dns_search,omitempty" toml:"dns_search,omitempty"`
	WorkDir     string   `mapstructure:"work_dir" json:"work_dir,omitempty" yaml:"work_dir,omitempty" toml:"work_dir,omitempty"`
	EntryPoint  string   `mapstructure:"entry_point" json:"entry_point,omitempty" yaml:"entry_point,omitempty" toml:"entry_point,omitempty"`
	HostName    string   `mapstructure:"host_name" json:"host_name,omitempty" yaml:"host_name,omitempty" toml:"host_name,omitempty"`
	DomainName  string   `mapstructure:"domain_name" json:"domain_name,omitempty" yaml:"domain_name,omitempty" toml:"domain_name,omitempty"`
	User        string   `json:"user,omitempty" yaml:"user,omitempty" toml:"user,omitempty"`
	CPUShares   int64    `mapstructure:"cpu_shares" json:"cpu_shares,omitempty,omitzero" yaml:"cpu_shares,omitempty" toml:"cpu_shares,omitempty,omitzero"`
	MemLimit    int64    `mapstructure:"mem_limit" json:"memory,omitempty,omitzero" yaml:"memory,omitempty" toml:"memory,omitempty,omitzero"`
}

func BlankService() *Service {
	return &Service{}
}
