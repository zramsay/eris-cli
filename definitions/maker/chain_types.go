package definitions

type ChainType struct {
	Name         string         `mapstructure:"name" json:"name" yaml:"name" toml:"name"`
	AccountTypes map[string]int `mapstructure:"account_types" json:"account_types" yaml:"account_types" toml:"account_types"`

	// currently unused
	ConsensusEngine    map[string]string `mapstructure:"consensus" json:"consensus" yaml:"consensus" toml:"consensus"`
	ApplicationManager map[string]string `mapstructure:"manager" json:"manager" yaml:"manager" toml:"manager"`
	Messenger          map[string]string `mapstructure:"messenger" json:"messenger" yaml:"messenger" toml:"messenger"`
}

func BlankChainType() *ChainType {
	return &ChainType{}
}
