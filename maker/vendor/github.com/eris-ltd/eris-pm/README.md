# Eris Package Manager

|[![GoDoc](https://godoc.org/github.com/eris-pm?status.png)](https://godoc.org/github.com/eris-ltd/eris-pm) | Linux |
|---|-------|
| Master | [![Circle CI](https://circleci.com/gh/eris-ltd/eris-pm/tree/master.svg?style=svg)](https://circleci.com/gh/eris-ltd/eris-pm/tree/master) |
| Develop | [![Circle CI](https://circleci.com/gh/eris-ltd/eris-pm/tree/develop.svg?style=svg)](https://circleci.com/gh/eris-ltd/eris-pm/tree/develop) |


The Eris Package Manager is a utility for deploying and testing smart contract packages. It is most commonly used by `eris pkgs do`, a command exposed through [eris-cli](https://monax.io/docs/documentation/cli), the entry point for the Eris Platform.

## Table of Contents

- [Background](#background)
- [Installation](#installation)
- [Usage](#usage)
  - [EPM Definition Files](#epm-definition-files)
- [Contribute](#contribute)
- [License](#license)

## Background

`epm` is a high level tool which provides easy access to most of the `eris-db` tooling. `epm` is used to deploy and test suites of smart contracts. In general it wraps the `eris-client` tooling, along with `eris-keys` and `eris-compilers` to provide a harmonized interface to the modular components of the [eris](https://monax.io/docs/documentation/) open source platform.

`epm` is closer to an ansible or chef like tool than it is `npm` in that it is a deployment sequence and testing tool. `epm` uses an **epm definition file** to tell the package manager what jobs should be ran and in what order.

In EPM a *job* is a single action which is performed (such as a transaction, a contract deployment, a call to a smart contract, or a query of information). The results of these jobs are then kept in variables and may be used in later jobs.

## Installation

`epm` is intended to be used by the `eris pkgs do` command vi [eris-cli](https://github.com/eris-ltd/eris-cli), which runs a temporary docker container to expose its functionality.

### For Developers

1. [Install go](https://golang.org/doc/install)
2. [Install glide](https://github.com/Masterminds/glide)
3. Ensure you have `gmp` installed (`sudo apt-get install libgmp3-dev || brew install gmp`)
4. `go get github.com/eris-ltd/eris-pm/cmd/epm`
5. `glide install -s -v`

## Usage

```
The Eris Package Manager Deploys and Tests Smart Contract Systems

Made with <3 by Monax Industries.

Complete documentation is available at https://monax.io/docs/documentation/

Version:
  0.12.0

Usage:
  epm [flags]

Flags:
  -a, --abi-path="./abi": path to the abi directory EPM should use when saving ABIs after the compile process; default respects $EPM_ABI_PATH
  -r, --address="": default address to use; operates the same way as the [account] job, only before the epm file is ran; default respects $EPM_ADDRESS
  -u, --amount="9999": default amount to use; default respects $EPM_AMOUNT
  -c, --chain="localhost:46657": <ip:port> of chain which EPM should use; default respects $EPM_CHAIN_ADDR
  -m, --compiler="https://compilers.monax.io:10114": <ip:port> of compiler which EPM should use; default respects $EPM_COMPILER_ADDR
  -p, --contracts-path="./contracts": path to the contracts EPM should use; default respects $EPM_CONTRACTS_PATH
  -d, --debug=false: debug level output; the most output available for epm; if it is too chatty use verbose flag; default respects $EPM_DEBUG
  -n, --fee="1234": default fee to use; default respects $EPM_FEE
  -f, --file="./epm.yaml": path to package file which EPM should use; default respects $EPM_FILE
  -g, --gas="1111111111": default gas to use; can be overridden for any single job; default respects $EPM_GAS
  -h, --help=false: help for epm
  -o, --output="csv": output format which epm should use [csv,json]; default respects $EPM_OUTPUT_FORMAT
  -e, --set=: default sets to use; operates the same way as the [set] jobs, only before the epm file is ran (and after default address; default respects $EPM_SETS
  -s, --sign="localhost:4767": <ip:port> of signer daemon which EPM should use; default respects $EPM_SIGNER_ADDR
  -t, --summary=true: output a table summarizing epm jobs; default respects $EPM_SUMMARY_TABLE
  -v, --verbose=false: verbose output; more output than no output flags; less output than debug level; default respects $EPM_VERBOSE
```

`epm` is a simple tool from the command line perspective in that it does not have subcommands. `epm` is the only command it will run. This command will execute the instructions of the epm definition file in the current directory (unless a different file is given via the `--file` flag or `$EPM_FILE` environment variable).

### EPM Definition Files

A sample EPM definition file, typically saved as `epm.yaml`, looks like: 

```yaml
jobs:
- name: account1
  job:
    account:
      address: 1040E6521541DAB4E7EE57F21226DD17CE9F0FB7

- name: val1
  job:
    set:
      val: 1234

- name: sendTxTest1
  job:
    send:
      source: $account1
      destination: 58FD1799AA32DED3F6EAC096A1DC77834A446B9C
      amount: $val1
      wait: true

- name: val1
  job:
    set:
      val: "eris_loves"

- name: val2
  job:
    set:
      val: "marmots"

- name: MinersFee
  job:
    set:
      val: 1234

- name: nameRegTest1
  job:
    register:
      name: $val1
      data: $val2
      fee: $MinersFee
      wait: true

- name: account_tgt
  job:
    set:
      val: 58FD1799AA32DED3F6EAC096A1DC77834A446B9C

- name: perm
  job:
    set:
      val: call

- name: permTest2
  job:
    permission:
      action: unset_base
      target: $account_tgt
      permission: $perm
      wait: false

- name: setStorage
  job:
    set:
      val: 0x5

- name: deployStorageK
  job:
    deploy:
      contract: storage.sol
      wait: true

- name: setStorage
  job:
    call:
      destination: $deployStorageK
      data: set $setStorage
      wait: true
```

For more about the jobs epm is capable of performing please see the [Jobs Specification](https://monax.io/docs/documentation/pm/latest/specifications/jobs_specification/).

### Variable Handling

`epm` will also handle variables; for more information please see the [Variables Specification](https://monax.io/docs/documentation/pm/latest/specifications/variable_specification/).

`epm` will save an `epm.log` file with the variables used and results of the jobs in the `pwd` unless another location is specified.

## Contribute

See the [eris platform contributing file here](https://github.com/eris-ltd/coding/blob/master/github/CONTRIBUTING.md).

## License

[GPL-3](LICENSE)
