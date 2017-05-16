---

layout: single
title: "Tutorials | Known Chain Making"
aliases:
  - /docs/known-chain-making
menu:
  tutorials:
    weight: 5

---

<div class="note">
{{% data_sites rename_docs %}}
</div>

There are three steps to making a permissioned chain with known keys:

1. Make for (or get from) the public keys for all parties
2. Make the `genesis.json` file and share it
3. Sort out a `config.toml` for each party
4. Instantiate the chain

We shall go through these in their logical order.

## Users Design

To do this we need to, first, consider, *who* will get *what* permissions and *why*. It is outside the scope of this tutorial to outline all of the considerations which would come into play when thinking about creating a permissioning system, but for the purposes of this tutorial, we will craft the genesis block to use the following paradigm:

* 1 Administrator (the developer who has **full** control over the chain) => this is a "Full Account" type
* 2 Validators (who participate in the consensus of the chain but do nothing else) => this is a "Validator Account" type

If you would like to understand all of the permissions which an monax chains smart contract network is capable of providing, [please see here for more information](/platform/db). 

We use an abstraction to simplify the chain making process called Account Types. This abstraction is just that, an abstraction to help users quickly get up to speed. In order to reduce the complexity of dealing with different types of accounts typically built on a chain, we use the idea of "account types". Account types are not restrictive in the sense that they are not the "only" types of accounts you can make with monax chains.

Account types are simply bundles of permissions no more no less. Using the monax tooling you can also create your own account types with your own bundles of permissions which will be helpful.

To learn about advanced chain making and account types, [see here](/docs/chain-making).

## Step 1: Make (or get) the public keys

Everyone who interacts with an monax chain will need to have a properly formated keypair. To make a keypair we will use `monax keys`.

`monax keys` usually operates as a signing daemon, but when we use monax keys to *create* key pairs what we are doing effectively is writing files. As is usual with the Monax tooling, `monax keys` is opinionated and will work by default against the following directory: `~/.monax/keys/data`. When a key pair is created, that key pair will get written into that directory.

These files will be written to a file system located inside the monax keys data container. As we go through this tutorial we will explain a bit about what that means. When we are using containers, these containers are not built to *hold* data, but rather are built to hold what is needed to run processes. But, if we're making keypairs, then we definitely want to *keep* these.

To accomplish this, we will use the `monax` tooling. First we need to start the `monax-keys` daemon:

```bash
monax services start keys
```

Check that is it indeed running with:

```bash
monax ls
```

You'll see something like:

```bash
SERVICE     ON     VERSION
keys        *      0.16.0 
```

which indicates (`*` rather than `-`) that the keys services is on (running). To see a more comprehensive output for your services, try `monax ls --all`

To see what we can do with monax keys we will run:

```bash
monax services exec keys "monax-keys -h"
```

This runs the `monax-keys -h` command "inside" the keys container.

Instead of dealing with the `monax-keys` service directly, however, we will use `monax keys` from the `monax` tool. The `monax keys` commands are basically wrappers around the `monax-keys` commands which are ran inside containers. To see the wrappers which the `monax` tooling provides around the `monax-keys` daemon, please type:

```bash
monax keys -h
```

Now it is time to generate some keys!

For the purposes of this tutorial **only** we will also create all of the necessary keys for all of the "users" of the chain and we will do so without passwords. Again, this is for demonstration purposes only, for a production system you will not do what we're about to do.

```bash
monax keys gen --save
```

This will create one key for you. The output here should look something like this:

```irc
Saving key to host      49CA2456F65B524BDEF50217AE539B8E10B37421
```

The `--save` flag exported your key, which will be in the directory at `~/.monax/keys/data/49CA2456F65B524BDEF50217AE539B8E10B37421/`. If omitted, you key will remain in the container and you can use `monax keys export 49CA2456F65B524BDEF50217AE539B8E10B37421` to save it to host afterwards.

To see the keys which monax-keys generated both *inside* the container type and available on your host machine type:

```bash
monax keys ls
```

Each of the three participants in this chain would run these series of commands independently, on their own trusted machine then submit their public key and address to whoever is making the `genesis.json`. For maximum trust in the chain, each party ought to generate their own `genesis.json` and ensure they match.

**Note** In version 0.16, we do not have a simple method of easily creating a `priv_validator.json` from an existing key. Thus, rather than creating keys via `monax keys gen --save`, we're going to take advantage of `monax chains make` as it creates keys and some other files we'll need. These three commands are meant to be run _seperately by each participant_ on their own machine:

```bash
monax chains make throwawayFull --account-type=Full:1
monax chains make throwawayVal --account-type=Validator:1
monax chains make throwawayVal --account-type=Validator:1
```

then getting the address and exporting the key from container to host (`monax keys export ADDR`). This will create a `priv_validator.json` in the chains directory on the host. It is needed for Step 4. Future versions will simplify this process and abstract/harden the way keys are handled.

**End Note**

Next, we'll make the all important `genesis.json` file.

## Step 2: Make the genesis.json

Before we begin, let's walk through the various files which are needed to run a chain:

1. the `genesis.json` which tells the chain how it should configure itself at the beginning (or, its genesis state).
2. the chain configuration file for Monax chains is called `config.toml`; see Step 3 for more information.
3. the keypair used by tendermint for signing blocks is the `priv_validator.json`.

All three files are usually located in the `~/.monax/chains/<your_chain>/<an_account>` directory after running `monax chains make`

The `genesis.json` is the primary file which tells monax chains how to instantiate a particular chain. It provides the "genesis" state of the chain including the accounts, permissions, and validators which will be used at the beginning of the chain. These can always be updated over the life of the chain of course, but the `genesis.json` provides the starting point.

With all that said, we're ready to make a known chain. Doing so requires preparing two additional files and using three additional flags to the `monax chains make` command.

First, you'll need to the public keys for each party. This can be found in the `priv_validator.json` as described in the Note in Step 1. Then, make a file named `accounts.csv` and replace the public keys seen below with the ones submitted by each participant:

```bash
0962E87A7A75B27174FB0F2C76FE9A54B78BBD5AD0E9605BF946E08F080BC657,99999999999999,admin_alice,16383,16383
C8E4C807152F70B5CE44E072D2CFB34F2521382CCA892AE90F1E058EE619E418,9999999999,validator_bob,32,16383
22774A27B9471BD7B0D9015B35067020FA99E7DD017721A577864127789DA0F2,9999999999,validator_charlie,32,16383
```

as well as a file named `validators.csv`, which is _nearly_ identical:

```bash
0962E87A7A75B27174FB0F2C76FE9A54B78BBD5AD0E9605BF946E08F080BC657,99999999999999,admin_alice,16383,16383
C8E4C807152F70B5CE44E072D2CFB34F2521382CCA892AE90F1E058EE619E418,9999999998,validator_bob,32,16383
22774A27B9471BD7B0D9015B35067020FA99E7DD017721A577864127789DA0F2,9999999998,validator_charlie,32,16383
```

Some notes on these two files: they are automatically generated when making unknown chains and are useful templates that can be used to regenerate a `genesis.json`. In this case, however, because all accounts happen to also be validators, the files are _nearly_ identical. With a chain where some accounts are not validators, the `accounts.csv` will have more accounts than the `validators.csv`. The latter two rows in the `validators.csv` have one less token allocated compared to the same rows in the `accounts.csv`. This is because the `99999...` column in the `accounts.csv` is the initial token allocation whereas in the `validators.csv` this number represent the number of _tokens to bond_. The third column in both the `.csv`'s is the account name and should, but need not, match the `moniker` field in the `config.toml`. The latter two columns handle permissions, a topic dealt with elsewhere.

We're ready to go!

```bash
monax chains make myCustomChain --known --accounts /path/to/accounts.csv --validators /path/to/validators.csv
```

This will output a `genesis.json` to stdout, so we recommend adding `>> genesis.json` to the end of the above command.

Finally, distribute this `genesis.json` to two validators who will be participating in the chain. They should also make one and confirm that it is identical.

## Step 3: Sort the config.toml

With the exception of known chains (this tutorial), the `monax chains make` command automatically makes a `config.toml` _for each account_. The moniker field for each account willtake on the name of the account. Optionally, the `--seeds-ip` flag will take a csv string of IP:PORT combinations and is used to point your consensus engine to the peers it should connect into. Review the [advanced chain deploying tutorial](/docs/chain-deploying) for more information.

Since we used `monax chains make` rather than (ideally, and in later versions) `monax keys gen` to make a key for each participant, they will each have an existing `config.toml`. At this point, `admin_alice` (the Full Account administrator) will provide his or her public IP address to the validators. Each user should now edit their `config.toml` to have the `moniker` field with their respective account names, while `validator_bob` and `validator_charlie` should edit `seeds = "IP-OF-ALICE:46656"`.

At this point, each user should have, on their own machine:

* keys service running, with one key in it
* the key saved (exported) to host
* a directory with each: `priv_validator.json`, `config.toml`, and `genesis.json`

For the next step, we are assuming that each user has the files in a directory like so:

```bash
~/.monax/chains/myCustomChain/admin_alice
~/.monax/chains/myCustomChain/validator_bob
~/.monax/chains/myCustomChain/validator_charlie
```

on their respective machines.

## Step 4: Instantiate the chain

Given the directory structure above and a chain name of myCustomChain, each user runs:

```bash
monax chains start myCustomChain --init-dir ~/.monax/chain/myCustomChain/admin_alice
monax chains start myCustomChain --init-dir ~/.monax/chain/myCustomChain/validator_bob
monax chains start myCustomChain --init-dir ~/.monax/chain/myCustomChain/validator_charlie
```

in that order (well, as long as `admin_alice` is run first). What will happen is that once `admin_alice` is up and running, each `validator_`, will, when started, "dial-in" to connect to `admin_alice`.

And there you have it; a custom chain built using pre-generated keys!


## [<i class="fa fa-chevron-circle-left" aria-hidden="true"></i> All Tutorials](/docs/)
