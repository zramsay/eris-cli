package definitions

type Machine struct {
	Requires []string `json:"requires,omitempty" yaml:"requires,omitempty" toml:"requires,omitempty"`
}

func BlankMachine() *Machine {
	return &Machine{}
}
