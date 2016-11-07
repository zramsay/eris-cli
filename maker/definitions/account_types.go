package definitions

type AccountType struct {
	Name        string         `mapstructure:"name" json:"name" yaml:"name" toml:"name"`
	Definition  string         `mapstructure:"definition" json:"definition" yaml:"definition" toml:"definition"`
	TypicalUser string         `mapstructure:"typical_user" json:"typical_user" yaml:"typical_user" toml:"typical_user"`
	Number      int            `mapstructure:"default_number" json:"default_number" yaml:"default_number" toml:"default_number"`
	Tokens      int            `mapstructure:"default_tokens" json:"default_tokens" yaml:"default_tokens" toml:"default_tokens"`
	ToBond      int            `mapstructure:"default_bond" json:"default_bond" yaml:"default_bond" toml:"default_bond"`
	Perms       map[string]int `mapstructure:"perms" json:"perms" yaml:"perms" toml:"perms"`
}

func BlankAccountType() *AccountType {
	return &AccountType{}
}
