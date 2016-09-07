package loaders

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	//"github.com/eris-ltd/eris-cli/config"
	def "github.com/eris-ltd/eris-cli/definitions"
	"github.com/eris-ltd/eris-cli/tests"
	"github.com/eris-ltd/eris-cli/util"

	"github.com/eris-ltd/common/go/common"
	log "github.com/eris-ltd/eris-logger"
)

// TODO major refactor given new config file from ecm

type ab struct {
	name string
	a, b interface{}
}

func TestMain(m *testing.M) {
	log.SetLevel(log.ErrorLevel)
	// log.SetLevel(log.InfoLevel)
	// log.SetLevel(log.DebugLevel)

	tests.IfExit(tests.TestsInit(tests.DontPull))

	exitCode := m.Run()

	tests.IfExit(tests.TestsTearDown())

	os.Exit(exitCode)
}

// const ()
// ^ some default config.toml
// than can be put anywhere
// perhaps with tests.FakeDefinitionFile
// or even better, tests.FakeChainConfigFile
// just to be clearer

// these tests should mock a chain config
// and,  when it is loaded, check that the
// appropriate fields are marshalled

// mock then load a simple chain
func TestLoadChainConfigSimple(t *testing.T) {
	// mock a chain
	// ...

	// load the chain config
	// ...
	if !checkChainConfigIsValid(t) {
		t.Fatalf("chain config not valid")

	}
}

// mock then load a complex chain
func TestLoadChainConfigNotSimple(t *testing.T) {
	// mock a chain
	// ...

	// load the chain config
	// ...
	if !checkChainConfigIsValid(t) {
		t.Fatalf("chain config not valid")
	}
}

// mock a bad chain config then ensure it doesn't load
func TestLoadBadChainConfig(t *testing.T) {
	// mock a chain
	// ...

	// load the chain config
	// ...
	//if checkChainConfigIsValid(t) {
	//	t.Fatalf("chain config is valid when it shouldn't be")
	//}
}

func checkChainConfigIsValid(t *testing.T) bool {
	// iterate through chain config fields
	// and ensure they are legit
	return true
}

// ----------------------------------------------------------
// ----------------------------------------------------------
// ---------------- OLD TESTS -------------------------------
func TestLoadChainConfigFileEmptyDefault(t *testing.T) {
}
func TestLoadChainConfigFileEmptyDefinition(t *testing.T) {
}
func TestLoadChainConfigFileEmptyDefaultAndDefinition(t *testing.T) {
}
func TestLoadChainConfigFileOverwrite(t *testing.T) {
}
func TestLoadChainConfigFileMissingDefault(t *testing.T) {
}
func TestLoadChainConfigFileBadFormatDefault(t *testing.T) {
}
func TestLoadChainConfigFileBadFormatDefinition(t *testing.T) {
}

// ----------------------------------------------------------
// ----------------------------------------------------------
// ----------------------------------------------------------
func _TestHarvestWhatWeCanFromThisExample(t *testing.T) {
	const (
		name = "test"

		defaultDefinition = `
		name = "` + name + `"
		chain_id = "` + name + `"
		description = "test chain"
		
		[service]
		name           = "random name"
		image          = "test image"
		data_container = true
		ports          = [ "1234" ]
		
		[dependencies]
		services = [ "keys" ]
		
		[maintainer]
		name = "Eris Industries"
		email = "support@erisindustries.com"
		`
	)

	if err := tests.FakeDefinitionFile(common.ChainsPath, "default", defaultDefinition); err != nil {
		t.Fatalf("cannot place a default definition file")
	}
	if err := tests.FakeDefinitionFile(filepath.Join(common.ChainsPath, name), name, ``); err != nil {
		t.Fatalf("cannot place a definition file")
	}

	d, err := LoadChainConfigFile(name)
	if err != nil {
		t.Fatalf("expected to load chain definition, got %v", err)
	}

	for _, entry := range []ab{
		{`Name`, d.Name, name},
		{`ChainID`, d.Chain.ChainID, ""},
		{`ContainerType`, d.Operations.ContainerType, def.TypeChain},
		{`SrvContainerName`, d.Operations.SrvContainerName, util.ChainContainerName(name)},
		{`DataContainerName`, d.Operations.DataContainerName, util.DataContainerName(name)},

		{`Labels["ERIS"]`, d.Operations.Labels[def.LabelEris], "true"},
		{`Labels["NAME"]`, d.Operations.Labels[def.LabelShortName], name},
		{`Labels["TYPE"]`, d.Operations.Labels[def.LabelType], def.TypeChain},

		{`Service.Name`, d.Service.Name, name},
		// [pv]: "data_container" is not loaded from the default.toml. A bug?
		{`Service.AutoData`, d.Service.AutoData, false},
		{`Service.Image`, d.Service.Image, "test image"},

		{`Dependencies`, d.Dependencies.Services, []string{"keys"}},
		{`Maintainer`, d.Maintainer.Email, "support@erisindustries.com"},
	} {
		if !reflect.DeepEqual(entry.a, entry.b) {
			t.Fatalf("marshalled definition expected %s = %#v, got %#v", entry.name, entry.b, entry.a)
		}
	}
}

func TestLoadDataDefinition(t *testing.T) {
	const (
		name = "test"
	)

	d := LoadDataDefinition(name)

	for _, entry := range []ab{
		{`ContainerType`, d.ContainerType, def.TypeData},
		{`SrvContainerName`, d.SrvContainerName, util.DataContainerName(name)},
		{`DataContainerName`, d.DataContainerName, util.DataContainerName(name)},

		{`Labels["ERIS"]`, d.Labels[def.LabelEris], "true"},
		{`Labels["NAME"]`, d.Labels[def.LabelShortName], name},
		{`Labels["TYPE"]`, d.Labels[def.LabelType], def.TypeData},
	} {
		if !reflect.DeepEqual(entry.a, entry.b) {
			t.Fatalf("definition expected %s = %#v, got %#v", entry.name, entry.b, entry.a)
		}
	}
}

func TestLoadPackageSimple(t *testing.T) {
	const (
		name = "test"

		definition = `
[eris]
name       = "` + name + `"
package_id = "` + name + `"
chain_name = "test chain"
chain_id   = "test id"
`
	)

	if err := tests.FakeDefinitionFile(common.ErisRoot, "package", definition); err != nil {
		t.Fatalf("cannot place a definition file")
	}

	d, err := LoadPackage(common.ErisRoot, name)
	if err != nil {
		t.Fatalf("expected to load definition file, got %v", err)
	}

	for _, entry := range []ab{
		{`Name`, d.Name, name},
		{`PackageID`, d.PackageID, name},
		{`ChainName`, d.ChainName, "test chain"},
		{`ChainID`, d.ChainID, "test id"},
	} {
		if !reflect.DeepEqual(entry.a, entry.b) {
			t.Fatalf("definition expected %s = %#v, got %#v", entry.name, entry.b, entry.a)
		}
	}
}

func TestLoadPackageDirectoryAndSpacesInAName(t *testing.T) {
	const (
		name = "test test"

		definition = `
name       = "` + name + `"

[eris]
name       = "` + name + `"
package_id = "` + name + `"
chain_name = "test chain"
chain_id   = "test id"
`
	)

	if err := tests.FakeDefinitionFile(common.ErisRoot, "package", definition); err != nil {
		t.Fatalf("cannot place a definition file")
	}

	d, err := LoadPackage(filepath.Join(common.ErisRoot, "package.toml"), name)
	if err != nil {
		t.Fatalf("expected to load definition file, got %v", err)
	}

	for _, entry := range []ab{
		{`Name`, d.Name, "test_test"},
		{`PackageID`, d.PackageID, name},
		{`ChainName`, d.ChainName, "test chain"},
		{`ChainID`, d.ChainID, "test id"},
	} {
		if !reflect.DeepEqual(entry.a, entry.b) {
			t.Fatalf("definition expected %s = %#v, got %#v", entry.name, entry.b, entry.a)
		}
	}
}

func TestLoadPackageNotFound1(t *testing.T) {
	const (
		name = "test"
	)

	os.Remove(filepath.Join(common.ErisRoot, "package.toml"))

	if _, err := LoadPackage("/non/existent/path", name); err == nil {
		t.Fatalf("expected definition fail to load")
	}
}

func TestLoadPackageNotFound2(t *testing.T) {
	const (
		name = "test"
	)

	os.Remove(filepath.Join(common.ErisRoot, "package.toml"))

	d, err := LoadPackage(common.ErisRoot, "")
	if err != nil {
		t.Fatalf("expected definition to load default, got %v", err)
	}

	for _, entry := range []ab{
		{`Name`, d.Name, "eris"},
		{`PackageID`, d.PackageID, ""},
		{`ChainName`, d.ChainName, ""},
	} {
		if !reflect.DeepEqual(entry.a, entry.b) {
			t.Fatalf("definition expected %s = %#v, got %#v", entry.name, entry.b, entry.a)
		}
	}
}

func TestLoadPackageNotFound3(t *testing.T) {
	const (
		name = "test"
	)

	os.Remove(filepath.Join(common.ErisRoot, "package.toml"))

	d, err := LoadPackage(common.ErisRoot, name)
	if err != nil {
		t.Fatalf("expected definition to load default, got %v", err)
	}

	for _, entry := range []ab{
		{`Name`, d.Name, "eris"},
		{`PackageID`, d.PackageID, ""},
		{`ChainName`, d.ChainName, name},
	} {
		if !reflect.DeepEqual(entry.a, entry.b) {
			t.Fatalf("definition expected %s = %#v, got %#v", entry.name, entry.b, entry.a)
		}
	}
}

func TestLoadPackageBadFormat(t *testing.T) {
	const (
		name = "test"

		definition = `
[eris]
name       = [ "keys"]
`
	)

	if err := tests.FakeDefinitionFile(common.ErisRoot, "package", definition); err != nil {
		t.Fatalf("cannot place a definition file")
	}

	if _, err := LoadPackage(common.ErisRoot, name); err == nil {
		t.Fatalf("expected definition fail to load")
	}
}

func TestLoadServiceDefinitionSimple(t *testing.T) {
	const (
		name       = "test"
		definition = `
name = "` + name + `"
description = "description"
status = "in production"

[service]
image = "test image"
data_container = true
ports = [ "1234" ]

[location]
repository = "https://example.com"
`
	)

	if err := tests.FakeDefinitionFile(common.ServicesPath, name, definition); err != nil {
		t.Fatalf("cannot place a definition file")
	}

	d, err := LoadServiceDefinition(name)
	if err != nil {
		t.Fatalf("expected definition to load, got %v", err)
	}

	for _, entry := range []ab{
		{`Name`, d.Name, name},

		{`ContainerType`, d.Operations.ContainerType, def.TypeService},
		{`SrvContainerName`, d.Operations.SrvContainerName, util.ServiceContainerName(name)},
		{`DataContainerName`, d.Operations.DataContainerName, util.DataContainerName(name)},

		{`Labels["ERIS"]`, d.Operations.Labels[def.LabelEris], "true"},
		{`Labels["NAME"]`, d.Operations.Labels[def.LabelShortName], name},
		{`Labels["TYPE"]`, d.Operations.Labels[def.LabelType], def.TypeService},

		{`Service.Name`, d.Service.Name, name},
		{`Service.AutoData`, d.Service.AutoData, true},
		{`Service.Image`, d.Service.Image, "test image"},
		{`Service.Ports`, d.Service.Ports, []string{"1234"}},

		{`Location`, d.Location.Repository, "https://example.com"},
	} {
		if !reflect.DeepEqual(entry.a, entry.b) {
			t.Fatalf("definition expected %s = %#v, got %#v", entry.name, entry.b, entry.a)
		}
	}
}

func TestLoadServiceDefinitionAlmostEmpty(t *testing.T) {
	const (
		name       = "test"
		definition = `
[service]
image = "test image"
`
	)

	if err := tests.FakeDefinitionFile(common.ServicesPath, name, definition); err != nil {
		t.Fatalf("cannot place a definition file")
	}

	d, err := LoadServiceDefinition(name)
	if err != nil {
		t.Fatalf("expected definition to load, got %v", err)
	}

	for _, entry := range []ab{
		{`Name`, d.Name, "test image"},

		{`ContainerType`, d.Operations.ContainerType, def.TypeService},
		{`SrvContainerName`, d.Operations.SrvContainerName, util.ServiceContainerName("test image")},
		{`DataContainerName`, d.Operations.DataContainerName, util.DataContainerName("test image")},

		{`Labels["ERIS"]`, d.Operations.Labels[def.LabelEris], "true"},
		{`Labels["NAME"]`, d.Operations.Labels[def.LabelShortName], name},
		{`Labels["TYPE"]`, d.Operations.Labels[def.LabelType], def.TypeService},

		{`Service.Name`, d.Service.Name, "test image"},
		{`Service.Image`, d.Service.Image, "test image"},
	} {
		if !reflect.DeepEqual(entry.a, entry.b) {
			t.Fatalf("definition expected %s = %#v, got %#v", entry.name, entry.b, entry.a)
		}
	}
}

func TestLoadServiceDefinitionEmpty(t *testing.T) {
	const (
		name = "test"
	)

	if err := tests.FakeDefinitionFile(common.ServicesPath, name, ``); err != nil {
		t.Fatalf("cannot place a definition file")
	}

	if _, err := LoadServiceDefinition(name); err == nil {
		t.Fatalf("expected definition fail to load")
	}
}

func TestLoadServiceDefinitionMissing(t *testing.T) {
	const (
		name = "test"
	)

	os.Remove(filepath.Join(common.ServicesPath, name+".toml"))

	if _, err := LoadServiceDefinition(name); err == nil {
		t.Fatalf("expected definition fail to load")
	}
}

func TestLoadServiceDefinitionBadFormat(t *testing.T) {
	const (
		name = "test"

		definition = `
[service]
image = [ "keys" ]
`
	)

	if err := tests.FakeDefinitionFile(common.ServicesPath, name, definition); err != nil {
		t.Fatalf("cannot place a definition file")
	}

	if _, err := LoadServiceDefinition(name); err == nil {
		t.Fatalf("expected definition fail to load")
	}
}

func TestMockServiceDefinition(t *testing.T) {
	const (
		name = "test"
	)

	d := MockServiceDefinition(name)

	for _, entry := range []ab{
		{`Name`, d.Name, name},

		{`ContainerType`, d.Operations.ContainerType, def.TypeService},
		{`SrvContainerName`, d.Operations.SrvContainerName, util.ServiceContainerName(name)},
		{`DataContainerName`, d.Operations.DataContainerName, util.DataContainerName(name)},

		{`Labels["ERIS"]`, d.Operations.Labels[def.LabelEris], "true"},
		{`Labels["NAME"]`, d.Operations.Labels[def.LabelShortName], name},
		{`Labels["TYPE"]`, d.Operations.Labels[def.LabelType], def.TypeService},

		{`Service.Name`, d.Service.Name, name},
		// [pv]: Mock is allowed to return an empty image while load isn't?
		{`Service.Image`, d.Service.Image, ""},
	} {
		if !reflect.DeepEqual(entry.a, entry.b) {
			t.Fatalf("definition expected %s = %#v, got %#v", entry.name, entry.b, entry.a)
		}
	}
}

func TestServiceFinalizeLoadBlankNames(t *testing.T) {
	const (
		name = "test"
	)

	d := def.BlankServiceDefinition()
	d.Service.Image = name

	ServiceFinalizeLoad(d)
	for _, entry := range []ab{
		{`Name`, d.Name, name},

		{`SrvContainerName`, d.Operations.SrvContainerName, util.ServiceContainerName(name)},
		{`DataContainerName`, d.Operations.DataContainerName, util.DataContainerName(name)},

		{`Service.Name`, d.Service.Name, name},
		{`Service.Image`, d.Service.Image, name},
	} {
		if !reflect.DeepEqual(entry.a, entry.b) {
			t.Fatalf("definition expected %s = %#v, got %#v", entry.name, entry.b, entry.a)
		}
	}
}

func TestServiceFinalizeLoadBlankName(t *testing.T) {
	const (
		name = "test"
	)

	d := def.BlankServiceDefinition()
	d.Service.Name = name

	ServiceFinalizeLoad(d)
	for _, entry := range []ab{
		{`Name`, d.Name, name},

		{`SrvContainerName`, d.Operations.SrvContainerName, util.ServiceContainerName(name)},
		{`DataContainerName`, d.Operations.DataContainerName, util.DataContainerName(name)},

		{`Service.Name`, d.Service.Name, name},
		{`Service.Image`, d.Service.Image, ""},
	} {
		if !reflect.DeepEqual(entry.a, entry.b) {
			t.Fatalf("definition expected %s = %#v, got %#v", entry.name, entry.b, entry.a)
		}
	}
}

func TestServiceFinalizeLoadBlankServiceName(t *testing.T) {
	const (
		name = "test"
	)

	d := def.BlankServiceDefinition()
	d.Name = name

	ServiceFinalizeLoad(d)
	for _, entry := range []ab{
		{`Name`, d.Name, name},

		{`SrvContainerName`, d.Operations.SrvContainerName, util.ServiceContainerName(name)},
		{`DataContainerName`, d.Operations.DataContainerName, util.DataContainerName(name)},

		{`Service.Name`, d.Service.Name, name},
		{`Service.Image`, d.Service.Image, ""},
	} {
		if !reflect.DeepEqual(entry.a, entry.b) {
			t.Fatalf("definition expected %s = %#v, got %#v", entry.name, entry.b, entry.a)
		}
	}
}

func TestServiceFinalizeLoadBlankAllTheThings(t *testing.T) {
	defer func() {
		recover()
	}()

	d := def.BlankServiceDefinition()

	ServiceFinalizeLoad(d)

	t.Fatalf("expected finalize to panic")
}
