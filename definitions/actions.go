package definitions

type Action struct {
	// TODO: harmonize with actions_definition_spec.md
	Name        string            `json:"name" yaml:"name" toml:"name"`
	Services    []string          `json:"services" yaml:"services" toml:"services"`
	Chains      []string          `json:"chains,omitempty" yaml:"chains,omitempty" toml:"chains,omitempty"`
	Steps       []string          `json:"steps" yaml:"steps" toml:"steps"`
	Environment map[string]string `json:"environment,omitempty" yaml:"environment,omitempty" toml:"environment,omitempty"`

	Maintainer *Maintainer       `json:"maintainer,omitempty" yaml:"maintainer,omitempty" toml:"maintainer,omitempty"`
	Location   *Location         `json:"location,omitempty" yaml:"location,omitempty" toml:"location,omitempty"`
	Machine    *Machine          `json:"machine,omitempty" yaml:"machine,omitempty" toml:"machine,omitempty"`
	Operations *ServiceOperation `json:"operations" yaml:"operations" toml:"operations"`
}
