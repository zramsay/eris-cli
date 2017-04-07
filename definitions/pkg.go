package definitions

type PackageDefinition struct {
	Name    string   `json:"name" yaml:"name" toml:"name"`
	Package *Package `mapstructure:"monax" json:"monax" yaml:"monax" toml:"monax"`
}

type Package struct {
	// name of the package
	Name string `mapstructure:"name" json:"name" yaml:"name" toml:"name"`
	// string ID of the package
	PackageID string `mapstructure:"package_id" json:"package_id" yaml:"package_id" toml:"package_id"`
	// environment variables required when running the package operations
	Environment map[string]string `mapstructure:"environment" json:"environment" yaml:"environment" toml:"environment"`
	// name of the chain to use (can utilize the $chain variable)
	ChainName string `mapstructure:"chain_name" json:"chain_name" yaml:"chain_name" toml:"chain_name"`
	// ID of the chain to use (currently this is not utilized)
	ChainID string `mapstructure:"chain_id" json:"chain_id" yaml:"chain_id" toml:"chain_id"`
	// ChainTypes the package is restricted to (currently this is not utilized)
	ChainTypes []string `mapstructure:"chain_types" json:"chain_types" yaml:"chain_types" toml:"chain_types"`
	// Dependencies to be booted before the package is ran
	Dependencies *Dependencies `mapstructure:"dependencies" json:"dependencies" yaml:"dependencies" toml:"dependencies"`

	Maintainer *Maintainer `json:"maintainer,omitempty" yaml:"maintainer,omitempty" toml:"maintainer,omitempty"`
	Location   *Location   `json:"location,omitempty" yaml:"location,omitempty" toml:"location,omitempty"`
	// AppType           *AppType    `json:"app_type,omitempty" yaml:"app_type,omitempty" toml:"app_type,omitempty"`
	Chain             *Chain
	Srvs              []*Service
	Operations        *Operation
	SkipContractsPath bool
	SkipABIPath       bool

	// from epm
	Account   string
	Jobs      []*Jobs
	Libraries map[string]string
}

func BlankPackageDefinition() *PackageDefinition {
	return &PackageDefinition{
		Package: BlankPackage(),
	}
}

func BlankPackage() *Package {
	return &Package{
		Dependencies: &Dependencies{},
		Location:     BlankLocation(),
		Operations:   BlankOperation(),
		// AppType:      BlankAppType(),
		Chain: BlankChain(),
	}
}
