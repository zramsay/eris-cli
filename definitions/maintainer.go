package definitions

type Maintainer struct {
	Name  string `json:"name,omitempty" yaml:"name,omitempty" toml:"name,omitempty"`
	Email string `json:"email,omitempty" yaml:"email,omitempty" toml:"email,omitempty"`
}

func BlankMaintainer() *Maintainer {
	return &Maintainer{}
}
