package definitions

type Chain struct {
  // TODO: harmonize with chains_definition_spec.md
  Name      string `json:"name" yaml:"name" toml:"name"`
  Type      string `json:"type" yaml:"type" toml:"type"`
  Directory string `json:"directory" yaml:"directory" toml:"directory"`
  Service   *Service
}
