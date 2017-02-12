# Introduction

In general what is going to happen here is that we are going to establish what we at Eris call a "peer sergeant major" node who is responsible for being the easy connection point for any nodes which need to connect into the system.

In addition to the one "peer sergeant major" we will also deploy six "peer sergeants" who will be cloud based validator nodes.

## Overview of Tutorial

In general we are going to take two steps in order to get the chain setup:

1. Deploy the chain to each machine using `eris`
2. Connect into the chain locally

# Step 1. Deploy the Chain to Each Machine

Now that we have our machines created we're ready to deploy the chain. We need to do one thing before we deploy the chain: we're going to need to change the config.toml files.

```bash
cd ~/.eris/chains/advchain
cp ../default/config.toml .
```

Before we edit the file, let's get the IP address of our peer sergeant major node.

```bash
docker-machine ip my-advchain-val-000
```

You'll want to copy that IP address.

Now let's open the `~/.eris/chains/advchain/config.toml` file in our favorite text editor. Edit the config file so it looks like this:

```toml
moniker = "something_different"
seeds = "XX.XX.XX.XX:46656"
fast_sync = false
db_backend = "leveldb"
log_level = "debug"
node_laddr = "0.0.0.0:46656"
rpc_laddr = "0.0.0.0:46657"
vm_log = false
```

**N.B.**

For decentralized purists that may not like a single point of failure, a comma delimited list of peers may be entered for `seeds`. To find the IP addresses of all docker machines:

```bash
docker-machine ip $(docker-machine ls -q)
```

Note that in the `seeds` field you will use the IP address from docker-machine ip command rather than the `XX.XX.XX.XX` in the above.

Now we will copy the config.toml into all of our directories.

```bash
find . -mindepth 1 -maxdepth 1 -type d -exec cp config.toml {} \;
```

Now it's time to turn on the chain on our peer server / validator nodes.

But first, a sidebar about ports. Understanding the ports is important for distributed software. If the blockchains *think* they are running on port X, but that port is exposed to the internet as port Y when they are doing their advertising to their peers they will be saying, "Hey, I'm located on IP address Z on port X". But the problem is that from the internet's perspective they should really be saying "Hey, I'm located on IP address Z on port Y".

So at Eris we routinely recommend that you simply "flow through" the ports rather than trying to do anything funky here; this means that whatever port you select in the `laddr` fields and in the chain definition file, that you publish the same port on the host (meaning don't have something like `11111:46656` in your chain definition file). It can be made to work, but it requires some doing to do that right. But for now we will only be running one chain on each of our cloud validators so there will not be any port conflicts to worry about.

One thing to watch if you hard code the ports which the host machine will expose is that you will need to have these be unique for each chain so you will either only be able to run one chain per node or you'll need to use different ports for the other chain.

Let's do one more thing before we start the chain on a bunch of machines. If you've been using eris so far you may have had to say `yes` when the marmots asked you if you wanted to pull images. Let's tell the marmots to shut up.

```bash
export ERIS_PULL_APPROVE="true"
```

Many of us at Eris put that in our ~/.bashrc, ~/.zshrc or equivalent. Now, deploying your chain to a specific machine with eris is pretty simple.

```bash
for i in `seq 0 6`
do
  eris chains start advchain --init-dir advchain/"advchain_validator_00$i" --machine "my-advchain-val-00$i"
done
```

You're chain should now be running.

We're now going to start a few services which help us manage cloud instances.

1. We're going to start a `logrotate` service. This service is **absolutely essential** when working with cloud boxes. We have had **dozens** of cloud nodes overfill on us due to logs overloading the allocated storage space on the node. To overcome this, we use a [logs rotator service](https://github.com/tutumcloud/logrotate) which discards the old logs.
2. To couterbalance this we also will be starting a `logspout` service. This service will "spout" our logs to a logs "collector" service. To provide this service we use PapertrailApp, but [you could use others](https://github.com/gliderlabs/logspout).
3. We're going to start a `watchtower` service. This service will ping the docker hub for the images on the docker machine and if there are any updates to the image, they will automatically pull in the updates and gracefully restart all our containers for us. We can do this because of docker's fine grained tags system allows us to fine tune what docker image we are using. Users get the benefit when turning a `watchtower` service on that any tested pushes or security fixes which the service providers push to the docker hub will automatically be updated within about 5 minutes of pushing.

But first we need to make a simple change to one file. Let's edit the logspout service.

```bash
eris services edit logspout
```

In this file you'll edit the following line:

```toml
command = "syslog://logs2.papertrailapp.com:XXXX"
```

You can use any of the services logspout provides. Or if you use PaperTrail, then just update with your subdomain and/or port. You will need to edit the `logspout.toml` file in each node that you are running either via adding a `--machine` flag on the edit logspout service command above or by `scp`'ing a locally created file into each node's `~/.eris/services/` folder.

Now let's get those services turned on.

```bash
for i in `seq 0 6`
do
  eris services start watchtower logrotate logspout --machine "my-advchain-val-00$i"
done
```

That's it, we added all that functionality to our system with that little command!

# Step 3. Connect Into The Chain Locally

Now we need to connect into the chain from our local nodes now that the cloud based validator nodes are all set up. That will take some time.

```bash
eris chains start advchain --init-dir advchain/advchain_root_000
```

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
eris chains logs advchain -f --machine my-advchain-val-001
```

Change the machine name to cycle thru the logs and make sure blocks are coming in.

Oh wait. That didn't take long at all. Now you're all set up. Connected up to custom built, permissioned smart contract network with cloud based validators, given yourself admin permissions, and in what essentially has boiled down to move a few files around, edit a few lines in a few config files, and enter a few commands, we're ready to build out our applications.

# Clean Up

Let's remove those validator machines since we will not use them for the rest of these tutorials and we don't want to drive up our cloud hosting bills any more!

```bash
for i in `seq 0 6`
do
  docker-machine rm -y "my-advchain-val-00$i"
done
```
