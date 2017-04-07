# Introduction

```
Monax is an application platform for building, testing, maintaining, and
operating applications built to run on an ecosystem level.
```

`monax` is a tool which makes it easy for developers to build, test, manage, and operate smart contract applications. **No matter the blockchain**.

[For the motivation behind this tool see this post](https://monax.io/platform/motivation).

# Install (For Developers)

* Install Docker.
* Install Go.

```
go get github.com/monax/cli/cmd/monax
monax init
```

See below for the directory structure created by `init`.

# Install (For Non-Developers)

Please see our [getting started page](https://monax.io/docs/getting-started) for those who are not familiar with go and/or docker.

# Overview

The `monax` tool is centered around a very few concepts:

* `services` — things that you turn on or off
* `chains` — develop permissioned chains
* `pkgs` — our smart contract tool chain
* `keys` — wrapping of our key management tooling
* `files` — working the IPFS "permanent web"
* `data` — take the pain out of data persistence on docker

These concepts provide the core functionality of what we think a true smart contract application platform requires.

To get started using `monax` to see what the tooling can do and how it can help your development patterns for smart contract applications, please see our [tutorial series](https://monax.io/docs).

# Architecture of the Tool

`monax` is mostly an opinionated wrapper around Docker's API. We have found that running applications locally which require sophisticated installation paths and/or complex configuration work best when used from Docker's container based system.

Each of the `concepts` listed above is described in a bit more detail below.

## Services

Services are "things that you turn on or off". Examples of services include:

* a PGP daemon
* an IPFS node
* a Bitcoin node
* an Ethereum node
* a Tendermint test chain node
* BigchainDB service
* ZCash node

Services work from a base of **service definition files**. These files are held on the host in the following location: `~/.monax/services`. Service definition files tell `monax` how a docker container should be started. The specification for service definition files is located [here](https://monax.io/docs/specs/services_specification).

To see the various ways in which `monax` can interact with services, please type:

```
monax services
```

## Chains

Chains are an opinionated toolchain around permissioned chains. They can be most easily thought of as your "develop" branch for chains. In other words, if you need to work **on** a permissioned chain, then it is best to use `monax chains`. Chains hardcode most of the service starting criteria, but still allow for some flexibility as to how chains are worked with.

To see the various ways in which `monax` can help you develop chains, please type:

```
monax chains
```

## Pkgs

Pkgs are an opinionated toolkit to help you deploy and test your smart contract packages on both permissioned and unpermissioned blockchain networks.

`monax pkgs` is a package manager to deal with contracts. The package manager is a yaml based automation framework which makes it trivial to deploy and test your smart contract systems. The specification for `monax:jobs` definition files is located [here](https://monax.io/docs/specs/jobs_specification).

Pkgs give you access to test your smart contracts both against "throwaway chains" which are one time use chains that are needed for the sole purpose of testing smart contract packages, as well as existing blockchain networks.

To see the various ways in which `monax` can help you develop smart contract applications, please type:

```
monax pkgs
```

## Keys

Keys is an opinionated toolchain around [monax-keys](https://github.com/monax/keys). Please note that this concept of the `monax` platform is **for development only** and should not be used in production because it has not been fully security audited **and we do not plan for it to be**. In production the keys service should be replaced with your audited security system of choice.

To see the various ways in which `monax` can help you manage your various key pairs, please type:

```
monax keys
```

## Files

Monax has a pretty handy wrapper around IPFS which is useful for quick file sharing from the host.

To see the various ways in which `monax` can help you with distributed file sharing, please type:

```
monax files
```

## Data

Monax can automagically utilize data containers for you.

If you turn the `data_container` variable to `true` in the service or chain definition file, then `monax` deposit the data utilized by that service or chain into a data container which can be managed separately from the "program" container. The advantage of working with data containers has been dealt with elsewhere (see, Google).

To see the various ways in which `monax` can help you manage your data containers, please type:

```
monax data
```

## Directory Structure

Created by `monax init` in $HOME directory:

```
├── .monax/
│   ├── monax.toml
│   ├── apps/
│   ├── bundles/
│   ├── chains/
│       ├── account-types/
│       ├── chain-types/
│   ├── keys/
│       ├── data/
│       ├── names/
│   ├── scratch/
│       ├── data/
│       ├── languages/
│       ├── lllc/
│       ├── ser/
│       ├── sol/
│   ├── services/
│       ├── ipfs.toml
│       ├── keys.toml
```

# Contributions

Are Welcome! Before submitting a pull request please:

* fork from `develop`
* go fmt your changes
* have tests
* pull request
* be awesome

A note about glide specifically as it regards to CLI:

To add a package as a dependency into CLI, make sure that you have [glide](http://glide.readthedocs.io/en/latest/#installing-glide) installed, and then follow these steps:

```
# add the package to the glide.yaml
glide get your/package/here
# make changes to the glide.yaml file for versioning purposes
# update the glide.lock file
glide up
# install the dependencies in vendor
glide install
# Use glide vendor cleaner to keep the vendor small
./cleanVendor.sh
# commit the vendor and push
git add vendor/*
git commit -sm "some helpful message here about the dependency added"
git push yourRepo yourBranch
```

See our [CONTRIBUTING.md](.github/CONTRIBUTING.md) and [PULL_REQUEST_TEMPLATE.md](.github/PULL_REQUEST_TEMPLATE.md) for more details.

Please note that this repository is GPLv3.0 per the LICENSE file. Any code which is contributed via pull request shall be deemed to have consented to GPLv3.0 via submission of the code (were such code accepted into the repository).



# Bug Reporting

Found a bug in our stack? Make an issue!

The [issue template](.github/ISSUE_TEMPLATE.md) specifies what needs to be included in your issue and will autopopulate the issue.

# License

[Proudly GPL-3](http://www.gnu.org/philosophy/enforcing-gpl.en.html). See [license file](LICENSE.md)
