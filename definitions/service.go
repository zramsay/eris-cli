package definitions

type Service struct {
	// name of the service
	Name string `json:"name" yaml:"name" toml:"name"`
	// docker image used by the service
	Image string `json:"image,omitempty" yaml:"image,omitempty" toml:"image,omitempty"`
	// whether monax should automagically handle a data container for this service
	AutoData bool `json:"data_container" yaml:"data_container" toml:"data_container"`
	// restart policy: "always" or "max:<#attempts>"
	Restart string `json:",omitempty" yaml:",omitempty" toml:",omitempty"`
	// maps directly to docker cmd
	Command string `json:"command,omitempty" yaml:"command,omitempty" toml:"command,omitempty"`
	// maps directly to docker links
	Links []string `mapstructure:"links" json:"links,omitempty" yaml:"links,omitempty" toml:"links,omitempty"`
	// maps directly to docker ports
	Ports []string `mapstructure:"ports" json:"ports,omitempty" yaml:"ports,omitempty" toml:"ports,omitempty"`
	// maps directly do docker expose
	Expose []string `mapstructure:"expose" json:"expose,omitempty" yaml:"expose,omitempty" toml:"expose,omitempty"`
	// maps directly to docker volumes
	Volumes []string `mapstructure:"volumes" json:"volumes,omitempty" yaml:"volumes,omitempty" toml:"volumes,omitempty"`
	// maps directly to docker volumes-from
	VolumesFrom []string `mapstructure:"volumes_from" json:"volumes_from,omitempty" yaml:"volumes_from,omitempty" toml:"volumes_from,omitempty"`
	// maps directly to docker environment
	Environment []string `json:"environment,omitempty" yaml:"environment,omitempty" toml:"environment,omitempty"`
	// maps directly to docker env-file
	EnvFile []string `mapstructure:"env_file" json:"env_file,omitempty" yaml:"env_file,omitempty" toml:"env_file,omitempty"`
	// maps directly to docker net
	Net string `json:"net,omitempty" yaml:"net,omitempty" toml:"net,omitempty"`
	// maps directly to docker PID
	PID string `json:"pid,omitempty" yaml:"pid,omitempty" toml:"pid,omitempty"`
	// maps directly to docker DNS
	DNS []string `mapstructure:"dns" json:"dns,omitempty" yaml:"dns,omitempty" toml:"dns,omitempty"`
	// maps directly to docker DNS-search
	DNSSearch []string `mapstructure:"dns_search" json:"dns_search,omitempty" yaml:"dns_search,omitempty" toml:"dns_search,omitempty"`
	// maps directly to docker workdir
	WorkDir string `mapstructure:"work_dir" json:"work_dir,omitempty" yaml:"work_dir,omitempty" toml:"work_dir,omitempty"`
	// maps directly to docker entrypoint
	EntryPoint string `mapstructure:"entry_point" json:"entry_point,omitempty" yaml:"entry_point,omitempty" toml:"entry_point,omitempty"`
	// maps directly to docker hostname
	HostName string `mapstructure:"host_name" json:"host_name,omitempty" yaml:"host_name,omitempty" toml:"host_name,omitempty"`
	// maps directly to docker domainname
	DomainName string `mapstructure:"domain_name" json:"domain_name,omitempty" yaml:"domain_name,omitempty" toml:"domain_name,omitempty"`
	// maps directly to docker username
	User string `json:"user,omitempty" yaml:"user,omitempty" toml:"user,omitempty"`
	// maps directly to docker cpu_shares
	CPUShares int64 `mapstructure:"cpu_shares" json:"cpu_shares,omitempty,omitzero" yaml:"cpu_shares,omitempty" toml:"cpu_shares,omitempty,omitzero"`
	// maps directly to docker mem_limit
	MemLimit int64 `mapstructure:"mem_limit" json:"memory,omitempty,omitzero" yaml:"memory,omitempty" toml:"memory,omitempty,omitzero"`

	// an env variable to set for when we are running `monax exec` so we can find the main container
	ExecHost string `mapstructure:"exec_host" json:"exec_host,omitempty,omitzero" yaml:"exec_host,omitempty" toml:"exec_host,omitempty,omitzero"`
}

func BlankService() *Service {
	return &Service{}
}
