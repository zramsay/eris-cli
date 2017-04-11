---

type:   docs
layout: single
title: "Deprecated | Chain Maintaining"

---

## Introduction

<div class="note">
{{% data_sites rename_docs %}}
</div>

In general what is going to happen here is that we are going to establish what we at Monax call a "peer sergeant major" node who is responsible for being the easy connection point for any nodes which need to connect into the system. While we understand that decentralized purists will not like the single point of failure, at this point it is the most viable way to orchestrate a blockchain network.

In addition to the one "peer sergeant major" we will also deploy six "peer sergeants" who will be cloud based validator nodes.

## Overview

In this tutorial we will cover the following maintenance tasks:

* [Adding a validator, user/participants, new developer to your team, etc.](#adding)
* [Removing a validator, etc.](#removing)
* [Understanding your chain's status](#status)
* [Chain Recovery Process](#recovery)
* [Chain Upgrade Process](#upgrade)

Over the course of working through these processes, we'll be introducing `eris actions` which has not, yet, been covered as there as not then a need for it. Now, we can leverage this aspect of `eris` to our benefit to greatly reduce the friction around standardized processes.

### Get Setup Part 1: The "Remote" Machines

For the purposes of this tutorial let's first make a chain with three validators (we're going to run them on virtualbox locally for this tutorial) and one root node (which we're going to run "locally"; or if you are on OSX or Windows in the main `eris` machine).

```bash
eris chains make maintainchain --account-types Validator:3,Root:1
```

Let's create some machines and start the chain on them. For this tutorial we will use the virtualbox driver for docker-machine but you can use any of the drivers which you prefer.

```bash
machine_base="my-valpool"
chain_name="maintainchain"
val_num=3
driver=virtualbox
# it can be annoying when working in bulk to manually approve each pull
export MONAX_PULL_APPROVE="true"
# we'll make enough validator machines to match our $val_num validators on the chain
for i in `seq 0 $(expr $val_num - 1)`
do
  # make the machine
  # eris chains stop -rxf --machine "$machine_base-$i" $chain_name
  docker-machine create "$machine_base-$i" --driver "$driver"
  # save the IP address of the previous machine
  if [ $i -eq 0 ]
  then
    peer_server_ip=$(docker-machine ip "$machine_base"-0)
  else
    peer_server_ip=$(docker-machine ip "$machine_base"-$(expr $i - 1))
  fi
  # perform an ugly text transform to get the configs sorted
  # note: seeds can be set in [chains make]
 cat ~/.eris/chains/$chain_name/"$chain_name"_validator_00"$i"/config.toml | \
    sed -e 's/seeds.*$/seeds = "'"$peer_server_ip"':46656"/g' | \
    > ~/.eris/chains/$chain_name/"$chain_name"_validator_00"$i"/config.toml
  # start the chain on this machine with a logs rotator on (a good practice for validator nodes)
  eris chains start $chain_name --init-dir $chain_name/"$chain_name"_validator_00"$i" --machine "$machine_base-$i" --logrotate
done
```

**N.B.** even though the Docker images *are* available either on your host (if you're on Linux) or within the `eris` machine (if you're on OSX or Windows), they will not be available on the newly created machines. The above sequence will kick off redundant downloading of the images. **Please do not run** this tutorial if you are on a slow connection as it will take **a very long time** just to pull all the appropriate images. If your connection is reasonably fast, it should not take too long.

**N.B. 2** `fast_sync` is a bool in the `config.toml` that should *always, by default* be set to `false`. The exception is when connecting a new peer/node to a long-running chain. Setting `fast_sync = true` will help the new peer sync faster than she otherwise would. *However, its behaviour is known to be unpredictable, especially with few validators and/or mulitple new peers trying to sync.* If you encounter problems with your chain, replace `true` with `false` and reset your chain.

**Temporary Hack**

That `cat | sed` sequence is ugly, we know. We'll be updating the tutorials to reflect the addition of the [--seeds-ip] flag for the [eris chains make] command.

**End Temporary Hack**

You can use that bash snippet for a wide variety of networks you need to establish. Just change the variables at the top of the snippet to suit your scenario.

### Get Setup Part 2: Turn on Your "Local" Node

The next step is to boot the chain locally and to make sure it is making blocks.

```bash
# copy over config and get our "local" node (which will utilize the root key) booted
cat ~/.eris/chains/$chain_name/"$chain_name"_validator_000/config.toml | \
  sed -e 's/moniker.*$/moniker = "imma_b_da_root"/g' \
  > ~/.eris/chains/$chain_name/"$chain_name"_root_000/config.toml
eris chains start $chain_name--init-dir $chain_name/"$chain_name"_root_000
sleep 10 # let it boot before we check the logs
eris chains logs "$chain_name"
```

**WAIT**. I can't read those logs you say. We know, it's not ideal. Moving away from having to read logs is on the workplan as well. To make sure your chain is connected and making blocks you'll scan the logs for output which looks like this:

```irc
INFO[03-20|14:03:12] Finalizing commit of block: Block{
  Header{
    ChainID:        maintainchain
    Height:         15
    Time:           2016-03-20 14:03:11.992 +0000 UTC
    Fees:           0
    NumTxs:         0
    LastBlockHash:  BA447E6C2DF7CA6FB905370A96D54CE24E4E30E6
    LastBlockParts: PartSet{T:1 D96802B334FD}
    StateHash:      A94DAE610BD590C4B7434C5307E0A0852496A649
  }#B53E39168416389BBE260534165BD1059EE3DE38
  Data{

  }#
  Validation{
    Precommits: Vote{14/00/2(Precommit) BA447E6C2DF7#PartSet{T:1 D96802B334FD} /DC9FF37F0BCE.../}
    Vote{14/00/2(Precommit) BA447E6C2DF7#PartSet{T:1 D96802B334FD} /61C71F8F1E06.../}
    Vote{14/00/2(Precommit) BA447E6C2DF7#PartSet{T:1 D96802B334FD} /CB24BEEC9406.../}
  }#24624E4310D98BACFB4F0CB7F6143BADBD00635C
}#B53E39168416389BBE260534165BD1059EE3DE38 module=consensus
```

That will let us know that our chain is all connected.

Now that we have a chain running let's do some fun tasks on it.

## Maintenance Task 1: Adding Actors

Let us start by making a key. Generally, best practice is for the actor you are trying to add to the permissioned chain network will generate the key on their machine and notice you (presumably one of the chain administrators) what their generated address is. But for the purposes of this tutorial we will simply do it all on our machine.

```bash
new_addr=$(eris keys gen)
echo $new_addr
```

OK, we have generated a key and saved it as a bash variable so we don't have to type it in the future. We need to get one more piece of information, namely the address of the root key.

```bash
root_addr=$(cat ~/.eris/chains/$chain_name/addresses.csv | grep "root_000" | cut -d ',' -f 1)
echo $root_addr
```

Now, we are about to add the $new_addr account on the chain. But what is our baseline? We need to have an easy way to get check if an address is in the accounts of our chain. There are various ways to get this baseline, but we'll cover the easiest way here.

```bash
chain_host=$(eris chains inspect $chain_name NetworkSettings.IPAddress)
if [ ! -z $(curl -s $chain_host:46657/list_accounts | jq '.result[1].accounts[].address' | grep $new_addr) ]
then
  echo "Account Present"
else
  echo "Account Not Present"
fi
```

Now this is a repeatable process which we are going to be using a lot. So instead of pasting a bunch of bash all the time. Let's abstract the above sequence into an `eris action`. But first, what are eris actions? Let's ask `eris`.

```bash
eris actions -h
```

Let's start by making a new action and we'll call it `account exists`.

```bash
eris actions new account exists
```

Next we will edit the action definition file.

```bash
eris actions edit account exists
```

Now we will paste in the following.

```toml
name = "account exists"
steps = [
  "if [ ! -z $(curl -s $(eris chains inspect $chain NetworkSettings.IPAddress):46657/list_accounts | jq '.result[1].accounts[].address' | grep $1) ] ; then echo \"Account Present\" ; else echo \"Account Not Present\" ; fi"
]
```

Now we have a handy shortcut for whenever we want to check an address. Let's test this:

```bash
eris actions do account exists $new_addr --chain $chain_name
eris actions do account exists $root_addr --chain $chain_name
```

OK. That's all well and good. But now we need to add an account to the chain. Let's say for the account we want it to have `call` and `send` permissions. Let's see how we can do that without `eris actions`. This time we'll use the mintx permission call.

```bash
eris chains exec $chain_name -- mintx permission set_base $new_addr call true --addr $root_addr --sign-addr=keys:4767 --node-addr=chain:46657 --chainID $chain_name --wait --broadcast --debug
```



#### Maintenance Task 2: Removing Actors



#### Maintenance Task 3: Understanding Chain Status



#### Maintenance Task 4: Chain Recovery



#### Maintenance Task 5: Chain Upgrade



## Cleanup

Let's remove those machines so they don't get in our way.

```bash
for i in `seq 0 2`
do
  docker-machine rm -y "$machine_base-$i"
done
```


## [<i class="fa fa-chevron-circle-left" aria-hidden="true"></i> All Deprecated](/docs/deprecated/)


