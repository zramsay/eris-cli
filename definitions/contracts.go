package definitions

type Package struct {
	Name      string     `json:"name" yaml:"name" toml:"name"`
	Contracts *Contracts `mapstructure:"eris" json:"eris" yaml:"eris" toml:"eris"`
}

type Contracts struct {
	// TODO: harmonize with dapps_definition_spec.md
	Name         string            `json:"name" yaml:"name" toml:"name"`
	PackageID    string            `json:"package_id" yaml:"package_id" toml:"package_id"`
	Environment  map[string]string `json:"environment" yaml:"environment" toml:"environment"`
	Dependencies *Dependencies     `json:"dependencies" yaml:"dependencies" toml:"dependencies"`
	ChainName    string            `json:"chain_name" yaml:"chain_name" toml:"chain_name"`
	ChainID      string            `mapstructure:"chain_id" json:"chain_id" yaml:"chain_id" toml:"chain_id"`
	ChainTypes   []string          `mapstructure:"chain_types" json:"chain_types" yaml:"chain_types" toml:"chain_types"`
	TestType     string            `mapstructure:"test_type" json:"test_type" yaml:"test_type" toml:"test_type"`
	DeployType   string            `mapstructure:"deploy_type" json:"deploy_type" yaml:"deploy_type" toml:"deploy_type"`
	TestTask     string            `mapstructure:"test_task" json:"test_task" yaml:"test_task" toml:"test_task"`
	DeployTask   string            `mapstructure:"deploy_task" json:"deploy_task" yaml:"deploy_task" toml:"deploy_task"`

	Maintainer *Maintainer `json:"maintainer,omitempty" yaml:"maintainer,omitempty" toml:"maintainer,omitempty"`
	Location   *Location   `json:"location,omitempty" yaml:"location,omitempty" toml:"location,omitempty"`
	Machine    *Machine    `json:"machine,omitempty" yaml:"machine,omitempty" toml:"machine,omitempty"`
	DappType   *DappType
	Chain      *Chain
	Srvs       []*Service
	Operations *Operation
}

func BlankPackage() *Package {
	return &Package{
		Contracts: BlankContracts(),
	}
}

func BlankContracts() *Contracts {
	return &Contracts{
		Maintainer: BlankMaintainer(),
		Location:   BlankLocation(),
		Machine:    BlankMachine(),
		Operations: BlankOperation(),
	}
}
