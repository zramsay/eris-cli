package definitions

type ServiceDefinition struct {
	// TODO: harmonize with services_definition_spec.md
	Name        string   `json:"name" yaml:"name" toml:"name"`
	ServiceID   string   `mapstructure:"service_id" json:"service_id" yaml:"service_id" toml:"service_id"`
	ServiceDeps []string `mapstructure:"services" json:"services,omitempty", yaml:"services,omitempty" toml:"services,omitempty"`
	Chain       string   `json:"chain" yaml:"chain" toml:"chain"`

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
