---

layout: single
title: "Tutorials | Known Chain Making"
aliases:
  - /docs/known-chain-making
menu:
  tutorials:
    weight: 5

---

// todo better intro

There are three steps to making a permissioned chain:

1. Make for (or get from) the public keys for all parties
2. Make the `genesis.json` file
3. Instantiate the chain

We shall go through these in their logical order.

## Users Design

To do this we need to, first, consider, *who* will get *what* permissions and *why*. It is outside the scope of this tutorial to outline all of the considerations which would come into play when thinking about creating a permissioning system, but for the purposes of this tutorial, we will craft the genesis block to use the following paradigm:

* 3 Administrators (these would be developers who have **full** control over the chain) (one of which will be "running" the chain performing validation)

If you would like to understand all of the permissions which an eris chains smart contract network is capable of providing, [please see the eris-db repository for more information](https://github.com/eris-ltd/eris-db/blob/master/README.md).

We use an abstraction to simplify the chain making process called Account Types. This abstraction is just that, an abstraction to help users quickly get up to speed. In order to reduce the complexity of dealing with different types of accounts typically built on a chain, we use the idea of "account types". Account types are not restrictive in the sense that they are not the "only" types of accounts you can make with eris chains.

Account types are simply bundles of permissions no more no less. Using the eris tooling you can also create your own account types with your own bundles of permissions which will be helpful.

## Step 2.a.1. Make (or Get) the Public Keys

Everyone who interacts with an eris chain will need to have a properly formated keypair. To make a keypair we will use `eris keys`.

`eris keys` usually operates as a signing daemon, but when we use eris keys to *create* key pairs what we are doing effectively is writing files. As is usual with the eris tooling, `eris keys` is opinionated and will work by default against the following directory: `~/.eris/keys/data`. When a key pair is created, that key pair will get written into that directory.

These files will be written to a file system located inside the eris keys data container. As we go through this tutorial we will explain a bit about what that means. When we are using containers, these containers are not built to *hold* data, but rather are built to hold what is needed to run processes. But, if we're making keypairs, then we definitely want to *keep* these.

To accomplish this, we will use the `eris` tooling. First we need to start the `eris-keys` daemon:

```bash
eris services start keys
```

By default, `eris` is a very "quiet" tool. To check that the keys service started correctly type:

```bash
eris services ls
```

You'll see something like:

```bash
SERVICE     ON     VERSION
keys        *      0.16.0 
```

which indicates that the keys services is on (running). To see a more comprehensive output for your services, try `eris services ls -a`.

To see what we can do with eris keys we will run:

```bash
eris services exec keys "eris-keys -h"
```

What this is doing is running the `eris-keys -h` "inside" the keys containers. 

Instead of dealing with the `eris-keys` service directly, however, we will use `eris keys` from the eris cli tool. The `eris keys` commands are basically wrappers around the `eris-keys` commands which are ran inside containers. To see the wrappers which the eris cli tooling provides around the `eris-keys` daemon, please type:

```bash
eris keys -h
```

Now it is time to generate some keys!

For the purposes of this tutorial **only** we will also create all of the necessary keys for all of the "users" of the chain and we will do so without passwords. Again, this is for demonstration purposes only, for a production system you will not do what we're about to do.

```bash
eris keys gen --save
```

This will create one key for you. The output here should look something like this:

```irc
Saving key to host      49CA2456F65B524BDEF50217AE539B8E10B37421
```

The `--save` flag exported your key, which will be in the directory at `~/.eris/keys/data/49CA2456F65B524BDEF50217AE539B8E10B37421/`. If omitted, you key will remain in the container and you can use `eris keys export 49CA2456F65B524BDEF50217AE539B8E10B37421` to save it to host afterwards.

To see the keys which eris-keys generated both *inside* the container type and available on your host machine type:

```bash
eris keys ls
```

Now, we're all ready to make a chain.

## Step 2.a.2. Make the genesis.json

Before we begin, we should quickly talk through the various files which are needed to run an eris chain.  This is to hold the default files for using eris chains. There are a few primary files used by eris chains:

1. the chain definition file for Eris chains is called `config.toml` and is located in your `~/.eris/chains/<your_chain>` directory.
2. the `genesis.json` which tells Eris chains how it should configure itself at the beginning of the chain (or, its genesis state)
3. the keypair which the tendermit consensus engine will use to sign blocks, etc. called the `priv_validator.json`

The three files you *may* need to edit are the `genesis.json` and `priv_validator.json` (both of which we're about to get "made" for us) and the `config.toml`.

In any chain with more than one validator the `config.toml` file will be edited to fill in the `seeds` and `moniker` fields. The `seeds` field is used to point your consensus engine to the peers it should connect into. For more information on how to deal with this please see our [advanced chain deploying tutorial](/docs/chain-deploying/). The `moniker` field is "your node's name on the network". It should be unique on the given network.

The `genesis.json` is the primary file which tells eris chains how to instantiate a particular chain. It provides the "genesis" state of the chain including the accounts, permissions, and validators which will be used at the beginning of the chain. These can always be updated over the life of the chain of course, but the genesis.json provides the starting point. Luckily `eris` takes care of making this for you and there is very little which should be required for you in way of editing.

With all that said, we're ready to make a chain. First let us make a "fake" chain just to get a tour of the chain maker tool. Once we go through that process then we will make our "real" chain which we will use for the rest of this tutorial series. Let's see what eris chains make can do for us.


## [<i class="fa fa-chevron-circle-left" aria-hidden="true"></i> All Tutorials](/docs/)