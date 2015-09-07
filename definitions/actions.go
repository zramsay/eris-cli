package definitions

type Action struct {
	// name of the action
	Name string `json:"name" yaml:"name" toml:"name"`
	// an array of strings listing the services which eris should start prior
	// to running the steps required for the action
	Dependencies *Dependencies `json:"dependencies" yaml:"dependencies" toml:"dependencies"`
	// a chain which should be started by eris prior to running the steps
	// required for the action. can take a `$chain` string which would then
	// be passed in via a command line flag
	Chain string `json:"chain" yaml:"chain" toml:"chain"`
	// an array of strings which should be ran in a sequence of subshells
	Steps []string `json:"steps" yaml:"steps" toml:"steps"`
	// environment variables to give the subshells
	Environment map[string]string `json:"environment" yaml:"environment" toml:"environment"`

	Maintainer *Maintainer `json:"maintainer,omitempty" yaml:"maintainer,omitempty" toml:"maintainer,omitempty"`
	Location   *Location   `json:"location,omitempty" yaml:"location,omitempty" toml:"location,omitempty"`
	Machine    *Machine    `json:"machine,omitempty" yaml:"machine,omitempty" toml:"machine,omitempty"`
	Srvs       []*Service
	Operations *Operation
}

func BlankAction() *Action {
	return &Action{
		Maintainer: BlankMaintainer(),
		Location:   BlankLocation(),
		Machine:    BlankMachine(),
		Operations: BlankOperation(),
	}
}
