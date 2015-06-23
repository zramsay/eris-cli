package definitions

type Chain struct {
	// TODO: harmonize with chains_definition_spec.md
	Name       string            `json:"name" yaml:"name" toml:"name"`
	Type       string            `json:"type" yaml:"type" toml:"type"`
	Manage     *ChainManager     `json:"manage,omitempty" yaml:manage,omitempty" toml:"manage,omitempty"`
	Service    *Service          `json:"service,omitempty" yaml:"service,omitempty" toml:"service,omitempty"`
	Maintainer *Maintainer       `json:"maintainer,omitempty" yaml:"maintainer,omitempty" toml:"maintainer,omitempty"`
	Location   *Location         `json:"location,omitempty" yaml:"location,omitempty" toml:"location,omitempty"`
	Operations *ServiceOperation `json:"operations" yaml:"operations" toml:"operations"`
}

type ChainManager struct {
	WorkingDir string `json:"working_dir" yaml:"working_dir" toml:"working_dir"`
	FetchCmd   string `json:"fetch" yaml:"fetch" toml:"fetch"`
	RunCmd     string `json:"run" yaml:"run" toml:"run"`
	NewCmd     string `json:"new" yaml:"new" toml:"new"`
}
