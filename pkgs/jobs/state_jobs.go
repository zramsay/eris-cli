package jobs

// ------------------------------------------------------------------------
// State Jobs
// ------------------------------------------------------------------------

type DumpState struct {
	WithValidators bool   `mapstructure:"include-validators" yaml:"include-validators"`
	ToIPFS         bool   `mapstructure:"to-ipfs" yaml:"to-ipfs"`
	ToFile         bool   `mapstructure:"to-file" yaml:"to-file"`
	IPFSHost       string `mapstructure:"ipfs-host" yaml:"ipfs-host"`
	FilePath       string `mapstructure:"file" yaml:"file"`
}

type RestoreState struct {
	FromIPFS bool   `mapstructure:"from-ipfs" yaml:"from-ipfs"`
	FromFile bool   `mapstructure:"from-file" yaml:"from-file"`
	IPFSHost string `mapstructure:"ipfs-host" yaml:"ipfs-host"`
	FilePath string `mapstructure:"file" yaml:"file"`
}
