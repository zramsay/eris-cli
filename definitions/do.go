package definitions

type Do struct {
	AddDir        bool     `mapstructure:"," json:"," yaml:"," toml:","`
	Force         bool     `mapstructure:"," json:"," yaml:"," toml:","`
	File          bool     `mapstructure:"," json:"," yaml:"," toml:","`
	Pull          bool     `mapstructure:"," json:"," yaml:"," toml:","`
	Quiet         bool     `mapstructure:"," json:"," yaml:"," toml:","`
	JSON          bool     `mapstructure:"," json:"," yaml:"," toml:","`
	All           bool     `mapstructure:"," json:"," yaml:"," toml:","`
	Follow        bool     `mapstructure:"," json:"," yaml:"," toml:","`
	Logrotate     bool     `mapstructure:"," json:"," yaml:"," toml:","`
	Rm            bool     `mapstructure:"," json:"," yaml:"," toml:","`
	RmImage       bool     `mapstructure:"," json:"," yaml:"," toml:","`
	RmD           bool     `mapstructure:"," json:"," yaml:"," toml:","`
	RmHF          bool     `mapstructure:"," json:"," yaml:"," toml:","`
	Services      bool     `mapstructure:"," json:"," yaml:"," toml:","`
	Tool          bool     `mapstructure:"," json:"," yaml:"," toml:","`
	Verbose       bool     `mapstructure:"," json:"," yaml:"," toml:","`
	Debug         bool     `mapstructure:"," json:"," yaml:"," toml:","`
	Yes           bool     `mapstructure:"," json:"," yaml:"," toml:","`
	Tarball       bool     `mapstructure:"," json:"," yaml:"," toml:","`
	ZipFile       bool     `mapstructure:"," json:"," yaml:"," toml:","`
	OutputTable   bool     `mapstructure:"," json:"," yaml:"," toml:","`
	Overwrite     bool     `mapstructure:"," json:"," yaml:"," toml:","`
	Dump          bool     `mapstructure:"," json:"," yaml:"," toml:","`
	Save          bool     `mapstructure:"," json:"," yaml:"," toml:","`
	Unsafe        bool     `mapstructure:"," json:"," yaml:"," toml:","`
	Wizard        bool     `mapstructure:"," json:"," yaml:"," toml:","`
	Lines         int      `mapstructure:"," json:"," yaml:"," toml:","`
	Timeout       uint     `mapstructure:"," json:"," yaml:"," toml:","`
	N             uint     `mapstructure:"," json:"," yaml:"," toml:","`
	Address       string   `mapstructure:"," json:"," yaml:"," toml:","`
	Pubkey        string   `mapstructure:"," json:"," yaml:"," toml:","`
	Type          string   `mapstructure:"," json:"," yaml:"," toml:","`
	Task          string   `mapstructure:"," json:"," yaml:"," toml:","`
	Tail          string   `mapstructure:"," json:"," yaml:"," toml:","`
	ChainName     string   `mapstructure:"," json:"," yaml:"," toml:","`
	ChainType     string   `mapstructure:"," json:"," yaml:"," toml:","`
	GenesisFile   string   `mapstructure:"," json:"," yaml:"," toml:","`
	Hash          string   `mapstructure:"," json:"," yaml:"," toml:","`
	Name          string   `mapstructure:"," json:"," yaml:"," toml:","`
	Image         string   `mapstructure:"," json:"," yaml:"," toml:","`
	Path          string   `mapstructure:"," json:"," yaml:"," toml:","`
	CSV           string   `mapstructure:"," json:"," yaml:"," toml:","`
	NewName       string   `mapstructure:"," json:"," yaml:"," toml:","`
	Format        string   `mapstructure:"," json:"," yaml:"," toml:","`
	Priv          string   `mapstructure:"," json:"," yaml:"," toml:","`
	Volume        string   `mapstructure:"," json:"," yaml:"," toml:","`
	EPMConfigFile string   `mapstructure:"," json:"," yaml:"," toml:","`
	KeysPort      string   `mapstructure:"," json:"," yaml:"," toml:","`
	ChainPort     string   `mapstructure:"," json:"," yaml:"," toml:","`
	PackagePath   string   `mapstructure:"," json:"," yaml:"," toml:","`
	ABIPath       string   `mapstructure:"," json:"," yaml:"," toml:","`
	BinPath       string   `mapstructure:"," json:"," yaml:"," toml:","`
	DefaultGas    string   `mapstructure:"," json:"," yaml:"," toml:","`
	Compiler      string   `mapstructure:"," json:"," yaml:"," toml:","`
	DefaultAddr   string   `mapstructure:"," json:"," yaml:"," toml:","`
	DefaultFee    string   `mapstructure:"," json:"," yaml:"," toml:","`
	DefaultAmount string   `mapstructure:"," json:"," yaml:"," toml:","`
	ChainMakeActs string   `mapstructure:"," json:"," yaml:"," toml:","`
	ChainMakeVals string   `mapstructure:"," json:"," yaml:"," toml:","`
	Output        string   `mapstructure:"," json:"," yaml:"," toml:","`
	SeedsIP       []string `mapstructure:"," json:"," yaml:"," toml:","`
	ServicesSlice []string `mapstructure:"," json:"," yaml:"," toml:","`
	ImagesSlice   []string `mapstructure:"," json:"," yaml:"," toml:","`
	ConfigOpts    []string `mapstructure:"," json:"," yaml:"," toml:","`
	AccountTypes  []string `mapstructure:"," json:"," yaml:"," toml:","`

	// for [monax pkgs do]
	YAMLPath      string   `mapstructure:"," json:"," yaml:"," toml:","`
	ContractsPath string   `mapstructure:"," json:"," yaml:"," toml:","`
	Signer        string   `mapstructure:"," json:"," yaml:"," toml:","`
	PublicKey     string   `mapstructure:"," json:"," yaml:"," toml:","`
	ChainID       string   `mapstructure:"," json:"," yaml:"," toml:","`
	ChainIP       string   `mapstructure:"," json:"," yaml:"," toml:","`
	ChainURL      string   `mapstructure:"," json:"," yaml:"," toml:","`
	DefaultOutput string   `mapstructure:"," json:"," yaml:"," toml:","`
	DefaultSets   []string `mapstructure:"," json:"," yaml:"," toml:","`
	Package       *Package

	//clean
	Containers bool `mapstructure:"," json:"," yaml:"," toml:","`
	ChnDirs    bool `mapstructure:"," json:"," yaml:"," toml:","`
	Scratch    bool `mapstructure:"," json:"," yaml:"," toml:","`
	Images     bool `mapstructure:"," json:"," yaml:"," toml:","`
	Uninstall  bool `mapstructure:"," json:"," yaml:"," toml:","`
	Volumes    bool `mapstructure:"," json:"," yaml:"," toml:","`

	//data import/export
	Source      string `mapstructure:"," json:"," yaml:"," toml:","`
	Destination string `mapstructure:"," json:"," yaml:"," toml:","`

	//listing functions
	Known     bool `mapstructure:"," json:"," yaml:"," toml:","`
	Running   bool `mapstructure:"," json:"," yaml:"," toml:","`
	Existing  bool `mapstructure:"," json:"," yaml:"," toml:","`
	Host      bool `mapstructure:"," json:"," yaml:"," toml:","` //keys ls
	Container bool `mapstructure:"," json:"," yaml:"," toml:","` //keys ls

	// <key>=<value> pairs
	Env []string `mapstructure:"," json:"," yaml:"," toml:","`

	// <container_name>:<internal_name> pairs
	Links []string `mapstructure:"," json:"," yaml:"," toml:","`

	// Objects
	ChainDefinition   *ChainDefinition
	Operations        *Operation
	Service           *Service
	ServiceDefinition *ServiceDefinition

	// service definitions for maker config.toml
	ChainImageName      string
	UseDataContainer    bool
	ExportedPorts       []string
	ContainerEntrypoint string

	// Return
	Result string
}

func NowDo() *Do {
	return &Do{
		ChainDefinition:   BlankChainDefinition(),
		Operations:        BlankOperation(),
		Service:           BlankService(),
		ServiceDefinition: BlankServiceDefinition(),
	}
}
