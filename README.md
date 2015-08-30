[![Circle CI](https://circleci.com/gh/eris-ltd/eris-cli/tree/master.svg?style=svg)](https://circleci.com/gh/eris-ltd/eris-cli/tree/master)

[![GoDoc](https://godoc.org/github.com/eris-ltd/eris-cli/cmd/eris?status.png)](https://godoc.org/github.com/eris-ltd/eris-cli/cmd/eris)

# Eris

The Distributed Application Platform.

# Go Vroom

Install Docker.

```
go get github.com/eris-ltd/eris-cli/cmd/eris
eris init
```

# Introduction

`eris` is a tool which makes it easy for developers to build, test, manage, and operate distributed applications.

**No matter the blockchain**.

[For the motivation behind this tool](docs/motivation.md).

# Eris: Today

```
Eris is a platform for building, testing, maintaining, and operating distributed
applications with a blockchain backend. Eris makes it easy and simple to wrangle
the dragons of smart contract blockchains.
```

The `eris` tool is centered around a very few concepts:

* `services` -- things that you turn on or off
* `chains` -- develop permissioned chains
* `actions` -- step by step processes
* `contracts` (still a WIP) -- the newest iteration of our smart contract tool chain

We intend to add the following concepts over time:

* `projects` -- actions, workers, agents, contracts scoping feature
* `agents` -- local or remote agents which can adjust the settings of any node running `eris`
* `workers` -- scripted processes which need more logic than actions allow

These concepts (along with a few other goodies) provide the core functionality of what we think a true distributed application would look like.

# Installation

We haven't perfected the installation path yet. We will ensure that this path is a bit smoother prior to our 0.11 release.

## Dependencies

**N.B.** We will be distributing `eris` via binary builds by September or so. Until such time we do require it be built.

We have prototyped an apt-get -able binary installation. **Warning, could break**. If you would like to protoype please run [this script](tests/hack/install_deb.sh) (as sudo).

We also have (experimental) downloadable builds from our [Github Releases Page](https://github.com/eris-ltd/eris-cli/releases).

### Docker

Installation requires that Docker be installed. Please see the [Docker](https://docs.docker.com/installation/) documentation for how to install.

At the current time, `eris` requires `docker` >= 1.7.1. You can check your docker version with `docker version`.

#### OSX

If you are on OSX, we **strongly recommend** that you install the [Docker Toolbox](https://www.docker.com/toolbox). The Toolbox will build Docker in a predictable way so that `eris` is able to connect into the Docker daemon.

If you do not install the Toolbox, then you will need to make sure that the `DOCKER_CERT_PATH` and the `DOCKER_HOST` environment variables have been set to wherever the certificates for connection to the Docker Daemon API's have been installed.

See [here](https://eng.erisindustries.com/tutorials/2015/08/05/ipfs-as-a-service/) for configuring a Toolbox-ed container with `eris files` IPFS functionality.

Note: The Toolbox obfuscates Kitematic in a convenient manner.

If you installed Docker via boot2docker, these *may* be set by running: `eval "$(boot2docker shellinit)"`.

#### Windows

If you are on Windows, we **strongly recommend** that you install the [Docker Toolbox](https://www.docker.com/toolbox). The Toolbox will build Docker in a predictable way so that `eris` is able to connect into the Docker daemon.

If you do not install the Toolbox, then you will need to make sure that the `DOCKER_CERT_PATH` and the `DOCKER_HOST` environment variables have been set to wherever the certificates for connection to the Docker Daemon API's have been installed.

See [here](https://eng.erisindustries.com/tutorials/2015/08/05/ipfs-as-a-service/) for configuring a Toolbox-ed container with `eris files` IPFS functionality.

Note: The Toolbox obfuscates Kitematic in a convenient manner.

If you installed Docker via boot2docker, these *may* be set by running: `eval "$(boot2docker shellinit)"`.

### Go

**Note** if you are one of the intrepid `apt-get` -ers then you do not need to have Go installed.

Installation requires that Go be installed. Please see the [Golang](https://golang.org/doc/install) documentation for how to install.

At the current time, `eris` requires `go` >= 1.4.2. You can check your go version with `go version`.

## Install Eris

```
go get github.com/eris-ltd/eris-cli/cmd/eris
eris init
```

That's it. You can now operate any of the services pre-built in the [eris:services](https://github.com/eris-ltd/eris-services) repository, among many others which are simply a single config file away.

Eris allows you to make and develop against numerous blockchains. You can

* import existing permissioned chains
* connect to public as well as permissioned blockchains
* perform many other functions which we have found useful in developing blockchain backed applications.

[For a further overview of this tool, please see here.](http://www.slideshare.net/CaseyKuhlman/erisplatform-introduction) or simply type:

```
eris
```

# Architecture of the Tool

From here on out, we're gonna go full nerd. Be forewarned.

The Eris CLI tool is mostly an opinionated wrapper around Docker's API.

We have found that running applications locally which require sophisticated installation paths and/or complex configuration work best when used from Docker's Container based system.

During our 0.9 series we have learned both the benefits of using Docker, as well as the challenges that using Docker present to distributed application developers and users. This learning, along with our user's feedback, have led directly the the current architecture. Each of the `concepts` listed above is described in a bit more detail below.

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
