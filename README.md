[![Circle CI](https://circleci.com/gh/eris-ltd/eris-cli/tree/master.svg?style=svg)](https://circleci.com/gh/eris-ltd/eris-cli/tree/master)

[![GoDoc](https://godoc.org/github.com/eris-ltd/eris-cli/cmd/eris?status.png)](https://godoc.org/github.com/eris-ltd/eris-cli/cmd/eris)

# Eris

The Distributed Application Platform.

# Go Vroom

Install Docker.

```bash
go get github.com/eris-ltd/eris-cli/cmd/eris
# TODO:::: remove next 2 lines post merge to master
cd $GOPATH/src/github.com/eris-ltd/eris-cli/cmd/eris
git checkout develop && go install
eris init
```

# Introduction

`eris` is a tool which makes it easy for developers to build, test, manage, and operate distributed applications.

**No matter the blockchain**.

# Why?

Blockchain applications are a nightmare. They really are. As currently implemented blockchain-backed applications are almost always structured in one of two ways:

1. They treat the blockchain as a simple microservice and run it alongside a "normal" webApplication stack.
2. They completely buy into a single blockchain and its ecosystem which will wrap you in a warm hug of "we've got everything you need right here".

But we have always thought there was a better way.

# Eris: The Background

At Eris our approach to blockchain technologies is that blockchains, but really more interestingly and importantly, smart contract networks (which just so happen to currently reside on blockchains), are an immensely helpful tool.

When used appropriately.

Here is a brief overview of our experiences designing, building, running, and distributing our own distributed applications with smart contract backbones.

## Take 1: The Original Eris

In the summer of 2014 when we built a reddit-like application on a smart contract backbone we built on a simple fork of eth's POC5 testnet, connected it to bittorrent, added in some pretty interesting smart contracts, and built a bunch of ruby glue to hold the thing together into an application built to provide users with a serverless application that had a view layer harmonized across nodes and where the data and processes were built across different and distinct distributed platform building blocks. Folks wanted to tinker with this concept. Some from a technical perspective. Some from a social perspective.

Whoever wanted to play with the damn thing, though, had to:

1. make sure ruby was installed locally
2. download poc5 c++ ethereum.
3. change some values in the c++ client so it would connect do a different peer server
4. get that installed for their platform.
5. (linux only) make sure that transmission was setup in the right way.
6. make sure git installed locally.
7. clone a repo.
8. turn on the client.
9. mine some fake ether.
10. go through a contract based registration process.
11. pray.

Needless to say, this was not a winning user experience.

Yet it was also a bit magical. We were building smart contracts which mediated content publishing. At the time we called it `contract controlled content dissemination` or (`c3d`). And all the content was served on a peer to peer basis.

Exciting stuff! When the damn thing ran.

## Take 2: The 0.9 Stack

In the winter of 2015 when we built a youtube-like application on a smart contract backbone we built on a very complicated and divergent fork of eth's POC8 goclient with smart contract technology embedded in the genesis block providing a permissions layer for an ethereum style POW blockchain, connected it to IPFS and built a sophisticated go wrapper we called `decerver` which had scripting capabilities and provided a much more rich, while also much more robust middle layer than the Ruby glue we originally used to tie together early eth and transmission. Also IPFS, FTW! Again folks wanted to tinker with this idea.

But whoever wanted to play with the damn thing, had to:

1. make sure go was installed locally
2. go get decerver
3. go get ipfs
4. run a start up script
5. type their username into the screen (if they hadn't registered their key)
6. pray.

We were getting better. Still we had edge cases and some other challenges with getting everything just right for users.

Again, not so bad, but what if someone wanted to take the application and deploy it to ethereum and use bittorrent_sync rather than IPFS? It would have been doable with a very few lines of code actually.

Yet, we could not help but wonder what kind of user experience would it be for the distributed application ecosystem if some users were on ethereum/bittorent sync application and others were on ethereum/eth-swarm application and others were on decerver/IPFS application.

How many binaries can even superusers with dev experience be expected to compile? Bloody nightmare.

It has been this journey which has led us directly to the many design and implementation details behind the first thing we've actually felt comfortable naming:

```bash
eris
```

# Eris: The Philosophy

We have learned along this arc that blockchain-backed applications are hard. They're hard to develop against. They're hard to work *on*. They're hard to explain. And that's probably (a part of) why very few folks are using blockchains nearly as much as they *could*.

(**N.B.**, This README does not presume to question your motives for being interested in a very interesting piece of technology; that nuanced dialog is better left for twitter.)

We have learned that its doable. It is doable to provide *something like* a web application experience with a completely distributed backend relying on smart contract and distributed data lake technologies.

We have learned that smart contracts are straight up legit. Verifiable, automate-able, distributed process.

We have learned that there are tons of great ideas out there. That tons of folks are working on incredibly interesting things. As such, we have learned the benefits (and reaped some of the costs) which come with a corporate philosophy that does not presume to establish which blockchain or which distributed file storage system or which peer to peer message system or any other system is right for **your** application.

We have learned that *application developers should have some choice* if the distributed application ecosystem is to blossom.

* Choice in crafting the set of technologies which is right for their distributed application.
* Choice in crafting which pieces of the application need to go in which data storage, data organization, and/or data dissemination facilities (which is what an application frontend -- no matter the backend -- needs).
* Choice in where and how users are able to interact with their applications in a participatory manner which allows users (particularly superusers) to help application developers share the cost of scaling their application.

We have learned that application superusers wanting to participate in the data storage, organization and/or dissemination of the application *need a sane way to run distributed applications* and perhaps even more importantly than a sane way to run blockchain-backed applications a sane way to install and try out such applications.

**No matter the blockchain**.

These are the lessons which underpin the design of the `eris` tool.

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

### Docker

Installation requires that Docker be installed. Please see the [Docker](https://docs.docker.com/installation/) documentation for how to install.

At the current time, `eris` requires `docker` >= 1.6. You can check your docker version with `docker version`.

#### OSX

If you are on OSX, we **strongly recommend** that you install Docker via [Kitematic](https://kitematic.com/). Kitematic will build Docker in a predictable way so that `eris` is able to connect into the Docker daemon.

If you do not use Kitematic, then you will need to make sure that the `DOCKER_CERT_PATH` has been set to wherever the certificates for connection to the Docker Daemon API's have been installed.

#### Windows

If you are on Windows, we **strongly recommend** that you install Docker via [Kitematic](https://kitematic.com/). Kitematic will build Docker in a predictable way so that `eris` is able to connect into the Docker daemon.

If you do not use Kitematic, then you will need to make sure that the `DOCKER_CERT_PATH` has been set to wherever the certificates for connection to the Docker Daemon API's have been installed.

### Go

Installation requires that Go be installed. Please see the [Golang](https://golang.org/doc/install) documentation for how to install.

At the current time, `eris` requires `go` >= 1.4. You can check your go version with `go version`.

## Install Eris

```bash
go get github.com/eris-ltd/eris-cli
# there will be an error here about no buildable go files. ignore that.
# TODO: remove next two lines post merge to master
cd $GOPATH/src/github.com/eris-ltd/eris-cli/cmd/eris
git checkout develop && go install
eris init
```

That's it. You can now operate any of the services pre-built in the [eris:services](https://github.com/eris-ltd/eris-services) repository, among many others which are simply a single config file away.

Eris allows you to make and develop against numerous blockchains. You can

* import existing permissioned chains
* connect to public as well as permissioned blockchains
* perform many other functions which we have found useful in developing blockchain backed applications.

[For a further overview of this tool, please see here.](http://www.slideshare.net/CaseyKuhlman/erisplatform-introduction) or simply type:

```bash
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

```bash
eris services
```

## Chains

Chains are an opinionated toolchain around permissioned tendermint blockchains. They can be most easily thought of as your "develop" branch for blockchains. In other words, if you need to work **on** a permissioned tendermint blockchain, then it is best to use `eris chains`.

Chains hardcode most of the service starting criteria, but still allow for some flexibility as to how chains are worked with. Chains are operated from a base of **chain definition files**. These files are held on the host in the following location: `~/.eris/chains`. The specification for chain definition files is located [here](docs/chains_specification.md).

To see the various ways in which `eris` can help you develop blockchains, please type:

```bash
eris chains
```

## Actions

Actions are step by step processes which need to take a few variables but otherwise should be scriptable. Actions are used for repetitive, or repeatable, tasks. The environment in which actions run is similar in nature to a modern continuous development environment. Actions are run on the **host** and from within a container. They have full access to containers either via the `eris` cli or via docker's cli.

Examples of things actions are made to support:

* register a domain entry via mindy
* drop a preformulated transaction into the btc network using a specific key

Actions work from a base of **action definition files**. These files are held on the host in the following location: `~/.eris/actions`. Action definition files tell `eris` what steps to take, what services to make available, what chain to run, and what steps to take. The specification for action definition files is located [here](docs/actions_specification.md).

To see the various ways in which `eris` can interact with actions, please type:

```bash
eris actions
```

## Data

Eris can automagically utilize [data containers](http://container42.com/2014/11/18/data-only-container-madness/) for you. If you turn the `data_container` variable to `true` in the service or chain definition file, then `eris` deposit "most" of the data utilized by that service or chain into a data container which can be managed separately from the "program" container. The advantage of working with data containers has been dealt with elsewhere.

To see the various ways in which `eris` can help you manage your data containers, please type:

```bash
eris data
```

## Files

Eris has a pretty handy wrapper around IPFS which is useful for quick file sharing from the host or for use by actions.

To see the various ways in which `eris` can help you with distributed file sharing, please type:

```bash
eris files
```

# Contributions

Are Welcome! Before submitting a pull request please:

* go fmt your changes
* have tests
* be awesome

That's pretty much it (for now).

# License

GPL-3. See [license file](https://github.com/eris-ltd/eris-cli/blob/master/LICENSE.md).