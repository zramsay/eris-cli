---

type:   docs
layout: single
title: "Specifications | Jobs Specification"

---

## Jobs Specification

<div class="note">
{{% data_sites rename_docs %}}
</div>

The goal of jobs is to enable automation of contractual steps and defining "use" for a package of smart contracts. While one might deploy utility contracts that can be shared broadly and used by many a developer, jobs allows one to execute a package of contracts as it was intended and purposed to do and allows this execution to be replicated on thousands of chains, thus enabling what we call "dual integration" so that smart contracts can then be potentially represented as real life contracts. Rather than creating an entire web app, one can also simply test all of their smart contracts through a yaml config file in the form of the `epm.yaml` file. This enables quick deployment and testing of smart contract functionality. 

When combined with a package manager to install and resolve dependencies, one can create scripts on top of their smart contracts to coordinate and calculate how certain contracts will interact with each other and evaluate how they will interact through development externalities in real time. Utilizing the monax `job runner` one can make smart contracts as easy as shell scripting.

Examples of monax job definition files are available in the [jobs_fixtures](https://github.com/monax/monax/tree/master/tests/jobs_fixtures) directory.

Each job will perform its required action and then it will save the result of its job in a variable which can be utilized by jobs later in the sequence using the job runner' [variable specification](/docs/specs/variable_specification).

By default, `monax pkgs do` will perform the entire sequence of jobs which has been outlined in a given jobs file.

## Job Types

Jobs are performed as sequentially based on the order they are given in the jobs definition file. By default the job runner will perform the entire sequence of jobs which has been outlined in a given jobs file.

The CLI will look for an `epm.yaml` file in the current directory unless specified with the `-f` flag. Once it has that, it will from there marshal the values in the file into a job definition which when filled direct the job runner on how to execute a task.

Jobs are categorized into:

* [contracts jobs](#contracts-jobs)
* [transaction jobs](#transaction-jobs)
* [test jobs](#test-jobs)
* [other jobs](#other-jobs)

### Transaction Jobs

Transaction jobs handle everything that is non contract related that involves updating of state in a chain. The following jobs are available to be utilized:

* [send](#send-jobs): a transaction which sends tokens from one account to another
* [register](#register-jobs): register a name in the native chain name registry
* [permission](#permission-jobs): update an account's permissions or roles (must be sent from an account which has root permissions on the chain)

#### Send Jobs

The send job will send tokens from one address to another address.

Ex.

```yaml
jobs:

- name: sendTx
  job:
    send:
      destination: 58FD1799AA32DED3F6EAC096A1DC77834A446B9C
      amount: 1234
```

#### Register Jobs

The register job will register a name and data associated with it to the native chain name registry. This is useful for creating DNS systems on a chain. It also can come from a csv file. Note that a fee must be given to activate the name registry.

Ex.

```yaml
- name: nameReg
  job:
    register:
      name: oldGrayBeard
      data: "King of the Mountain"
      amount: 5000
      fee: 1234
```

#### Permission Jobs

The permission job can set the permissions on certain accounts using Burrow's secure native permissioning framework.
A detailed description of Monax permissions can be found [here](https://github.com/monax/burrow/blob/master/permission/types/permissions.go#L35). We expect to have a more complete spec available shortly. 

```yaml
- name: permTest1
  job:
    permission:
      action: setBase
      target: $addr2
      permission: call
      value: "true"

- name: permTest2
  job:
    permission:
      action: unsetBase
      target: $addr2
      permission: call

- name: permTest3
  job:
    permission:
      action: addRole
      target: $addr2
      role: 1234

- name: permTest4
  job:
    permission:
      action: removeRole
      target: $addr2
      role: 1234

- name: permTest5
  job:
    permission:
      action: setGlobal
      permission: call
      value: "true"
 ```


### Contracts Jobs

Contract jobs handle everything contract related. There are two jobs currently available:

* [deploy](#deploy-jobs): deploy a single contract
* [call](#call-jobs): send a transaction to a contract (can only be sent to existing contracts)

More jobs are currently planned for development with regard to contracts. Stay tuned.

#### Deploy Jobs

The `deploy` job will deploy a smart contract. Constructor values are accessed through the `data` key. If you have more than one object in your contract file, you can use the `instance` key to denote which instance you would like to deploy, or if you would like to deploy all of them, set instance to `all`. When deployed, the result will be the address of the contract you have deployed and will be referenceable. To link contract libraries, use the `libraries` key just as you would in [Solidity](https://solidity.readthedocs.io/en/latest/using-the-compiler.html#using-the-commandline-compiler). After deployment, an `abi` folder will present itself in your `pwd`. Use this to interact further with your smart contracts. One can also save the binary
of a contract by setting the `save` key to true. The artifacts of this save will present itself in a `bin` directory. You can also play with gas, fees, nonce, and tokens to send with the deploy as well using the fields `gas`, `fee`, `nonce`, and `amount` respectively. 

Ex.

```yaml
- name: deploySingleLib
  job:
    deploy:
      contract: single-lib.sol
      instance: Search

- name: deployConsumingContract
  job:
    deploy:
      contract: consuming-contract.sol
      libraries: Search:$deploySingleLib

- name: deployIntConstructor
  job:
    deploy:
      contract: contracts/storage.sol
      instance: SimpleConstructorInt
      data: 
        - $setIntStorage
        - 3
```

#### Call Jobs

The `call` job is the primary way for interacting with smart contracts resting on a chain. It takes a function name, destination address, arguments, and can take what ABI to use, and allows you to specify gas, value, nonce and fees just like in the deploy job.

Ex.

```yaml
- name: callWithIntArray
  job:
    call:
      destination: $deployC
      function: intCallWithArray 
      data: 
        - [1,2,3,4]

- name: setStorage
  job:
    call:
      destination: $deployStorageK
      function: set
      data:
        - $setStorageBase

- name: createGSContract
  job:
    call:
      destination: $deployGSFactory
      function: create
      abi: GSFactory
```

### Test Jobs

Test jobs query and test the state of the chain as it currently stands. The following jobs are available to be utilized:

* [query-account](#queryaccount-jobs): get information about an account on the blockchain
* [query-contract](#querycontract-jobs): perform a simulated call against a specific contract (usually used to trigger accessor functions which retrieve information from a return variable of a contract's function)
* [query-name](#queryname-jobs): get information about a registered name using burrow's name registry functionality
* [query-vals](#queryvals-jobs): get information about the validator set
* [assert](#assert-jobs): assert a relationship between a key and a value (useful for testing purposes to make sure everything deployed properly and/or is working as it should)

#### QueryAccount Jobs

The `query-account` job will query an account for its current balance or its permissions. 

Ex.

```yaml
- name: queryPerm1
  job:
    query-account:
      account: $addr2
      field: permissions.roles

- name: queryPerm2
  job:
    query-account:
      account: $addr2
      field: permissions.base

- name: sendTxQuery1
  job:
    query-account:
      account: $receipient
      field: balance
```

#### QueryContract Jobs

The `query-contract` job will query a contract for its current state based on a method input. It is very similar to a call but does not update state nor does it cost anything transactionally to make a query of the state.

Ex.

```yaml
- name: getLastAddr
  job:
    query-contract:
      destination: $deployGSFactory
      function: last
      data: ["some", "data", 4, "u"]
      abi: GSFactory
```

#### QueryName Jobs

The `query-name` job returns information about the name registry, usually data is what you'll be querying to get the name:

Ex.

```yaml
- name: queryReg2
  job:
    query-name:
      name: marmots
      field: data
```

#### QueryVals Jobs

The `query-vals` job allows you to see what current validators are contributing proof of stake to the chain.

Ex.

```yaml
- name: queryBonded
  job:
    query-vals:
      field: bonded_validators

- name: queryUnbonding
  job:
    query-vals:
      field: unbonding_validators
```

#### Assert Jobs

Asserts can be used to compare two "things". These "things" may be the result of two jobs or the result against one job against a baseline. (Indeed, it could be the comparison of two baselines but that wouldn't really get folks anywhere).

[Read the Assert Jobs Specification &nbsp;<i class="fa fa-chevron-circle-right" aria-hidden="true"></i>](/docs/specs/asserts_specification)

### Utility Jobs

Utility jobs are made for making your jobs definition file easier to work and write with and understand:

* [account](#account-jobs): set the account to use
* [set](#set-jobs): set the value of a variable

#### Account Jobs

An account job takes in an account address to be referenced later in transaction and contract jobs. It should also be noted that for now there is a default job written at the beginning of every job run.

Ex:

```yaml
jobs:

- name: val1
  job:
    account:
      address: 58FD1799AA32DED3F6EAC096A1DC77834A446B9C
```

#### Set Jobs

A set job takes any kind of value. This is usually used to reference as a key in an assertion job.

Ex:

```yaml
jobs:

- name: val1
  job:
    set:
      val: 1234
```

### Extension/Creating your own Job - Coming soon!


## [<i class="fa fa-chevron-circle-left" aria-hidden="true"></i> All Specifications](/docs/specs)
