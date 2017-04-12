**DEPRECATION WARNING:** As of the 0.17 release of the [Monax Platform](https://github.com/monax/cli) this repository will be deprecated and the functionality directly integrated into the platform.

# Compilers

The Compilers Service is a helper tool to help in grabbing necessary data such as binaries and ABIs from your preferred language for smart contracts in a simple manner. Currently that language is Solidity, but the service is easily extensible to other languages in the future.

## Table of Contents

- [Background](#background)
- [Installation](#installation)
- [Usage](#usage)
- [Contribute](#contribute)
- [License](#license)

## Background

A web server and client for compiling smart contract languages.

**Features:**
- compiles Solidity
- returns smart contract abis and binaries
- handles included files recursively with regex matching
- client side and server side caching
- configuration file with per-language options
- easily extensible to new languages

Monax Industries' own public facing compiler server (at https://compilers.monax.io) is hardcoded into the source,
so you can start compiling smart contract language right out of the box with no extra tools required.

## Installation

`monax-compilers` is intended to run as a service in a docker container via the [Monax CLI](https://monax.io/docs). The server can be started with: `monax services start compilers`.

### For Developers

1. [Install go](https://golang.org/doc/install)
3. `go get github.com/monax/compilers/cmd/monax-compilers`
2. (Optional) [Install Solidity](http://solidity.readthedocs.org/en/latest/installing-solidity.html)

## Usage

### As A Library

```
import (
  client "github.com/monax/compilers/network"
)

url := "https://compilers.monax.io:9099/compile"
filename := "maSolcFile.sol"
optimize := true
librariesString := "maLibrariez:0x1234567890"

output, err := client.BeginCompile(url, filename, optimize, librariesString)

contractName := output.Objects[0].Objectname // contract C would give you C here
binary := output.Objects[0].Bytecode // gives you the binary
abi := output.Objects[0].ABI // gives you the ABI
```

### Compile Remotely

```
monax-compilers compile test.sol
```

Will by default compile directly using the monax servers. You can configure this to call a different server by checking out the `--help` option.

### Compile Locally

Make sure you have the appropriate compiler installed and configured (you may need to adjust the `cmd` field in the config file)

```
monax-compilers compile --local test.sol
```

### Run a server yourself

```
monax-compilers server --no-ssl
```

will run a simple http server. For encryption, pass in a key with the `--key` flag, or a certificate with the `--cert` flag and drop the `--no-ssl`.

### Support

Run `monax-compilers server --help` or `monax-compilers compile --help` for more info, or come talk to us on [Slack](https://slack.monax.io).

If you are working on a language, and would like to have it supported, please create an issue!

## Contribute

See the [monax platform contributing file here](https://github.com/monax/cli/blob/master/.github/CONTRIBUTING.md).

## License

[GPL-3](license.md)
