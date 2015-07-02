package definitions

type Machine struct {
	Include  []string `json:"include,omitempty" yaml:"include,omitempty" toml:"include,omitempty"`
	Requires []string `json:"requires,omitempty" yaml:"requires,omitempty" toml:"requires,omitempty"`
}
