package definitions

type ServiceDefinition struct {
	// name of the service
	Name string `json:"name" yaml:"name" toml:"name"`
	// id of the service
	ServiceID string `mapstructure:"service_id" json:"service_id" yaml:"service_id" toml:"service_id"`
	// array of strings of other services which should be started prior to this service starting
	ServiceDeps []string `mapstructure:"services" json:"services,omitempty", yaml:"services,omitempty" toml:"services,omitempty"`
	// a chain which must be started prior to this service starting. can take a `$chain` string
	// which would then be passed in via a command line flag
	Chain string `json:"chain" yaml:"chain" toml:"chain"`

	Service    *Service    `json:"service" yaml:"service" toml:"service"`
	Maintainer *Maintainer `json:"maintainer,omitempty" yaml:"maintainer,omitempty" toml:"maintainer,omitempty"`
	Location   *Location   `json:"location,omitempty" yaml:"location,omitempty" toml:"location,omitempty"`
	Machine    *Machine    `json:"machine,omitempty" yaml:"machine,omitempty" toml:"machine,omitempty"`
	Srvs       []*Service
	Operations *Operation
}

func BlankServiceDefinition() *ServiceDefinition {
	return &ServiceDefinition{
		Service:    BlankService(),
		Maintainer: BlankMaintainer(),
		Location:   BlankLocation(),
		Machine:    BlankMachine(),
		Operations: BlankOperation(),
	}
}
