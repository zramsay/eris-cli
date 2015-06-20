package definitions

type Action struct {
  // TODO: harmonize with actions_definition_spec.md
  Name        string            `json:"name" yaml:"name" toml:"name"`
  Services    []string          `json:"services" yaml:"services" toml:"services"`
  Chains      []string          `json:"chains" yaml:"chains" toml:"chains"`
  Steps       []string          `json:"steps" yaml:"steps" toml:"steps"`
  Environment map[string]string `json:"environment" yaml:"environment" toml:"environment"`

  // Used internally
  lastRan    string
}
