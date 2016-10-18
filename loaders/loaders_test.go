package loaders

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/eris-ltd/eris-cli/config"
	"github.com/eris-ltd/eris-cli/definitions"
	"github.com/eris-ltd/eris-cli/log"
	"github.com/eris-ltd/eris-cli/testutil"
	"github.com/eris-ltd/eris-cli/util"

	"github.com/eris-ltd/common/go/common"
)

type ab struct {
	name string
	a, b interface{}
}

func TestMain(m *testing.M) {
	log.SetLevel(log.ErrorLevel)
	// log.SetLevel(log.InfoLevel)
	// log.SetLevel(log.DebugLevel)

	testutil.IfExit(testutil.Init())

	exitCode := m.Run()

	testutil.IfExit(testutil.TearDown())

	os.Exit(exitCode)
}

func TestLoadChainDefinitionEmptyDefault(t *testing.T) {
	const (
		name = "test"

		definition = `
name = "` + name + `"
chain_id = "` + name + `"
description = "test chain"

[service]
name           = "random name"
image          = "test image"
data_container = true
ports          = [ "1234" ]
`
	)

	mockConfigPathFile(t, name)
	defer removeConfigPathFile(t, name)

	if err := testutil.FakeDefinitionFile(filepath.Join(common.ChainsPath, name), "config", definition); err != nil {
		t.Fatalf("cannot place a definition file")
	}

	d, err := LoadChainDefinition(name)
	if err != nil {
		t.Fatalf("expected to load chain definition, got %v", err)
	}

	for _, entry := range []ab{
		{`Name`, d.Name, name},
		{`ContainerType`, d.Operations.ContainerType, definitions.TypeChain},
		{`SrvContainerName`, d.Operations.SrvContainerName, util.ChainContainerName(name)},
		{`DataContainerName`, d.Operations.DataContainerName, util.DataContainerName(name)},

		{`Labels["ERIS"]`, d.Operations.Labels[definitions.LabelEris], "true"},
		{`Labels["NAME"]`, d.Operations.Labels[definitions.LabelShortName], name},
		{`Labels["TYPE"]`, d.Operations.Labels[definitions.LabelType], definitions.TypeChain},

		{`Service.Name`, d.Service.Name, name},
		{`Service.AutoData`, d.Service.AutoData, true},
		{`Service.Image`, d.Service.Image, "test image"},
		{`Service.Ports`, d.Service.Ports, []string{"1234"}},
	} {
		if !reflect.DeepEqual(entry.a, entry.b) {
			t.Fatalf("marshalled definition expected %s = %#v, got %#v", entry.name, entry.b, entry.a)
		}
	}
}

func TestLoadChainDefinitionWithoutCONFIGPATH(t *testing.T) {

	const (
		name = "test"

		definition = `
name = "` + name + `"
chain_id = "` + name + `"
description = "test chain"

[service]
name           = "random name"
image          = "test image"
data_container = true
ports          = [ "1234" ]
`
	)

	//mockConfigPathFile(t, name)
	//defer removeConfigPathFile(t, name)

	if err := testutil.FakeDefinitionFile(filepath.Join(common.ChainsPath, name), "config", definition); err != nil {
		t.Fatalf("cannot place a definition file")
	}

	_, err := LoadChainDefinition(name)
	if err == nil {
		t.Fatalf("expected chain definition to not load, got %v", err)
	}
}

func TestLoadChainDefinitionEmptyDefinition(t *testing.T) {
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
name = "Monax Industries"
email = "support@monax.io"
`
	)

	mockConfigPathFile(t, name)
	defer removeConfigPathFile(t, name)

	if err := testutil.FakeDefinitionFile(common.ChainsPath, "default", defaultDefinition); err != nil {
		t.Fatalf("cannot place a default definition file")
	}
	if err := testutil.FakeDefinitionFile(filepath.Join(common.ChainsPath, name), name, ``); err != nil {
		t.Fatalf("cannot place a definition file")
	}

	d, err := LoadChainDefinition(name)
	if err != nil {
		t.Fatalf("expected to load chain definition, got %v", err)
	}

	for _, entry := range []ab{
		{`Name`, d.Name, name},
		{`ContainerType`, d.Operations.ContainerType, definitions.TypeChain},
		{`SrvContainerName`, d.Operations.SrvContainerName, util.ChainContainerName(name)},
		{`DataContainerName`, d.Operations.DataContainerName, util.DataContainerName(name)},

		{`Labels["ERIS"]`, d.Operations.Labels[definitions.LabelEris], "true"},
		{`Labels["NAME"]`, d.Operations.Labels[definitions.LabelShortName], name},
		{`Labels["TYPE"]`, d.Operations.Labels[definitions.LabelType], definitions.TypeChain},

		{`Service.Name`, d.Service.Name, name},
		{`Service.Image`, d.Service.Image, "test image"},
	} {
		if !reflect.DeepEqual(entry.a, entry.b) {
			t.Fatalf("marshalled definition expected %s = %#v, got %#v", entry.name, entry.b, entry.a)
		}
	}
}

func TestLoadChainDefinitionEmptyDefaultAndDefinition(t *testing.T) {
	const (
		name = "test"
	)

	mockConfigPathFile(t, name)
	defer removeConfigPathFile(t, name)

	if err := testutil.FakeDefinitionFile(common.ChainsPath, "default", ``); err != nil {
		t.Fatalf("cannot place a default definition file")
	}
	if err := testutil.FakeDefinitionFile(filepath.Join(common.ChainsPath, name), name, ``); err != nil {
		t.Fatalf("cannot place a definition file")
	}

	d, err := LoadChainDefinition(name)
	if err != nil {
		t.Fatalf("expected to load chain definition, got %v", err)
	}

	for _, entry := range []ab{
		{`Name`, d.Name, name},
		{`ContainerType`, d.Operations.ContainerType, definitions.TypeChain},
		{`SrvContainerName`, d.Operations.SrvContainerName, util.ChainContainerName(name)},
		{`DataContainerName`, d.Operations.DataContainerName, util.DataContainerName(name)},

		{`Labels["ERIS"]`, d.Operations.Labels[definitions.LabelEris], "true"},
		{`Labels["NAME"]`, d.Operations.Labels[definitions.LabelShortName], name},
		{`Labels["TYPE"]`, d.Operations.Labels[definitions.LabelType], definitions.TypeChain},

		{`Service.Name`, d.Service.Name, name},
	} {
		if !reflect.DeepEqual(entry.a, entry.b) {
			t.Fatalf("marshalled definition expected %s = %#v, got %#v", entry.name, entry.b, entry.a)
		}
	}
}

func TestLoadChainDefinitionOverwrite(t *testing.T) {
	const (
		name = "test"

		definition = `
name = "` + name + `"
chain_id = "` + name + `"
description = "test chain"

[service]
name           = "random name"
image          = "test image"
data_container = true
ports          = [ "4321" ]
`
	)

	mockConfigPathFile(t, name)
	defer removeConfigPathFile(t, name)

	if err := testutil.FakeDefinitionFile(filepath.Join(common.ChainsPath, name), name, definition); err != nil {
		t.Fatalf("cannot place a definition file")
	}

	d, err := LoadChainDefinition(name)
	if err != nil {
		t.Fatalf("expected to load chain definition, got %v", err)
	}

	for _, entry := range []ab{
		{`Name`, d.Name, name},
		{`ContainerType`, d.Operations.ContainerType, definitions.TypeChain},
		{`SrvContainerName`, d.Operations.SrvContainerName, util.ChainContainerName(name)},
		{`DataContainerName`, d.Operations.DataContainerName, util.DataContainerName(name)},

		{`Labels["ERIS"]`, d.Operations.Labels[definitions.LabelEris], "true"},
		{`Labels["NAME"]`, d.Operations.Labels[definitions.LabelShortName], name},
		{`Labels["TYPE"]`, d.Operations.Labels[definitions.LabelType], definitions.TypeChain},

		{`Service.Name`, d.Service.Name, name},
		{`Service.AutoData`, d.Service.AutoData, true},
		{`Service.Image`, d.Service.Image, "test image"},
	} {
		if !reflect.DeepEqual(entry.a, entry.b) {
			t.Fatalf("marshalled definition expected %s = %#v, got %#v", entry.name, entry.b, entry.a)
		}
	}
}

func TestChainsAsAServiceSimple(t *testing.T) {
	const (
		name = "test"

		definition = `
name = "` + name + `"
chain_id = "` + name + `"
description = "test chain"

[service]
name           = "random name"
data_container = true
ports          = [ "1234" ]
image          = "test image"
`
	)

	mockConfigPathFile(t, name)
	defer removeConfigPathFile(t, name)

	if err := testutil.FakeDefinitionFile(common.ChainsPath, "default", ``); err != nil {
		t.Fatalf("cannot place a default definition file")
	}
	if err := testutil.FakeDefinitionFile(filepath.Join(common.ChainsPath, name), name, definition); err != nil {
		t.Fatalf("cannot place a definition file")
	}

	s, err := ChainsAsAService(name)
	if err != nil {
		t.Fatalf("expected to load chain definition, got %v", err)
	}

	chainID := fmt.Sprintf("CHAIN_ID=%s", name)
	chainName := fmt.Sprintf("CHAIN_NAME=%s", name)

	for _, entry := range []ab{
		{`Name`, s.Name, name},
		{`ContainerType`, s.Operations.ContainerType, definitions.TypeChain},
		{`SrvContainerName`, s.Operations.SrvContainerName, util.ChainContainerName(name)},
		{`DataContainerName`, s.Operations.DataContainerName, util.DataContainerName(name)},

		{`Labels["ERIS"]`, s.Operations.Labels[definitions.LabelEris], "true"},
		{`Labels["NAME"]`, s.Operations.Labels[definitions.LabelShortName], name},
		{`Labels["TYPE"]`, s.Operations.Labels[definitions.LabelType], definitions.TypeChain},

		{`Service.Name`, s.Service.Name, name},
		{`Service.AutoData`, s.Service.AutoData, true},
		// [pv]: not "test image", but erisdb image. A bug?
		{`Service.Image`, s.Service.Image, path.Join(config.Global.DefaultRegistry, config.Global.ImageDB)},
		{`Service.Environment`, s.Service.Environment, []string{chainID, chainName}},
	} {
		if !reflect.DeepEqual(entry.a, entry.b) {
			t.Fatalf("marshalled definition expected %s = %#v, got %#v", entry.name, entry.b, entry.a)
		}
	}
}

func _TestChainsAsAServiceMissing(t *testing.T) {
	const (
		name = "test"
	)

	os.Remove(filepath.Join(common.ChainsPath, name, name+".toml"))

	if err := testutil.FakeDefinitionFile(common.ChainsPath, "default", ``); err != nil {
		t.Fatalf("cannot place a default definition file")
	}

	if _, err := ChainsAsAService(name); err == nil {
		t.Fatalf("expected chains as a service to fail")
	}
}

func TestLoadDataDefinition(t *testing.T) {
	const (
		name = "test"
	)

	d := LoadDataDefinition(name)

	for _, entry := range []ab{
		{`ContainerType`, d.ContainerType, definitions.TypeData},
		{`SrvContainerName`, d.SrvContainerName, util.DataContainerName(name)},
		{`DataContainerName`, d.DataContainerName, util.DataContainerName(name)},

		{`Labels["ERIS"]`, d.Labels[definitions.LabelEris], "true"},
		{`Labels["NAME"]`, d.Labels[definitions.LabelShortName], name},
		{`Labels["TYPE"]`, d.Labels[definitions.LabelType], definitions.TypeData},
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

	if err := testutil.FakeDefinitionFile(common.ErisRoot, "package", definition); err != nil {
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

	if err := testutil.FakeDefinitionFile(common.ErisRoot, "package", definition); err != nil {
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

	if err := testutil.FakeDefinitionFile(common.ErisRoot, "package", definition); err != nil {
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

	if err := testutil.FakeDefinitionFile(common.ServicesPath, name, definition); err != nil {
		t.Fatalf("cannot place a definition file")
	}

	d, err := LoadServiceDefinition(name)
	if err != nil {
		t.Fatalf("expected definition to load, got %v", err)
	}

	for _, entry := range []ab{
		{`Name`, d.Name, name},

		{`ContainerType`, d.Operations.ContainerType, definitions.TypeService},
		{`SrvContainerName`, d.Operations.SrvContainerName, util.ServiceContainerName(name)},
		{`DataContainerName`, d.Operations.DataContainerName, util.DataContainerName(name)},

		{`Labels["ERIS"]`, d.Operations.Labels[definitions.LabelEris], "true"},
		{`Labels["NAME"]`, d.Operations.Labels[definitions.LabelShortName], name},
		{`Labels["TYPE"]`, d.Operations.Labels[definitions.LabelType], definitions.TypeService},

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

	if err := testutil.FakeDefinitionFile(common.ServicesPath, name, definition); err != nil {
		t.Fatalf("cannot place a definition file")
	}

	d, err := LoadServiceDefinition(name)
	if err != nil {
		t.Fatalf("expected definition to load, got %v", err)
	}

	for _, entry := range []ab{
		{`Name`, d.Name, "test image"},

		{`ContainerType`, d.Operations.ContainerType, definitions.TypeService},
		{`SrvContainerName`, d.Operations.SrvContainerName, util.ServiceContainerName("test image")},
		{`DataContainerName`, d.Operations.DataContainerName, util.DataContainerName("test image")},

		{`Labels["ERIS"]`, d.Operations.Labels[definitions.LabelEris], "true"},
		{`Labels["NAME"]`, d.Operations.Labels[definitions.LabelShortName], name},
		{`Labels["TYPE"]`, d.Operations.Labels[definitions.LabelType], definitions.TypeService},

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

	if err := testutil.FakeDefinitionFile(common.ServicesPath, name, ``); err != nil {
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

	if err := testutil.FakeDefinitionFile(common.ServicesPath, name, definition); err != nil {
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

		{`ContainerType`, d.Operations.ContainerType, definitions.TypeService},
		{`SrvContainerName`, d.Operations.SrvContainerName, util.ServiceContainerName(name)},
		{`DataContainerName`, d.Operations.DataContainerName, util.DataContainerName(name)},

		{`Labels["ERIS"]`, d.Operations.Labels[definitions.LabelEris], "true"},
		{`Labels["NAME"]`, d.Operations.Labels[definitions.LabelShortName], name},
		{`Labels["TYPE"]`, d.Operations.Labels[definitions.LabelType], definitions.TypeService},

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

	d := definitions.BlankServiceDefinition()
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

	d := definitions.BlankServiceDefinition()
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

	d := definitions.BlankServiceDefinition()
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

	d := definitions.BlankServiceDefinition()

	ServiceFinalizeLoad(d)

	t.Fatalf("expected finalize to panic")
}

func mockConfigPathFile(t *testing.T, name string) {
	// this occures in [func setupChain] in chains/operate.go
	// here we mock it for LoadChainDefinition to work
	configPath := filepath.Join(common.ChainsPath, name)
	fileName := filepath.Join(common.ChainsPath, name, "CONFIG_PATH")

	if err := os.MkdirAll(configPath, 0777); err != nil {
		t.Fatalf("error making chain directory: %v", err)
	}

	if err := ioutil.WriteFile(fileName, []byte(configPath), 0666); err != nil {
		t.Fatalf("error writing CONFIG_PATH file: %v", err)
	}
}

func removeConfigPathFile(t *testing.T, name string) {
	fileName := filepath.Join(common.ChainsPath, name, "CONFIG_PATH")
	if err := os.Remove(fileName); err != nil {
		t.Fatalf("error making chain directory: %v", err)
	}
}
