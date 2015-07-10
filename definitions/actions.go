package definitions

type Action struct {
	// TODO: harmonize with actions_definition_spec.md
	Name        string            `json:"name" yaml:"name" toml:"name"`
	ServiceDeps []string          `mapstructure:"services" json:"services" yaml:"services" toml:"services"`
	Chain       string            `json:"chain" yaml:"chain" toml:"chain"`
	Steps       []string          `mapstructure:"steps" json:"steps" yaml:"steps" toml:"steps"`
	Environment map[string]string `json:"environment" yaml:"environment" toml:"environment"`

	Maintainer  *Maintainer       `json:"maintainer,omitempty" yaml:"maintainer,omitempty" toml:"maintainer,omitempty"`
	Location    *Location         `json:"location,omitempty" yaml:"location,omitempty" toml:"location,omitempty"`
	Machine     *Machine          `json:"machine,omitempty" yaml:"machine,omitempty" toml:"machine,omitempty"`
	Srvs        []*Service
	Operations  *Operation
}

func BlankAction() *Action {
	return &Action{
		Maintainer: BlankMaintainer(),
		Location:   BlankLocation(),
		Machine:    BlankMachine(),
		Operations: BlankOperation(),
	}
}
