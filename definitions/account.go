package definitions

type ErisDBAccount struct {
	Name        string                    `mapstructure:"name" json:"name"`
	Address     string                    `mapstructure:"address" json:"address"`
	Amount      int                       `mapstructure:"amount" json:"amount"`
	Permissions *ErisDBAccountPermissions `mapstructure:"permissions" json:"permissions"`

	//ignored fields
	Validator         bool                      `json:"-"`
	PermissionsMap    map[string]int            `json:"-"`
	ErisDBPermissions *ErisDBAccountPermissions `json:"-"`
	MintKey           *MintPrivValidator        `json:"-"`
	PubKey            string                    `json:"-"`
	ToBond            int                       `json:"-"`
}
