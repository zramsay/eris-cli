package definitions

type ErisDBAccount struct {
	Name    string `mapstructure:"," json:"," yaml:"," toml:","`
	Address string `mapstructure:"," json:"," yaml:"," toml:","`
	PubKey  string `mapstructure:"," json:"," yaml:"," toml:","`
	Tokens  int    `mapstructure:"," json:"," yaml:"," toml:","`
	ToBond  int    `mapstructure:"," json:"," yaml:"," toml:","`

	// [zr] from definitions/erisdb_chains.go
	//Address     string                    `json:"address"`
	Amount int `json:"amount"`
	//Name        string                    `json:"name"`
	Permissions *ErisDBAccountPermissions `json:"permissions"`

	Validator         bool
	PermissionsMap    map[string]int
	ErisDBPermissions *ErisDBAccountPermissions
	MintKey           *MintPrivValidator
}
