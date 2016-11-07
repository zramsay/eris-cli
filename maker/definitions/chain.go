package definitions

type Chain struct {
	Name    string
	Type    string
	Genesis *interface{}
}

func BlankChain() *Chain {
	return &Chain{}
}
