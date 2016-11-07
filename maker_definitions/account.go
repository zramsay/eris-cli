package definitions

type Account struct {
	Name    string `mapstructure:"," json:"," yaml:"," toml:","`
	Address string `mapstructure:"," json:"," yaml:"," toml:","`
	PubKey  string `mapstructure:"," json:"," yaml:"," toml:","`
	Tokens  int    `mapstructure:"," json:"," yaml:"," toml:","`
	ToBond  int    `mapstructure:"," json:"," yaml:"," toml:","`

	Validator       bool
	PermissionsMap  map[string]int
	MintPermissions *MintAccountPermissions
	MintKey         *MintPrivValidator
}
