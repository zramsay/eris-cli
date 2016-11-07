package definitions

type Do struct {
	Debug        bool
	Verbose      bool
	Name         string
	ChainType    string
	CSV          string
	AccountTypes []string
	Zip          bool
	Tarball      bool
	Output       bool
	Accounts     []*Account
	Result       string

	// service definitions
	ChainImageName      string
	UseDataContainer    bool
	ExportedPorts       []string
	ContainerEntrypoint string
}

func NowDo() *Do {
	return &Do{}
}
