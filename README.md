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
go install
eris init
```

# Introduction

`eris` is a tool which makes it easy for developers to build, test, manage, and operate their distributed applications.

**No matter the blockchain**.

# Why?

Blockchain applications are a nightmare. They really are. As currently implemented blockchain-backed applications are almost always structured in one of two ways:

1. Treat the blockchain as a simple microservice and run it alongside a "normal" webApplication stack.
2. Completely buy into a single blockchain and its ecosystem which will wrap you in a warm hug of "we've got everything you need right here".

But we have always thought there was a better way.

# Eris: The Background

At Eris our approach to blockchain technologies is that blockchains, but really more interestingly and importantly, smart contract networks (which just so happen to currently mostly reside on blockchains), are an immensely helpful tool.

When used appropriately.

## Take 1: The Original Eris

In the summer of 2014 when we built reddit on a smart contract backbone we built a simple fork of eth's POC5 testnet, connected it to bittorrent (via trackerless torrentIDs put into contracts on our testNet using transmission), added in some pretty interesting smart contracts, and built a bunch of ruby glue to hold the thing together into an application built to provide users with a serverless application that had a view layer harmonized across nodes and where the data and processes were built across different and distinct distributed platforms for users. Folks wanted to tinker with this concept. Some from a technical perspective. Some from a social perspective. Whoever wanted to play with the damn thing, though, had to:

1. make sure ruby was installed locally
2. download poc5 c++ ethereum. get that installed for their platform.
3. (linux only) make sure that transmission was setup in the right way.
4. make sure you had git installed locally.
5. clone a repo.
6. turn on the client.
7. mine some fake ether.
8. go through a contract based registration process.
9. pray.

Needless to say, this was not a winning user experience.

Yet it was also a bit magical. We were building smart contracts which mediated content publishing. At the time we called it `contract controlled content dissemination` or (`c3d`). And all the content was served on a peer to peer basis. Exciting stuff! When the damn thing ran.

## Take 2: 2gather

In the winter of 2015 when we built youtube on a smart contract backbone we built a very complicated and divergent fork of eth's POC8 goclient with smart contract technology embedded in the genesis block providing a permissions layer for an ethereum style blockchain, connected it to ipfs and built a sophisticated go wrapper we called `decerver` which had scripting capabilities and provided a much more rich, while also much more robust middle layer than the Ruby glue we originally used to tie together early eth and transmission. Also ipfs, FTW! Again folks wanted to tinker with this idea. But whoever wanted to play with the damn thing, had to:

1. make sure go was installed locally
2. `go get github.com/eris-ltd/decerver/cmd/decerver`
3. go get ipfs
4. run a start up script
5. type their username into the screen (if they hadn't registered their key)
6. pray.

We were getting better. Still we had edge cases and some other challenges with getting everything just right for users.

Again, not so bad, but what if someone wanted to take the application and deploy it to ethereum and use bittorrent_sync rather than ipfs? It would have been doable with a very few lines of code actually. Yet, what kind of user experience would it be for the distributed application ecosystem if some users were on ethereum/bittorent sync application and others were on ethereum/swarm application and others were on decerver/ipfs application. And so help you god if you needed to switch from one to the other. How many binaries can even superusers with dev experience be expected to compile? Bloody nightmare.

# Eris: Philosophy

We have learned along this arc that blockchain-backed applications are hard. They're hard to develop against. They're hard to work *on*. They're hard to explain. And that's probably (a part of) why very few folks are using blockchains nearly as much as they *could*. (**N.B.**, This README does not presume to question your motives for being interested in a very interesting piece of technology; that nuanced dialog is better left for twitter.)

We have learned that its doable. It is doable to provide *something like* a web application experience with a completely distributed backend relying on smart contract and distributed data lake technologies.

We have learned that smart contracts are straight up legit. Verifiable, automate-able, distributed process. FTW!

We have learned that there are tons of great ideas out there. That tons of folks are working on incredibly interesting things. As such, we have learned the benefits (and reaped some of the costs) which come with a corporate philosophy that does not presume to establish which blockchain or which distributed file storage system or which peer to peer message system or any other system is right for **your** application.

We have learned that *application developers should have some choice* if the distributed application ecosystem is to blossom.

Choice in crafting the set of technologies which is right for their distributed application.

Choice in crafting which pieces of the application need to go in which data storage, data organization, and/or data dissemination facilities (which is what an application frontend -- no matter the backend -- needs).

Choice in where and how users are able to interact with their applications in a participatory manner which allows users (particularly superusers) to help application developers share the cost of scaling their application.

We have learned that *application users need a sane way to run distributed applications* and perhaps even more importantly than a sane way to run blockchain-backed applications but also a sane way to install and try out such applications.

**No matter the blockchain**.

These are the lessons which underpin our design of the `eris` tool.

# Eris: Today

`eris` is centered around a very few concepts:

* `services` -- things that you turn on or off
* `chains` -- develop permissioned chains
* `actions` -- step by step processes
* `contracts` (still a WIP) -- the newest iteration of our smart contract tool chain relying heavily on Andreas' solU and eris-db.js work.

We intend to add the following concepts over time:

* `projects` -- actions, workers, agents, contracts scoping feature
* `agents` -- local or remote agents which can adjust the settings of any node running `eris`
* `workers` -- scripted processes which need more logic than actions allow (node based, evented, middle layer)

These concepts (along with a few other goodies) provide the core functionality of what we think a true distributed application would look like.

# Installation

We haven't perfected the installation path yet. We will ensure that this path is a bit smoother prior to our 0.11 release.

## Dependencies

**N.B.** We will be distributing `eris` via binary builds by September or so. Until such time we do require it be built.

Installation requires that Docker be installed. Please see the [Docker](https://docs.docker.com/installation/) documentation for how to install.

At the current time, `eris` requires `docker` >= 1.6. You can check your docker version with `docker version`.

Installation requires that Go be installed. Please see the [Golang](https://golang.org/doc/install) documentation for how to install.

At the current time, `eris` requires `go` >= 1.4. You can check your go version with `go version`.

## Install Eris

```bash
go get github.com/eris-ltd/eris-cli
# there will be an error here about no buildable go files. ignore that.
cd $GOPATH/src/github.com/eris-ltd/eris-cli/cmd/eris && go install
eris init
```

That's it. You now can now operate the following services:

* btcd
* eth
* mint
* ipfs
* sandstorm
* keys (the eris-keys signing daemon)

among many others which are simply a single config file away.

Eris allows you to make and develop on chains. Import existing permissioned chains. Connect to public as well as permissioned blockchains. And perform many other functions which we have found useful in developing blockchain backed applications.

[For a further overview of this tool, please see here.](http://www.slideshare.net/CaseyKuhlman/erisplatform-introduction) or simply type:

```bash
eris
```

# Architecture of the Tool

From here on out, we're gonna go full nerd. Be forewarned.

[TODO]

# Contributions

Are Welcome! Before submitting a pull request please:

* go fmt your changes
* have tests
* be awesome

That's pretty much it (for now).

# License

GPL-3. See license file.