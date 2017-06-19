package definitions

type ServiceDefinition struct {
	// name of the service
	Name string `json:"name" yaml:"name" toml:"name"`
	// id of the service
	ServiceID string `mapstructure:"service_id,omitempty" json:"service_id,omitempty" yaml:"service_id,omitempty" toml:"service_id,omitempty"`

	Description string `json:"description,omitempty" yaml:"description,omitempty" toml:"description,omitempty"`
	Status      string `json:"status,omitempty" yaml:"status,omitempty" toml:"status,omitempty"`

	// a chain which must be started prior to this service starting. can take a `$chain` string
	// which would then be passed in via a command line flag
	Chain string `json:"chain,omitempty" yaml:"chain,omitempty" toml:"chain,omitempty"`

	Service      *Service      `json:"service" yaml:"service" toml:"service"`
	Dependencies *Dependencies `json:"dependencies,omitempty" yaml:"dependencies,omitempty" toml:"dependencies,omitempty"`
	Maintainer   *Maintainer   `json:"maintainer,omitempty" yaml:"maintainer,omitempty" toml:"maintainer,omitempty"`
	Location     *Location     `json:"location,omitempty" yaml:"location,omitempty" toml:"location,omitempty"`
	Srvs         []*Service
	Operations   *Operation
}

type Dependencies struct {
	Chains   []string `json:"chains,omitempty" yaml:"chains,omitempty" toml:"chains,omitempty"`
	Services []string `json:"services,omitempty" yaml:"services,omitempty" toml:"services,omitempty"`
}

func BlankServiceDefinition() *ServiceDefinition {
	return &ServiceDefinition{
		Service:    BlankService(),
		Maintainer: BlankMaintainer(),
		Location:   BlankLocation(),
		Operations: BlankOperation(),
	}
}

func BlankDependencies() *Dependencies {
	return &Dependencies{}
}
