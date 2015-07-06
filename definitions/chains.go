package definitions

type Chain struct {
	// TODO: harmonize with chains_definition_spec.md
	Name    string `json:"name" yaml:"name" toml:"name"`
	ChainID string `mapstructure:"chain_id" json:"chain_id" yaml:"chain_id" toml:"chain_id"`

	Service    *Service    `json:"service,omitempty" yaml:"service,omitempty" toml:"service,omitempty"`
	Maintainer *Maintainer `json:"maintainer,omitempty" yaml:"maintainer,omitempty" toml:"maintainer,omitempty"`
	Location   *Location   `json:"location,omitempty" yaml:"location,omitempty" toml:"location,omitempty"`
	Machine    *Machine    `json:"machine,omitempty" yaml:"machine,omitempty" toml:"machine,omitempty"`
	Operations *Operation
}

func BlankChain() *Chain {
	return &Chain{
		Service:    BlankService(),
		Maintainer: BlankMaintainer(),
		Location:   BlankLocation(),
		Machine:    BlankMachine(),
		Operations: BlankOperation(),
	}
}
