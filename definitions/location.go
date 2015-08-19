package definitions

type Location struct {
	Repository string `json:"repository,omitempty" yaml:"repository,omitempty" toml:"repository,omitempty"`
	IPFSHash   string `json:"ipfs_hash,omitempty" yaml:"ipfs_hash,omitempty" toml:"ipfs_hash,omitempty"`
}

func BlankLocation() *Location {
	return &Location{}
}
