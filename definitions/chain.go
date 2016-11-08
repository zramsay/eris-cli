package definitions

// fields written to the config.toml
// can probably be trimmed to only chain_id?
type Chain struct {
	ChainID string `mapstructure:"assert_chain_id" json:"assert_chain_id" yaml:"assert_chain_id" toml:"assert_chain_id"`
	Name    string
	Type    string
	Genesis *interface{}
}

func BlankChain() *Chain {
	return &Chain{}
}
