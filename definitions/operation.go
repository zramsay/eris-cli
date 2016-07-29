package definitions

type Operation struct {
	// Filled in dynamically prerun.
	SrvContainerName  string            `json:",omitempty" yaml:",omitempty" toml:",omitempty"`
	SrvContainerID    string            `json:",omitempty" yaml:",omitempty" toml:",omitempty"`
	DataContainerName string            `json:",omitempty" yaml:",omitempty" toml:",omitempty"`
	DataContainerID   string            `json:",omitempty" yaml:",omitempty" toml:",omitempty"`
	ContainerType     string            `json:",omitempty" yaml:",omitempty" toml:",omitempty"`
	Remove            bool              `json:",omitempty" yaml:",omitempty" toml:",omitempty"`
	Privileged        bool              `json:",omitempty" yaml:",omitempty" toml:",omitempty"`
	Interactive       bool              `json:",omitempty" yaml:",omitempty" toml:",omitempty"`
	Terminal          bool              `json:",omitempty" yaml:",omitempty" toml:",omitempty"`
	Follow            bool              `json:",omitempty" yaml:",omitempty" toml:",omitempty"`
	SkipCheck         bool              `json:",omitempty" yaml:",omitempty" toml:",omitempty"`
	SkipLink          bool              `json:",omitempty" yaml:",omitempty" toml:",omitempty"`
	AppName           string            `json:",omitempty" yaml:",omitempty" toml:",omitempty"`
	DockerHostConn    string            `json:",omitempty" yaml:",omitempty" toml:",omitempty"`
	Volume            string            `json:",omitempty" yaml:",omitempty" toml:",omitempty"`
	Ports             string            `json:",omitempty" yaml:",omitempty" toml:",omitempty"`
	Labels            map[string]string `json:",omitempty" yaml:",omitempty" toml:",omitempty"`
	PublishAllPorts   bool              `json:",omitempty" yaml:",omitempty" toml:",omitempty"`
	CapAdd            []string          `mapstructure:",omitempty" json:",omitempty" yaml:",omitempty" toml:",omitempty"`
	CapDrop           []string          `mapstructure:",omitempty" json:",omitempty" yaml:",omitempty" toml:",omitempty"`
	Args              []string          `mapstructure:",omitempty" json:",omitempty" yaml:",omitempty" toml:",omitempty"`
}

func BlankOperation() *Operation {
	return &Operation{}
}
