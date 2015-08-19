package definitions

type DappType struct {
	Name       string
	BaseImage  string
	DeployCmd  string
	TestCmd    string
	EntryPoint string
	ChainTypes []string
}

func AllDappTypes() map[string]*DappType {
	dapps := make(map[string]*DappType)
	dapps["embark"] = EmbarkDapp()
	dapps["pyepm"] = PyEpmDapp()
	dapps["sunit"] = SUnitDapp()
	dapps["manual"] = GulpDapp()
	return dapps
}

func EmbarkDapp() *DappType {
	dapp := BlankDappType()
	dapp.Name = "embark"
	dapp.BaseImage = "erisindustries/embark_base"
	dapp.DeployCmd = "deploy" // +blockchainname
	dapp.TestCmd = "spec"
	dapp.EntryPoint = "embark"
	dapp.ChainTypes = []string{"eth"}
	return dapp
}

func PyEpmDapp() *DappType {
	dapp := BlankDappType()
	dapp.Name = "pyepm"
	dapp.BaseImage = "erisindustries/pyepm_base"
	dapp.DeployCmd = ""  // +YAML to deploy
	dapp.TestCmd = "nil" //n/a
	dapp.EntryPoint = "pyepm"
	dapp.ChainTypes = []string{"eth"}
	return dapp
}

func SUnitDapp() *DappType {
	dapp := BlankDappType()
	dapp.Name = "sunit"
	dapp.BaseImage = "erisindustries/sunit_base"
	dapp.DeployCmd = "nil" // n/a
	dapp.TestCmd = "--coverage"
	dapp.EntryPoint = "sunit"
	dapp.ChainTypes = []string{"mint"}
	return dapp
}

func GulpDapp() *DappType {
	dapp := BlankDappType()
	dapp.Name = "manual"
	dapp.BaseImage = "erisindustries/gulp"
	dapp.DeployCmd = "" //+TASK
	dapp.TestCmd = ""   //+TASK
	dapp.EntryPoint = "gulp"
	dapp.ChainTypes = []string{"eth", "mint"}
	return dapp
}

func BlankDappType() *DappType {
	return &DappType{}
}
