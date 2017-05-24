package loaders

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/monax/monax/config"
	"github.com/monax/monax/definitions"
	"github.com/monax/monax/log"
	"github.com/monax/monax/testutil"
	"github.com/monax/monax/util"
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

[dependencies]
services       = [ "keys" ]
`
	)

	if err := testutil.FakeDefinitionFile(filepath.Join(config.ChainsPath, name), "config", definition); err != nil {
		t.Fatalf("cannot place a definition file")
	}

	d, err := LoadChainDefinition(name, filepath.Join(config.ChainsPath, name, "config"))
	if err != nil {
		t.Fatalf("expected to load chain definition, got %v", err)
	}

	for _, entry := range []ab{
		{`Name`, d.Name, name},
		{`ContainerType`, d.Operations.ContainerType, definitions.TypeChain},
		{`SrvContainerName`, d.Operations.SrvContainerName, util.ChainContainerName(name)},
		{`DataContainerName`, d.Operations.DataContainerName, util.DataContainerName(name)},

		{`Labels["MONAX"]`, d.Operations.Labels[definitions.LabelMonax], "true"},
		{`Labels["NAME"]`, d.Operations.Labels[definitions.LabelShortName], name},
		{`Labels["TYPE"]`, d.Operations.Labels[definitions.LabelType], definitions.TypeChain},

		{`Service.Name`, d.Service.Name, name},
		{`Service.AutoData`, d.Service.AutoData, true},
		{`Service.Image`, d.Service.Image, "test image"},
		{`Service.Ports`, d.Service.Ports, []string{"1234"}},

		{`Dependencies`, d.Dependencies.Services, []string{"keys"}},
	} {
		if !reflect.DeepEqual(entry.a, entry.b) {
			t.Fatalf("marshalled definition expected %s = %#v, got %#v", entry.name, entry.b, entry.a)
		}
	}
}

func TestLoadChainDefinitionWithoutPath(t *testing.T) {
	const (
		name = "test"
	)

	d, err := LoadChainDefinition(name)
	if err != nil {
		t.Fatalf("expected chain definition to load, got %v", err)
	}

	for _, entry := range []ab{
		{`Name`, d.Name, name},
		{`ContainerType`, d.Operations.ContainerType, definitions.TypeChain},
		{`SrvContainerName`, d.Operations.SrvContainerName, util.ChainContainerName(name)},
		{`DataContainerName`, d.Operations.DataContainerName, util.DataContainerName(name)},

		{`Labels["MONAX"]`, d.Operations.Labels[definitions.LabelMonax], "true"},
		{`Labels["NAME"]`, d.Operations.Labels[definitions.LabelShortName], name},
		{`Labels["TYPE"]`, d.Operations.Labels[definitions.LabelType], definitions.TypeChain},

		{`Service.Name`, d.Service.Name, name},
	} {
		if !reflect.DeepEqual(entry.a, entry.b) {
			t.Fatalf("marshalled definition expected %s = %#v, got %#v", entry.name, entry.b, entry.a)
		}
	}

}

func TestLoadChainDefinitionEmptyDefinition(t *testing.T) {
	const (
		name = "test"
	)

	if err := testutil.FakeDefinitionFile(filepath.Join(config.ChainsPath, name), name, ``); err != nil {
		t.Fatalf("cannot place a definition file")
	}

	d, err := LoadChainDefinition(name, filepath.Join(config.ChainsPath, name, name))
	if err != nil {
		t.Fatalf("expected to load chain definition, got %v", err)
	}

	for _, entry := range []ab{
		{`Name`, d.Name, name},
		{`ContainerType`, d.Operations.ContainerType, definitions.TypeChain},
		{`SrvContainerName`, d.Operations.SrvContainerName, util.ChainContainerName(name)},
		{`DataContainerName`, d.Operations.DataContainerName, util.DataContainerName(name)},

		{`Labels["MONAX"]`, d.Operations.Labels[definitions.LabelMonax], "true"},
		{`Labels["NAME"]`, d.Operations.Labels[definitions.LabelShortName], name},
		{`Labels["TYPE"]`, d.Operations.Labels[definitions.LabelType], definitions.TypeChain},

		{`Service.Name`, d.Service.Name, name},
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

	if err := testutil.FakeDefinitionFile(config.ChainsPath, "default", ``); err != nil {
		t.Fatalf("cannot place a default definition file")
	}
	if err := testutil.FakeDefinitionFile(filepath.Join(config.ChainsPath, name), name, ``); err != nil {
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

		{`Labels["MONAX"]`, d.Operations.Labels[definitions.LabelMonax], "true"},
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

	if err := testutil.FakeDefinitionFile(filepath.Join(config.ChainsPath, name), name, definition); err != nil {
		t.Fatalf("cannot place a definition file")
	}

	d, err := LoadChainDefinition(name, filepath.Join(config.ChainsPath, name, name))
	if err != nil {
		t.Fatalf("expected to load chain definition, got %v", err)
	}

	for _, entry := range []ab{
		{`Name`, d.Name, name},
		{`ContainerType`, d.Operations.ContainerType, definitions.TypeChain},
		{`SrvContainerName`, d.Operations.SrvContainerName, util.ChainContainerName(name)},
		{`DataContainerName`, d.Operations.DataContainerName, util.DataContainerName(name)},

		{`Labels["MONAX"]`, d.Operations.Labels[definitions.LabelMonax], "true"},
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

func TestLoadDataDefinition(t *testing.T) {
	const (
		name = "test"
	)

	d := LoadDataDefinition(name)

	for _, entry := range []ab{
		{`ContainerType`, d.ContainerType, definitions.TypeData},
		{`SrvContainerName`, d.SrvContainerName, util.DataContainerName(name)},
		{`DataContainerName`, d.DataContainerName, util.DataContainerName(name)},

		{`Labels["MONAX"]`, d.Labels[definitions.LabelMonax], "true"},
		{`Labels["NAME"]`, d.Labels[definitions.LabelShortName], name},
		{`Labels["TYPE"]`, d.Labels[definitions.LabelType], definitions.TypeData},
	} {
		if !reflect.DeepEqual(entry.a, entry.b) {
			t.Fatalf("definition expected %s = %#v, got %#v", entry.name, entry.b, entry.a)
		}
	}
}

/* TODO: [RJ] - https://github.com/monax/monax/issues/1173
func TestLoadPackageSimple(t *testing.T) {
	const (
		name = "test"

		definition = `
[monax]
name       = "` + name + `"
package_id = "` + name + `"
chain_name = "test chain"
chain_id   = "test id"
`
	)

	if err := testutil.FakeDefinitionFile(config.MonaxRoot, "package", definition); err != nil {
		t.Fatalf("cannot place a definition file")
	}

	d, err := LoadPackage(name)
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

[monax]
name       = "` + name + `"
package_id = "` + name + `"
chain_name = "test chain"
chain_id   = "test id"
`
	)

	if err := testutil.FakeDefinitionFile(config.MonaxRoot, "package", definition); err != nil {
		t.Fatalf("cannot place a definition file")
	}

	d, err := LoadPackage(name)
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

	os.Remove(filepath.Join(config.MonaxRoot, "package.toml"))

	if _, err := LoadPackage(name); err == nil {
		t.Fatalf("expected definition fail to load")
	}
}

func TestLoadPackageNotFound2(t *testing.T) {
	const (
		name = "test"
	)

	os.Remove(filepath.Join(config.MonaxRoot, "package.toml"))

	d, err := LoadPackage("")
	if err != nil {
		t.Fatalf("expected definition to load default, got %v", err)
	}

	for _, entry := range []ab{
		{`Name`, d.Name, "monax"},
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

	os.Remove(filepath.Join(config.MonaxRoot, "package.toml"))

	d, err := LoadPackage(name)
	if err != nil {
		t.Fatalf("expected definition to load default, got %v", err)
	}

	for _, entry := range []ab{
		{`Name`, d.Name, "monax"},
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
[monax]
name       = [ "keys"]
`
	)

	if err := testutil.FakeDefinitionFile(config.MonaxRoot, "package", definition); err != nil {
		t.Fatalf("cannot place a definition file")
	}

	if _, err := LoadPackage(name); err == nil {
		t.Fatalf("expected definition fail to load")
	}
}
*/

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

	if err := testutil.FakeDefinitionFile(config.ServicesPath, name, definition); err != nil {
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

		{`Labels["MONAX"]`, d.Operations.Labels[definitions.LabelMonax], "true"},
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

	if err := testutil.FakeDefinitionFile(config.ServicesPath, name, definition); err != nil {
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

		{`Labels["MONAX"]`, d.Operations.Labels[definitions.LabelMonax], "true"},
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

	if err := testutil.FakeDefinitionFile(config.ServicesPath, name, ``); err != nil {
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

	os.Remove(filepath.Join(config.ServicesPath, name+".toml"))

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

	if err := testutil.FakeDefinitionFile(config.ServicesPath, name, definition); err != nil {
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

		{`Labels["MONAX"]`, d.Operations.Labels[definitions.LabelMonax], "true"},
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
