package loaders

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/monax/cli/config"
	"github.com/monax/cli/definitions"
	"github.com/monax/cli/log"
	"github.com/monax/cli/testutil"
	"github.com/monax/cli/util"
	"github.com/monax/cli/version"
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

		defaultDefinition = ``
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

	if err := testutil.FakeDefinitionFile(config.ChainsPath, "default", ``); err != nil {
		t.Fatalf("cannot place a default definition file")
	}
	if err := testutil.FakeDefinitionFile(filepath.Join(config.ChainsPath, name), name, definition); err != nil {
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

		{`Labels["MONAX"]`, s.Operations.Labels[definitions.LabelMonax], "true"},
		{`Labels["NAME"]`, s.Operations.Labels[definitions.LabelShortName], name},
		{`Labels["TYPE"]`, s.Operations.Labels[definitions.LabelType], definitions.TypeChain},

		{`Service.Name`, s.Service.Name, name},
		{`Service.AutoData`, s.Service.AutoData, true},
		// [pv]: not "test image", but monaxdb image. A bug?
		{`Service.Image`, s.Service.Image, path.Join(version.DefaultRegistry, version.ImageDB)},
		{`Service.Environment`, s.Service.Environment, []string{chainID, chainName}},
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

func TestLoadUtilJobsSimple(t *testing.T) {
	const (
		filename = "./epm.yaml"
		jobs     = `
jobs:
- name: setStorageBase
  set:
    val: 5
- name: setAccount
  account:
    address: 1234567890
`
	)
	err := ioutil.WriteFile(filename, []byte(jobs), 0644)
	defer os.Remove(filename)
	if err != nil {
		t.Fatalf("cannot write config file %v", err)
	}
	do := definitions.NowDo()
	do.YAMLPath = filename
	output, err := LoadJobs(do)
	if err != nil {
		t.Fatalf("could not load jobs: %v", err)
	}
	for _, entry := range []ab{
		{`SetName`, output.Jobs[0].Name, "setStorageBase"},
		{`SetVal`, output.Jobs[0].Set.Value, 5},
		{`AccountName`, output.Jobs[1].Name, "setAccount"},
		{`AccountVal`, output.Jobs[1].Account.Address, "1234567890"},
	} {
		if !reflect.DeepEqual(entry.a, entry.b) {
			t.Fatalf("definition expected %s = %#v, got %#v", entry.name, entry.b, entry.a)
		}
	}
}

func TestLoadContractJobsSimple(t *testing.T) {
	const (
		filename = "./epm.yaml"
		jobs     = `
jobs:
- name: deploySomething
  deploy:
    source: 1234567890
    contract: storage.sol
    instance: C
    libraries: ["someLib:0x1234567890","anotherLib:0x1234567890"]
    data: [1, 2, 3]
    amount: 1
    fee: 1
    gas: 1
    nonce: 2
- name: callSomething
  call:
    source: 1234567890
    destination: $deploySomething
    function: someFunc
    data: [1, 2, 3]
    amount: 1
    fee: 1
    gas: 1
    nonce: 2
`
	)
	err := ioutil.WriteFile(filename, []byte(jobs), 0644)
	defer os.Remove(filename)
	if err != nil {
		t.Fatalf("cannot write config file %v", err)
	}
	do := definitions.NowDo()
	do.YAMLPath = filename
	output, err := LoadJobs(do)
	if err != nil {
		t.Fatalf("could not load jobs: %v", err)
	}
	for _, entry := range []ab{
		{`DeployName`, output.Jobs[0].Name, "deploySomething"},
		{`DeploySource`, output.Jobs[0].Deploy.Source, "1234567890"},
		{`DeployContract`, output.Jobs[0].Deploy.Contract, "storage.sol"},
		{`DeployInstance`, output.Jobs[0].Deploy.Instance, "C"},
		{`DeployLibs`, output.Jobs[0].Deploy.Libraries, []string{"someLib:0x1234567890", "anotherLib:0x1234567890"}},
		{`DeployData`, output.Jobs[0].Deploy.Data, []interface{}{1, 2, 3}},
		{`DeployAmount`, output.Jobs[0].Deploy.Amount, "1"},
		{`DeployFee`, output.Jobs[0].Deploy.Fee, "1"},
		{`DeployGas`, output.Jobs[0].Deploy.Gas, "1"},
		{`DeployNonce`, output.Jobs[0].Deploy.Nonce, "2"},
		{`CallName`, output.Jobs[1].Name, "callSomething"},
		{`CallSource`, output.Jobs[1].Call.Source, "1234567890"},
		{`CallDestination`, output.Jobs[1].Call.Destination, "$deploySomething"},
		{`CallFunction`, output.Jobs[1].Call.Function, "someFunc"},
		{`CallData`, output.Jobs[1].Call.Data, []interface{}{1, 2, 3}},
		{`CallAmount`, output.Jobs[1].Call.Amount, "1"},
		{`CallFee`, output.Jobs[1].Call.Fee, "1"},
		{`CallGas`, output.Jobs[1].Call.Gas, "1"},
		{`CallNonce`, output.Jobs[1].Call.Nonce, "2"},
	} {
		if !reflect.DeepEqual(entry.a, entry.b) {
			t.Fatalf("definition expected %s = %#v, got %#v", entry.name, entry.b, entry.a)
		}
	}
}

func TestLoadTestJobsSimple(t *testing.T) {
	const (
		filename = "./epm.yaml"
		jobs     = `
jobs:
- name: querySomething
  query-contract:
    source: 1234567890
    destination: $deploySomething
    function: someFunc
    data: [1, 2, 3]
- name: queryAnAccount
  query-account:
    account: 1234567890
    field: permissions.base
- name: queryAName
  query-name:
    name: fred
    field: data
- name: queryValidators
  query-vals:
    field: bonded_validators
- name: assertSomething
  assert:
    key: someVal
    relation: eq 
    val: anotherVal
`
	)
	err := ioutil.WriteFile(filename, []byte(jobs), 0644)
	defer os.Remove(filename)
	if err != nil {
		t.Fatalf("cannot write config file %v", err)
	}
	do := definitions.NowDo()
	do.YAMLPath = filename
	output, err := LoadJobs(do)
	if err != nil {
		t.Fatalf("could not load jobs: %v", err)
	}
	for _, entry := range []ab{
		{`QueryContractName`, output.Jobs[0].Name, "querySomething"},
		{`QueryContractSource`, output.Jobs[0].QueryContract.Source, "1234567890"},
		{`QueryContractDestination`, output.Jobs[0].QueryContract.Destination, "$deploySomething"},
		{`QueryContractFunction`, output.Jobs[0].QueryContract.Function, "someFunc"},
		{`QueryContractData`, output.Jobs[0].QueryContract.Data, []interface{}{1, 2, 3}},
		{`QueryAccountName`, output.Jobs[1].Name, "queryAnAccount"},
		{`QueryAccountAccount`, output.Jobs[1].QueryAccount.Account, "1234567890"},
		{`QueryAccountField`, output.Jobs[1].QueryAccount.Field, "permissions.base"},
		{`QueryNameJobName`, output.Jobs[2].Name, "queryAName"},
		{`QueryNameName`, output.Jobs[2].QueryName.Name, "fred"},
		{`QueryNameField`, output.Jobs[2].QueryName.Field, "data"},
		{`QueryValsName`, output.Jobs[3].Name, "queryValidators"},
		{`QueryValsField`, output.Jobs[3].QueryVals.Field, "bonded_validators"},
		{`AssertName`, output.Jobs[4].Name, "assertSomething"},
		{`AssertSource`, output.Jobs[4].Assert.Key, "someVal"},
		{`AssertDestination`, output.Jobs[4].Assert.Relation, "eq"},
		{`AssertFunction`, output.Jobs[4].Assert.Value, "anotherVal"},
	} {
		if !reflect.DeepEqual(entry.a, entry.b) {
			t.Fatalf("definition expected %s = %#v, got %#v", entry.name, entry.b, entry.a)
		}
	}
}

func TestLoadTransactJobsSimple(t *testing.T) {
	const (
		filename = "./epm.yaml"
		jobs     = `
jobs:
- name: sendSomething
  send:
    source: 1234567890
    destination: $deploySomething
    amount: 1
    nonce: 3
- name: regName
  register:
    source: 1234567890
    name: fred
    data: someData
    data_file: something.csv
    amount: 1
    fee: 2
    nonce: 3
- name: updatePerms
  permission:
    source: 1234567890
    action: set_base
    permission: call
    value: "true"
    target: 1234567890
    role: 1234
    nonce: 3
- name: bondVal
  bond:
    pub_key: 1234567890
    account: 1234567890
    amount: 1
    nonce: 3
- name: unbondVal
  unbond:
    account: 1234567890
    height: $block
- name: rebondVal
  rebond:
    account: 1234567890
    height: $block
`
	)
	err := ioutil.WriteFile(filename, []byte(jobs), 0644)
	defer os.Remove(filename)
	if err != nil {
		t.Fatalf("cannot write config file %v", err)
	}
	do := definitions.NowDo()
	do.YAMLPath = filename
	output, err := LoadJobs(do)
	if err != nil {
		t.Fatalf("could not load jobs: %v", err)
	}
	for _, entry := range []ab{
		{`SendName`, output.Jobs[0].Name, "sendSomething"},
		{`SendSource`, output.Jobs[0].Send.Source, "1234567890"},
		{`SendDestination`, output.Jobs[0].Send.Destination, "$deploySomething"},
		{`SendAmount`, output.Jobs[0].Send.Amount, "1"},
		{`SendNonce`, output.Jobs[0].Send.Nonce, "3"},
		{`RegisterNameJobName`, output.Jobs[1].Name, "regName"},
		{`RegisterNameName`, output.Jobs[1].RegisterName.Name, "fred"},
		{`RegisterNameSource`, output.Jobs[1].RegisterName.Source, "1234567890"},
		{`RegisterNameData`, output.Jobs[1].RegisterName.Data, "someData"},
		{`RegisterNameDataFile`, output.Jobs[1].RegisterName.DataFile, "something.csv"},
		{`RegisterNameAmount`, output.Jobs[1].RegisterName.Amount, "1"},
		{`RegisterNameFee`, output.Jobs[1].RegisterName.Fee, "2"},
		{`RegisterNameNonce`, output.Jobs[1].RegisterName.Nonce, "3"},
		{`PermissionName`, output.Jobs[2].Name, "updatePerms"},
		{`PermissionAction`, output.Jobs[2].Permission.Action, "set_base"},
		{`PermissionSource`, output.Jobs[2].Permission.Source, "1234567890"},
		{`PermissionPermissionFlag`, output.Jobs[2].Permission.PermissionFlag, "call"},
		{`PermissionValue`, output.Jobs[2].Permission.Value, "true"},
		{`PermissionTarget`, output.Jobs[2].Permission.Target, "1234567890"},
		{`PermissionRole`, output.Jobs[2].Permission.Role, "1234"},
		{`PermissionNonce`, output.Jobs[2].Permission.Nonce, "3"},
		{`BondName`, output.Jobs[3].Name, "bondVal"},
		{`BondPubKey`, output.Jobs[3].Bond.PublicKey, "1234567890"},
		{`BondAccount`, output.Jobs[3].Bond.Account, "1234567890"},
		{`BondNonce`, output.Jobs[3].Bond.Nonce, "3"},
		{`BondAmount`, output.Jobs[3].Bond.Amount, "1"},
		{`UnbondName`, output.Jobs[4].Name, "unbondVal"},
		{`UnbondHeight`, output.Jobs[4].Unbond.Height, "$block"},
		{`UnbondAccount`, output.Jobs[4].Unbond.Account, "1234567890"},
		{`RebondName`, output.Jobs[5].Name, "rebondVal"},
		{`RebondHeight`, output.Jobs[5].Rebond.Height, "$block"},
		{`RebondAccount`, output.Jobs[5].Rebond.Account, "1234567890"},
	} {
		if !reflect.DeepEqual(entry.a, entry.b) {
			t.Fatalf("definition expected %s = %#v, got %#v", entry.name, entry.b, entry.a)
		}
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
