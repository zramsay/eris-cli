|[![GoDoc](https://godoc.org/github.com/eris-ltd/eris-cli/cmd/eris?status.png)](https://godoc.org/github.com/eris-ltd/eris-cli/cmd/eris) | Linux | OSX | Windows |
|---|-------|-----|---------|
| Master | [![Linux](https://circleci.com/gh/eris-ltd/eris-cli/tree/master.svg?style=svg)](https://circleci.com/gh/eris-ltd/eris-cli) | [![OSX](https://travis-ci.org/eris-ltd/eris-cli.svg?branch=master)](https://travis-ci.org/eris-ltd/eris-cli) | [![Windows](https://ci.appveyor.com/api/projects/status/lfkvvy6h7u0owv19/branch/master?svg=true)](https://ci.appveyor.com/project/eris-ltd/eris-cli) |
| Develop | [![Linux](https://circleci.com/gh/eris-ltd/eris-cli/tree/develop.svg?style=svg)](https://circleci.com/gh/eris-ltd/eris-cli) | [![OSX](https://travis-ci.org/eris-ltd/eris-cli.svg?branch=develop)](https://travis-ci.org/eris-ltd/eris-cli) | [![Windows](https://ci.appveyor.com/api/projects/status/lfkvvy6h7u0owv19/branch/develop?svg=true)](https://ci.appveyor.com/project/eris-ltd/eris-cli) |

# Introduction

```
Eris is a platform for building, testing, maintaining, and operating distributed
applications with a blockchain backend. Eris makes it easy and simple to wrangle
the dragons of smart contract blockchains.
```

`eris:cli` is a tool which makes it easy for developers to build, test, manage, and operate smart contract applications. **No matter the blockchain**.

[For the motivation behind this tool see this post](https://docs.erisindustries.com/documentation/eris-cli/latest/motivation/).

# Install (For Developers)

* Install Docker.
* Install Go.

```
go get github.com/eris-ltd/eris-cli/cmd/eris
eris init
```

See below for the directory structure created by `init`.

# Install (For Non-Developers)

Please see our [getting started page](https://docs.erisindustries.com/tutorials/getting-started/) for those who are not familiar with go and/or docker.

# Overview

The `eris` tool is centered around a very few concepts:

* `services` -- things that you turn on or off
* `chains` -- develop permissioned chains
* `pkgs` -- our smart contract tool chain
* `keys` -- wrapping of our key management tooling
* `actions` -- step by step processes
* `files` -- working the IPFS "permanent web"
* `data` -- take the pain out of data persistence on docker

These concepts provide the core functionality of what we think a true smart contract application platform requires.

To get started using the `eris` platform to see what the tooling can do and how it can help your development patterns for smart contract applications, please see our [getting started tutorial series](https://docs.erisindustries.com/tutorials/).

# Architecture of the Tool

`eris:cli` is mostly an opinionated wrapper around Docker's API. We have found that running applications locally which require sophisticated installation paths and/or complex configuration work best when used from Docker's Container based system.

Each of the `concepts` listed above is described in a bit more detail below.

## Services

Services are "things that you turn on or off". Examples of services include:

* a pgp daemon
* an ipfs node
* a bitcoin node
* a bitcoin node with its rpc on
* a bitcoin-testnet node with its rpc on
* an ethereum node
* a tendermint-testchain node
* a tinydns daemon

Services work from a base of **service definition files**. These files are held on the host in the following location: `~/.eris/services`. Service definition files tell `eris` how a docker container should be started. The specification for service definition files is located [here](https://docs.erisindustries.com/documentation/eris-cli/latest/services_specification/).

To see the various ways in which `eris` can interact with services, please type:

```
eris services
```

## Chains

Chains are an opinionated toolchain around permissioned chains. They can be most easily thought of as your "develop" branch for chains. In other words, if you need to work **on** a permissioned chain, then it is best to use `eris chains`. Chains hardcode most of the service starting criteria, but still allow for some flexibility as to how chains are worked with.

Chains are operated from a base of **chain definition files**. These files are held on the host in the following location: `~/.eris/chains`. The specification for chain definition files is located [here](https://docs.erisindustries.com/documentation/eris-cli/latest/chains_specification/).

To see the various ways in which `eris` can help you develop chains, please type:

```
eris chains
```

## Pkgs

Pkgs are an opinionated toolkit to help you deploy and test your smart contract packages on both permissioned and unpermissioned blockchain networks.

`eris pkgs` utilizes the [eris:package_manager](https://docs.erisindustries.com/documentation/epm/) to deal with contracts. `eris:package_manager` is a yaml based automation framework which makes it trivial to deploy and test your smart contract systems. The specification for `eris:package_manager` definition files is located [here](https://docs.erisindustries.com/documentation/eris-pm/latest/jobs_specification/).

Pkgs give you access to test your smart contracts both against "throwaway chains" which are one time use chains that are needed for the sole purpose of testing smart contract packages, as well as existing blockchain networks.

To see the various ways in which `eris` can help you develop smart contract applications, please type:

```
eris pkgs
```

## Keys

Keys is an opinionated toolchain around [eris:keys](https://docs.erisindustries.com/documentation/ekeys/). Please note that this concept of the `eris` platform is **for development only** and should not be used in production because it has not been fully security audited **and we do not plan for it to be**. In production the keys service should be replaced with your audited security system of choice.

To see the various ways in which `eris` can help you manage your various key pairs, please type:

```
eris keys
```

## Actions

Actions are step by step processes which need to take a few variables but otherwise should be scriptable. Actions are used for repetitive, or repeatable, tasks.

The environment in which actions run is similar in nature to a modern continuous development environment. Actions are run on the **host** and **not** from within a container. They have full access to containers either via the `eris` cli or via docker's cli.

Examples of things actions are made to support:

* setting up an application
* configuring a range of services
* register a domain entry via mindy
* drop a preformulated transaction into the btc network using a specific key

Actions work from a base of **action definition files**. These files are held on the host in the following location: `~/.eris/actions`. Action definition files tell `eris` what steps to take, what services to make available, what chain to run, and what steps to take. The specification for action definition files is located [here](https://docs.erisindustries.com/documentation/eris-cli/latest/actions_specification/).

To see the various ways in which `eris` can interact with actions, please type:

```
eris actions
```

## Files

Eris has a pretty handy wrapper around IPFS which is useful for quick file sharing from the host or for use by actions.

To see the various ways in which `eris` can help you with distributed file sharing, please type:

```
eris files
```

## Data

Eris can automagically utilize data containers for you.

If you turn the `data_container` variable to `true` in the service or chain definition file, then `eris` deposit the data utilized by that service or chain into a data container which can be managed separately from the "program" container. The advantage of working with data containers has been dealt with elsewhere (see, Google).

To see the various ways in which `eris` can help you manage your data containers, please type:

```
eris data
```

## Directory Structure

Created by `eris init` in $HOME directory:

```
├── .eris/
│   ├── eris.toml
│   ├── actions/
│   ├── apps/
│   ├── bundles/
│   ├── chains/
│       ├── default/config.toml
│   ├── keys/
│       ├── data/
│   ├── remotes/
│   ├── scratch/
│       ├── data/
│       ├── languages/
│       ├── lllc/
│       ├── ser/
│       ├── sol/
│   ├── services/
│       ├── global/
│       ├── btcd.toml
│       ├── ipfs.toml
│       ├── keys.toml
```

With several more default services also pulled in.

# Contributions

Are Welcome! Before submitting a pull request please:

* read up on [How The Marmots Git](https://github.com/eris-ltd/coding/wiki/How-The-Marmots-Git)
* fork from `develop`
* go fmt your changes
* have tests
* pull request
* be awesome

That's pretty much it. 

See our [CONTRIBUTING.md](.github/CONTRIBUTING.md) and [PULL_REQUEST_TEMPLATE.md](.github/PULL_REQUEST_TEMPLATE.md) for more details.

Please note that this repository is GPLv3.0 per the LICENSE file. Any code which is contributed via pull request shall be deemed to have consented to GPLv3.0 via submission of the code (were such code accepted into the repository).

# Bug Reporting

Found a bug in our stack? Make an issue!

The [issue template](.github/ISSUE_TEMPLATE.md] specifies what needs to be included in your issue and will autopopulate the issue.

# License

[Proudly GPL-3](http://www.gnu.org/philosophy/enforcing-gpl.en.html). See [license file](https://github.com/eris-ltd/eris-cli/blob/master/LICENSE.md).
