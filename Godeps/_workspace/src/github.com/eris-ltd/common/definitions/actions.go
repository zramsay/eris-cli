package definitions

type Action struct {
  // TODO: harmonize with actions_definition_spec.md
  Name        string            `json:"name" yaml:"name" toml:"name"`
  Services    []string          `json:"services" yaml:"services" toml:"services"`
  Chains      []string          `json:"chains,omitempty" yaml:"chains,omitempty" toml:"chains,omitempty"`
  Steps       []string          `json:"steps" yaml:"steps" toml:"steps"`
  Environment map[string]string `json:"environment,omitempty" yaml:"environment,omitempty" toml:"environment,omitempty"`

  // Used internally
  lastRan    string
}
