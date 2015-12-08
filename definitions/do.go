package definitions

type Do struct {
	AddDir        bool     `mapstructure:"," json:"," yaml:"," toml:","`
	Actions       bool     `mapstructure:"," json:"," yaml:"," toml:","`
	Force         bool     `mapstructure:"," json:"," yaml:"," toml:","`
	File          bool     `mapstructure:"," json:"," yaml:"," toml:","`
	Pull          bool     `mapstructure:"," json:"," yaml:"," toml:","`
	Quiet         bool     `mapstructure:"," json:"," yaml:"," toml:","`
	All           bool     `mapstructure:"," json:"," yaml:"," toml:","`
	Follow        bool     `mapstructure:"," json:"," yaml:"," toml:","`
	Run           bool     `mapstructure:"," json:"," yaml:"," toml:","`
	Rm            bool     `mapstructure:"," json:"," yaml:"," toml:","`
	RmD           bool     `mapstructure:"," json:"," yaml:"," toml:","`
	RmHF          bool     `mapstructure:"," json:"," yaml:"," toml:","`
	Services      bool     `mapstructure:"," json:"," yaml:"," toml:","`
	Tool          bool     `mapstructure:"," json:"," yaml:"," toml:","`
	Verbose       bool     `mapstructure:"," json:"," yaml:"," toml:","`
	Debug         bool     `mapstructure:"," json:"," yaml:"," toml:","`
	Yes           bool     `mapstructure:"," json:"," yaml:"," toml:","`
	OutputTable   bool     `mapstructure:"," json:"," yaml:"," toml:","`
	Dump          bool     `mapstructure:"," json:"," yaml:"," toml:","`
	Lines         int      `mapstructure:"," json:"," yaml:"," toml:","` // XXX: for tail and logs
	Timeout       uint     `mapstructure:"," json:"," yaml:"," toml:","`
	N             uint     `mapstructure:"," json:"," yaml:"," toml:","`
	Address       string   `mapstructure:"," json:"," yaml:"," toml:","`
	Pubkey        string   `mapstructure:"," json:"," yaml:"," toml:","`
	Type          string   `mapstructure:"," json:"," yaml:"," toml:","`
	Task          string   `mapstructure:"," json:"," yaml:"," toml:","`
	Tail          string   `mapstructure:"," json:"," yaml:"," toml:","`
	Branch        string   `mapstructure:"," json:"," yaml:"," toml:","`
	ChainName     string   `mapstructure:"," json:"," yaml:"," toml:","`
	GenesisFile   string   `mapstructure:"," json:"," yaml:"," toml:","`
	ConfigFile    string   `mapstructure:"," json:"," yaml:"," toml:","`
	ServerConf    string   `mapstructure:"," json:"," yaml:"," toml:","`
	ChainID       string   `mapstructure:"," json:"," yaml:"," toml:","`
	Hash          string   `mapstructure:"," json:"," yaml:"," toml:","`
	Gateway       string   `mapstructure:"," json:"," yaml:"," toml:","`
	MachineName   string   `mapstructure:"," json:"," yaml:"," toml:","`
	Name          string   `mapstructure:"," json:"," yaml:"," toml:","`
	Image         string   `mapstructure:"," json:"," yaml:"," toml:","`
	Path          string   `mapstructure:"," json:"," yaml:"," toml:","`
	CSV           string   `mapstructure:"," json:"," yaml:"," toml:","`
	NewName       string   `mapstructure:"," json:"," yaml:"," toml:","`
	ResultFormt   string   `mapstructure:"," json:"," yaml:"," toml:","`
	Priv          string   `mapstructure:"," json:"," yaml:"," toml:","`
	Volume        string   `mapstructure:"," json:"," yaml:"," toml:","`
	EPMConfigFile string   `mapstructure:"," json:"," yaml:"," toml:","`
	ContractsPath string   `mapstructure:"," json:"," yaml:"," toml:","`
	ABIPath       string   `mapstructure:"," json:"," yaml:"," toml:","`
	DefaultGas    string   `mapstructure:"," json:"," yaml:"," toml:","`
	Compiler      string   `mapstructure:"," json:"," yaml:"," toml:","`
	DefaultAddr   string   `mapstructure:"," json:"," yaml:"," toml:","`
	DefaultFee    string   `mapstructure:"," json:"," yaml:"," toml:","`
	DefaultAmount string   `mapstructure:"," json:"," yaml:"," toml:","`
	ServicesSlice []string `mapstructure:"," json:"," yaml:"," toml:","`
	ConfigOpts    []string `mapstructure:"," json:"," yaml:"," toml:","`
	//clean
	Images    bool `mapstructure:"," json:"," yaml:"," toml:","`
	Uninstall bool `mapstructure:"," json:"," yaml:"," toml:","`
	Volumes   bool `mapstructure:"," json:"," yaml:"," toml:","`
	//data import/export
	Source      string `mapstructure:"," json:"," yaml:"," toml:","`
	Destination string `mapstructure:"," json:"," yaml:"," toml:","`

	//listing functions
	Known    bool `mapstructure:"," json:"," yaml:"," toml:","`
	Running  bool `mapstructure:"," json:"," yaml:"," toml:","`
	Existing bool `mapstructure:"," json:"," yaml:"," toml:","`

	// <key>=<value> pairs
	Env []string `mapstructure:"," json:"," yaml:"," toml:","`

	// <container_name>:<internal_name> pairs
	Links []string `mapstructure:"," json:"," yaml:"," toml:","`

	// Objects
	Action            *Action
	Chain             *Chain
	Operations        *Operation
	Service           *Service
	ServiceDefinition *ServiceDefinition

	// Return
	Result string
}

func NowDo() *Do {
	return &Do{
		Action:            BlankAction(),
		Chain:             BlankChain(),
		Operations:        BlankOperation(),
		Service:           BlankService(),
		ServiceDefinition: BlankServiceDefinition(),
	}
}
