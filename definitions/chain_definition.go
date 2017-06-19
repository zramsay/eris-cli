package definitions

type ChainDefinition struct {
	// name of the chain
	Name string `json:"name" yaml:"name" toml:"name"`
	// chain_id of the chain is in definitions/chain/.go

	// type of the chain
	ChainType string `mapstructure:"chain_type" json:"chain_type" yaml:"chain_type" toml:"chain_type"`

	// same fields as in the Service Struct/Service Specification
	Service      *Service      `json:"service,omitempty" yaml:"service,omitempty" toml:"service,omitempty"`
	Chain        *Chain        `json:"chain,omitempty" yaml:"chain,omitempty" toml:"chain,omitempty"`
	Dependencies *Dependencies `json:"dependencies,omitempty" yaml:"dependencies,omitempty" toml:"dependencies,omitempty"`
	Maintainer   *Maintainer   `json:"maintainer,omitempty" yaml:"maintainer,omitempty" toml:"maintainer,omitempty"`
	Location     *Location     `json:"location,omitempty" yaml:"location,omitempty" toml:"location,omitempty"`
	Operations   *Operation
}

func BlankChainDefinition() *ChainDefinition {
	return &ChainDefinition{
		Service:      BlankService(),
		Chain:        BlankChain(),
		Dependencies: BlankDependencies(),
		Maintainer:   BlankMaintainer(),
		Location:     BlankLocation(),
		Operations:   BlankOperation(),
	}
}
