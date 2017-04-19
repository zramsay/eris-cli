package compilers

// An interface to denote a compiler. It implements one function, Compile, which takes in a slice of files and a version string
// and returns a Compile Return or an error
type Compiler interface {
	Compile(files []string, version string) (Return, error)
}

//Practicing inheritance, this struct gives us access to all types of returns
//This is written to be extendable to other compilers
type Return struct {
	SolcReturn
	//Enter your return struct here...
}
