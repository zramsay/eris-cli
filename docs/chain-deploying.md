---

layout: single
title: "Tutorials | Multi Node Chains"
aliases:
  - /docs/chain-deploying
menu:
  tutorials:
    weight: 5

---

# Introduction

In general what is going to happen here is that we are going to establish what we call a "peer sergeant major" cloud node who is responsible for being the easy connection point for any nodes which need to connect into the system.

In addition to the one "peer sergeant major" we will also deploy 3 "peer sergeants", as cloud based validator nodes.

Ideally, and as pre-requisite for a trustless consortium chain, we'd be using a [known chain](/known-chain-making), however, for the sake of simplicity, we'll be making our own keys in this tutorial.

Note also that we are using four validators for this chain. This means it will tolerate one validator node being "down" because tendermint consensus requires >2/3 validators online to "move forward".

Previously, the recommended way of creating a multi-node chain was with docker-machine, but we have [deprecated this tutorial](../deprecated/using-docker-machine-with-eris) and as of the 0.17.0 release, will be eliminating the global `--machine` flag.

# Overview of Tutorial

We are going to take these steps in order to get the chain setup:

1. Create four cloud machines and get their public IP addresses
2. Make the required files for the chain using `eris`
3. Copy the appropriate files to each cloud machine
4. Start the node on each machine using `eris`
5. Start additional services to ensure chain longevity
6. Inspect the health of our chain

# Step 1: Create cloud machine

Using any cloud provider of your choice, create four seperate instances and note their IP addresses. For the sake of this tutorial, we'll refer to these instances as CL0, CL1, CL2, and CL3 respectively. For example:

```
CL0: 159.134.23.97
CL1: 55.276.44.31
CL2: 276.37.22.79
CL3: 48.413.82.16
```

In this case, the IP addresses are fake so take note your own. You'll need to [install eris](/getting-started) and run `eris init` on each machine. Ensure `ssh` is enable on all machines.

# Step 2: Make the chain

We'll use `CL0` as our "peer seargent major", so ssh yourself in. For simplicty, we'll use one Full account and three Validator accounts:

```bash
eris chains make multichain --account-types=Full:1,Validator:3 --seeds-ip=159.134.23.97:46656,55.276.44.31:46656,276.37.22.79:46656,48.413.82.16:46656
```

What's this 46656 port? That's tendermint's p2p port that allows the nodes to find each other. For more information on ports, see below -> [^1]

The `--seeds-ip` flag was introduced in version 0.16.0 and will fill the `seeds` field in **each** `config.toml`, rather than the previously required manual entry method. Another new feature is that the `moniker` field will now take on the account name such that each `config.toml` now has a unique moniker.

We created a handful of directories within `~/.eris/chains/multichain`. Feel free to take a peek or head over to the [chain making tutorial](/chain-making) for a comprehensive explanation of these files.

For this tutorial, we'll be copying the raw directories as-is, however, note that the `eris chains make` command can be run with either the `--tar` or `--zip` flag as required for your scripting purposes.

# Step 3: Copy the files around

The following describes which directories are required for each cloud machine:

```
CL0: ~/.eris/chain/multichain/multichain_full_000
CL1: ~/.eris/chain/multichain/multichain_validator_000
CL2: ~/.eris/chain/multichain/multichain_validator_001
CL3: ~/.eris/chain/multichain/multichain_validator_002
```

Using `scp` or your preferred method, ensure each directory is on each machine.

## Step 4: Start the node on each cloud machine

You'll have to `ssh` into each machine:

On `CL0`, run:

```bash
eris chains start multichain --init-dir ~/.eris/chains/multichain_full_000 --logrotate
```

On CL1, run:

```bash
eris chains start multichain --init-dir ~/.eris/chains/multichain_validator_000 --logrotate
```

On CL2, run:

```bash
eris chains start multichain --init-dir ~/.eris/chains/multichain_validator_001 --logrotate
```

On CL3, run:

```bash
eris chains start multichain --init-dir ~/.eris/chains/multichain_validator_002 --logrotate
```

And voila! You multi-node, permissioned chain is started!

## Step 5: Start some services

We're now going to start a few services which help us manage cloud instances. You'll notice we used the `--logrotate` flag when starting the chains. This service is **absolutely essential** when working with cloud boxes. We have had **dozens** of cloud nodes overfill on us due to logs overloading the allocated storage space on the node. To overcome this, we use a [logs rotator service](https://github.com/tutumcloud/logrotate) which discards the old logs. If you forgot to use the flag, don't fret! `eris services start logrotate` will get you squared away.

To couterbalance the discarded logs we also will be starting a `logspout` service. This service will "spout" our logs to a logs "collector" service. To provide this service we use PapertrailApp, but [you could use others](https://github.com/gliderlabs/logspout).


But first we need to make a simple change to one file. Let's edit the logspout service.

```bash
eris services edit logspout
```

In this file you'll edit the following line:

```toml
command = "syslog://logs2.papertrailapp.com:XXXX"
```

You can use any of the services logspout provides. Or if you use PaperTrail, then just update with your subdomain and/or port. You will need to edit the `logspout.toml` file in each node that you are running either via adding a `--machine` flag on the edit logspout service command above or by `scp`'ing a locally created file into each node's `~/.eris/services/` folder.

That's it, we added all that functionality to our system with that little command! Optionally, you can use watchtower to automatically pull in the latest updates for your chain -> [^2]

## Step 6: Inspect health of chain

... better health inspection

Check that it is running:

```bash
eris chains ls
```

And see what its doing:

```bash
eris chains logs advchain -f
```

(`ctrl+c` to exit the logs following.) You can also pull the logs for one of the validators

```bash
eris chains logs advchain -f
```


Now you're all set up. Connected up to custom built, permissioned smart contract network with cloud based validators, given yourself admin permissions, and in what essentially has boiled down to move a few files around, edit a few lines in a few config files, and enter a few commands, we're ready to build out our applications.


[^1]

Understanding the ports is important for distributed software. If the blockchains *think* they are running on port X, but that port is exposed to the internet as port Y when they are doing their advertising to their peers they will be saying, "Hey, I'm located on IP address Z on port X". But the problem is that from the internet's perspective they should really be saying "Hey, I'm located on IP address Z on port Y".

So at Monax we routinely recommend that you simply "flow through" the ports rather than trying to do anything funky here; this means that whatever port you select in the `laddr` fields and in the chain definition file, that you publish the same port on the host (meaning don't have something like `11111:46656` in your chain definition file). It can be made to work, but it requires some doing to do that right. But for now we will only be running one chain on each of our cloud validators so there will not be any port conflicts to worry about.

One thing to watch if you hard code the ports which the host machine will expose is that you will need to have these be unique for each chain so you will either only be able to run one chain per node or you'll need to use different ports for the other chain.

[^2]

The watchtower service will ping the docker hub for the images on the docker machine and if there are any updates to the image, they will automatically pull in the updates and gracefully restart all our containers for us. We can do this because of docker's fine grained tags system allows us to fine tune what docker image we are using. Users get the benefit when turning a `watchtower` service on that any tested pushes or security fixes which the service providers push to the docker hub will automatically be updated within about 5 minutes of pushing.
