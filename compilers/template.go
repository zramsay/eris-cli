package compilers

type Compiler interface {
	Compile(files []string, version string) (Return, error)
}

//Practicing inheritance, this struct gives us access to all types of returns
type Return struct {
	SolcReturn
	//Enter your return struct here...
}
