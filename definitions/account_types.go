package definitions

type MonaxDBAccountType struct {
	Name          string          `mapstructure:"name" json:"name" yaml:"name" toml:"name"`
	Description   string          `mapstructure:"description" json:"description" yaml:"description" toml:"description"`
	TypicalUser   string          `mapstructure:"typical_user" json:"typical_user" yaml:"typical_user" toml:"typical_user"`
	DefaultNumber int64           `mapstructure:"default_number" json:"default_number" yaml:"default_number" toml:"default_number"`
	DefaultTokens int64           `mapstructure:"default_tokens" json:"default_tokens" yaml:"default_tokens" toml:"default_tokens"`
	DefaultBond   int64           `mapstructure:"default_bond" json:"default_bond" yaml:"default_bond" toml:"default_bond"`
	Perms         map[string]bool `mapstructure:"perms" json:"perms" yaml:"perms" toml:"perms"`
}

func BlankAccountType() *MonaxDBAccountType {
	return &MonaxDBAccountType{}
}
