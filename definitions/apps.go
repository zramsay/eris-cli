package definitions

import (
	"path"

	"github.com/eris-ltd/eris-cli/version"
)

// [csk]: TODO: refactor what the hell we're doing here. move away from the restrictive eth app types into the larger area of play.
// NOTE: this is currently unused.
type AppType struct {
	Name       string
	BaseImage  string
	DeployCmd  string
	TestCmd    string
	EntryPoint string
	ChainTypes []string
}

func AllAppTypes() map[string]*AppType {
	apps := make(map[string]*AppType)
	apps["epm"] = EPMApp()
	apps["embark"] = EmbarkApp()
	apps["sunit"] = SUnitApp()
	apps["manual"] = GulpApp()
	return apps
}

func EPMApp() *AppType {
	app := BlankAppType()
	app.Name = "epm"
	app.BaseImage = path.Join(version.DefaultRegistry, version.ImagePM)
	app.EntryPoint = "epm --chain tcp://chain:46657 --sign http://keys:4767"
	app.DeployCmd = ""
	app.TestCmd = ""
	app.ChainTypes = []string{"mint"}
	return app
}

func EmbarkApp() *AppType {
	app := BlankAppType()
	app.Name = "embark"
	app.BaseImage = "quay.io/eris/embark_base"
	app.EntryPoint = "embark"
	app.DeployCmd = "deploy" // +blockchainname
	app.TestCmd = "spec"
	app.ChainTypes = []string{"eth"}
	return app
}

func SUnitApp() *AppType {
	app := BlankAppType()
	app.Name = "sunit"
	app.BaseImage = "quay.io/eris/sunit_base"
	app.EntryPoint = "sunit"
	app.DeployCmd = "nil" // n/a
	app.TestCmd = "--coverage"
	app.ChainTypes = []string{"mint"}
	return app
}

func GulpApp() *AppType {
	app := BlankAppType()
	app.Name = "manual"
	app.BaseImage = "quay.io/eris/gulp"
	app.EntryPoint = "gulp"
	app.DeployCmd = "" //+TASK
	app.TestCmd = ""   //+TASK
	app.ChainTypes = []string{"eth", "mint"}
	return app
}

func BlankAppType() *AppType {
	return &AppType{}
}
