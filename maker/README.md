# Eris Chain Manager

|[![GoDoc](https://godoc.org/github.com/eris-cm?status.png)](https://godoc.org/github.com/eris-ltd/eris-cm) | Linux |
|---|-------|
| Master | [![Circle CI](https://circleci.com/gh/eris-ltd/eris-cm/tree/master.svg?style=svg)](https://circleci.com/gh/eris-ltd/eris-cm/tree/master) |
| Develop | [![Circle CI](https://circleci.com/gh/eris-ltd/eris-cm/tree/develop.svg?style=svg)](https://circleci.com/gh/eris-ltd/eris-cm/tree/develop) |

The Eris Chain Manager is a utility for performing complex operations on `eris chains`. This command is exposed through [eris-cli](https://monax.io/docs/documentation/cli), the entry point for the Eris Platform.

## Table of Contents

- [Background](#background)
- [Installation](#installation)
- [Usage](#usage)
- [Contribute](#contribute)
- [License](#license)

## Background

`eris-cm` is a high level tool for working with `eris chains`. It is similar in nature, design, and level as the `eris-pm` which is built to handle smart contract packages and other packages necessary for building blockchain backed applications. It is used to provide a harmonized interface to the modular components of the [eris](https://monax.io/docs/documentation) open source platform.

## Installation

`eris-cm` is intended to be used by the `eris chains` command via [eris-cli](https://monax.io/docs/documentation/cli/latest/eris_chains/).

### For Developers
Should you want/desire/need to install this repository natively on your host make sure you have go installed and then:

1. [Install go](https://golang.org/doc/install)
2. Ensure you have `gmp` installed (`sudo apt-get install libgmp3-dev || brew install gmp`)
3. `go get github.com/eris-ltd/eris-cm/cmd/eris-cm`

## Usage

```
The Eris Chain Manager is a utility for performing complex operations on eris chains.

Made with <3 by Monax Industries.

Complete documentation is available at https://monax.io/docs/documentation/

Version:
  0.12.0

Usage:
  eris-cm [flags]
  eris-cm [command]

Available Commands:
  make        The Eris Chain Maker is a utility for easily creating the files necessary to build eris chains

Flags:
  -d, --debug[=false]: debug level output; the most output available for eris-cm; if it is too chatty use verbose flag; default respects $ERIS_CHAINMANAGER_DEBUG
  -h, --help[=false]: help for eris-cm
  -o, --output[=true]: should eris-cm provide an output of its job; default respects $ERIS_CHAINMANAGER_OUTPUT
  -v, --verbose[=false]: verbose output; more output than no output flags; less output than debug level; default respects $ERIS_CHAINMANAGER_VERBOSE
```

or

```
The Eris Chain Maker is a utility for easily creating the files necessary to build eris chains.

Usage:
  eris-cm make [flags]

Examples:
$ eris-cm make myChain -- will use the chain-making wizard and make your chain named myChain using eris-keys defaults (available via localhost) (interactive)
$ eris-cm make myChain --chain-type=simplechain --  will use the chain type definition files to make your chain named myChain using eris-keys defaults (non-interactive)
$ eris-cm make myChain --account-types=Root:1,Developer:0,Validator:0,Participant:1 -- will use the flag to make your chain named myChain using eris-keys defaults (non-interactive)
$ eris-cm make myChain --account-types=Root:1,Developer:0,Validator:0,Participant:1 --chain-type=simplechain -- account types trump chain types, this command will use the flags to make the chain (non-interactive)
$ eris-cm make myChain --csv /path/to/csv -- will use the csv file to make your chain named myChain using eris-keys defaults (non-interactive)

Flags:
  -t, --account-types=[]: what number of account types should we use? find these in ~/.eris/chains/account-types; incompatible with and overrides chain-type; default respects $ERIS_CHAINMANAGER_ACCOUNTTYPES
  -c, --chain-type="": which chain type definition should we use? find these in ~/.eris/chains/chain-types; default respects $ERIS_CHAINMANAGER_CHAINTYPE
  -s, --csv-file="": csv file in the form `account-type,number,tokens,toBond,perms; default respects $ERIS_CHAINMANAGER_CSVFILE
  -h, --help[=false]: help for make
  -k, --keys-server="http://localhost:4767": keys server which should be used to generate keys; default respects $ERIS_KEYS_PATH
  -r, --tar[=false]: instead of making directories in ~/.chains, make tarballs; incompatible with and overrides zip; default respects $ERIS_CHAINMANAGER_TARBALLS
  -z, --zip[=false]: instead of making directories in ~/.chains, make zip files; default respects $ERIS_CHAINMANAGER_ZIPFILES

Global Flags:
  -d, --debug[=false]: debug level output; the most output available for eris-cm; if it is too chatty use verbose flag; default respects $ERIS_CHAINMANAGER_DEBUG
  -o, --output[=true]: should eris-cm provide an output of its job; default respects $ERIS_CHAINMANAGER_OUTPUT
  -v, --verbose[=false]: verbose output; more output than no output flags; less output than debug level; default respects $ERIS_CHAINMANAGER_VERBOSE
```

## Contribute

See the [eris platform contributing file here](https://github.com/eris-ltd/coding/blob/master/github/CONTRIBUTING.md).

## License

[GPL-3](LICENSE)
