package definitions

type ChainType struct {
	Name         string         `mapstructure:"name" json:"name" yaml:"name" toml:"name"`
	Description  string         `mapstructure:"description" json:"description" yaml:"description" toml:"description"`
	AccountTypes map[string]int `mapstructure:"account_types" json:"account_types" yaml:"account_types" toml:"account_types"`

	// currently unused
	ConsensusEngine    map[string]string `mapstructure:"tendermint" json:"tendermint" yaml:"tendermint" toml:"tendermint"`
	ApplicationManager map[string]string `mapstructure:"erismint" json:"erismint" yaml:"erismint" toml:"erismint"`
	Messenger          map[string]string `mapstructure:"servers" json:"servers" yaml:"servers" toml:"servers"`
}

func BlankChainType() *ChainType {
	return &ChainType{}
}
