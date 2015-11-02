[![Circle CI](https://circleci.com/gh/eris-ltd/eris-cli/tree/master.svg?style=svg)](https://circleci.com/gh/eris-ltd/eris-cli)

[![GoDoc](https://godoc.org/github.com/eris-ltd/eris-cli/cmd/eris?status.png)](https://godoc.org/github.com/eris-ltd/eris-cli/cmd/eris)

```
Eris is a platform for building, testing, maintaining, and operating distributed
applications with a blockchain backend. Eris makes it easy and simple to wrangle
the dragons of smart contract blockchains.
```

`eris` is a tool which makes it easy for developers to build, test, manage, and operate distributed applications.

**No matter the blockchain**.

[For the motivation behind this tool](https://github.com/eris-ltd/eris-cli/blob/master/docs/motivation.md).

# Install

* Install Docker.
* Install Go.

```
go get github.com/eris-ltd/eris-cli/cmd/eris
eris init
```

Please see our [getting started page](https://docs.erisindustries.com/tutorials/getting-started/) for those who no familiar with go and/or docker.

# Eris: Overview

The `eris` tool is centered around a very few concepts:

* `services` -- things that you turn on or off
* `chains` -- develop permissioned chains
* `actions` -- step by step processes
* `contracts` -- the newest iteration of our smart contract tool chain

These concepts (along with a few other goodies) provide the core functionality of what we think a true distributed application would look like.

# Architecture of the Tool

The Eris CLI tool is mostly an opinionated wrapper around Docker's API. We have found that running applications locally which require sophisticated installation paths and/or complex configuration work best when used from Docker's Container based system. Each of the `concepts` listed above is described in a bit more detail below.

## Services

Services are "things that you turn on or off". Examples of services include:

* a pgp daemon
* an ipfs node
* a bitcoin node
* a bitcoin node with its rpc on
* a bitcoin-testnet node with its rpc on
* an ethereum-frontier node
* a counterparty node
* a ripple server or gateway
* a tendermint-testchain node
* a tinydns daemon

Services work from a base of **service definition files**. These files are held on the host in the following location: `~/.eris/services`. Service definition files tell `eris` how a docker container should be started. The specification for service definition files is located [here](docs/services_specification.md).

To see the various ways in which `eris` can interact with services, please type:

```
eris services
```

## Chains

Chains are an opinionated toolchain around permissioned tendermint blockchains. They can be most easily thought of as your "develop" branch for blockchains. In other words, if you need to work **on** a permissioned tendermint blockchain, then it is best to use `eris chains`.

Chains hardcode most of the service starting criteria, but still allow for some flexibility as to how chains are worked with. Chains are operated from a base of **chain definition files**. These files are held on the host in the following location: `~/.eris/chains`. The specification for chain definition files is located [here](docs/chains_specification.md).

To see the various ways in which `eris` can help you develop blockchains, please type:

```
eris chains
```

## Actions

Actions are step by step processes which need to take a few variables but otherwise should be scriptable. Actions are used for repetitive, or repeatable, tasks. The environment in which actions run is similar in nature to a modern continuous development environment. Actions are run on the **host** and from within a container. They have full access to containers either via the `eris` cli or via docker's cli.

Examples of things actions are made to support:

* register a domain entry via mindy
* drop a preformulated transaction into the btc network using a specific key

Actions work from a base of **action definition files**. These files are held on the host in the following location: `~/.eris/actions`. Action definition files tell `eris` what steps to take, what services to make available, what chain to run, and what steps to take. The specification for action definition files is located [here](docs/actions_specification.md).

To see the various ways in which `eris` can interact with actions, please type:

```
eris actions
```

## Data

Eris can automagically utilize [data containers](http://container42.com/2014/11/18/data-only-container-madness/) for you. If you turn the `data_container` variable to `true` in the service or chain definition file, then `eris` deposit "most" of the data utilized by that service or chain into a data container which can be managed separately from the "program" container. The advantage of working with data containers has been dealt with elsewhere.

To see the various ways in which `eris` can help you manage your data containers, please type:

```
eris data
```

## Files

Eris has a pretty handy wrapper around IPFS which is useful for quick file sharing from the host or for use by actions.

To see the various ways in which `eris` can help you with distributed file sharing, please type:

```
eris files
```

# Contributions

Are Welcome! Before submitting a pull request please:

* go fmt your changes
* have tests
* pull request
* be awesome

That's pretty much it (for now).

Please note that this repository is GPLv3.0 per the LICENSE file. Any code which is contributed via pull request shall be deemed to have consented to GPLv3.0 via submission of the code (were such code accepted into the repository).

# Bug Reporting

Found a bug in our stack? Make an issue!

Issues should contain four things:

* The operating system. Please be specific. Include the Docker version and, if applicable, which VM you are using (Toolbox/Kitematic/boot2docker).
* The reproduction steps. Starting from a fresh environment, what are all the steps that lead to the bug? Also include the branch you're working from.
* What you expected to happen. Provide a sample output.
* What actually happened. Error messages, logs, etc. Use `-d` to provide the most information. For lengthy outputs, link to a gist or pastebin please.

Finally, add a label to your bug (critical or minor). Critical bugs will likely be addressed quickly while minor ones may take awhile. Pull requests welcome for either, just let us know you're working on one in the issue (we use the in-progress label accordingly).

# License

[Proudly GPL-3](http://www.gnu.org/philosophy/enforcing-gpl.en.html). See [license file](https://github.com/eris-ltd/eris-cli/blob/master/LICENSE.md).
