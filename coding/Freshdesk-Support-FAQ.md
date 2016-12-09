## 1. Transaction timeout or app.js errors

**QUESTION:** After upgrading the Eris version I see the error: 
![](https://s3.amazonaws.com/cdn.freshdesk.com/data/helpdesk/attachments/production/6027802067/original/blob1468894112412.png?1468894095)

**ANSWER 1:**
The debug checklist for that error is here: https://github.com/eris-ltd/eris-pm/blob/master/util/errors.go#L25-L30
* is the ADDRESS account right?
* is the account you want to use in your keys service: `eris keys ls` ?
* is the account you want to use in your genesis.json: `eris chains cat CHAIN_NAME genesis` ?
* is your chain making blocks: `eris chains logs -f CHAIN_NAME` ?
* do you have permissions to do what you're trying to do on the chain?
* do you have **typos** in your `app.js` file?

**ANSWER 2:**
If you are getting a **timeout** that usually means that you do not have enough 
validators present to formulate a quorum. On proof of stake consensus engines 
like Tendermint, which eris-chains currently uses, you need > 2/3rds of 
the bonded stake online in order for the chain to make new blocks. 
The reason that eris-pm is timing out is because it is waiting for a new 
block that confirms the entry of the transaction in it. If no blocks are 
getting created then the eris pkgs do command will time out. 

If you have multiple validators all with the same amount of bonded stake 
then start > 2/3rds of them, make sure they are connected via the same 
tutorial you reference. After that ping the /status end point for 
the IP:RPC_PORT for a node to make sure that the blockheight 
variable is increasing. This is how you'll know that the chain 
is making blocks. Once you are sure that the chain is making blocks 
then you should not have any trouble with the timeout.

> https://support.erisindustries.com/helpdesk/tickets/341

> https://support.erisindustries.com/helpdesk/tickets/332

> https://support.erisindustries.com/helpdesk/tickets/305

> https://support.erisindustries.com/helpdesk/tickets/330

> https://support.erisindustries.com/helpdesk/tickets/290

> https://support.erisindustries.com/helpdesk/tickets/276

## 2. Adding blocks if there are no transactions

**QUESTION:** I have a question concerning the chain height. It seems that the current 
implementation  keeps adding block to the chain even if there is no transaction 
(almost every second). Is there a way to change it?

**ANSWER:** This is to do with the consensus algorithm, Tendermint, that eris:db uses. 
Transactions are arranged into blocks, which are proposed and voted on 
regardless of the number of transactions available to a validating node 
at the time. A validator has a certain amount of time in which to propose 
a block when it is its turn to do so. If it is too slow then other 
validators may start to suspect it of foul play. So it does not wait 
indefinitely, we also do not want to wait for a **full load** of transactions 
as this could make the latency of a transaction being finalised too high 
in the case low transaction throughput. Still, in the 'degenerate' 
case where we have no transactions we could behave differently 
and not produce a block, but in that case we would still need some message 
from the validator whose turn it is to propose to say "nothing to see here, 
but I'm still alive and doing my job". In this case I think it is simply 
**more elegant** to use an **empty block** for this sign-of-life, since the logic 
is then just: look at your mempool, get transactions (possibly none), form a block.

You can find out more about Tendermint here: http://tendermint.com/, 
or in more detail (work in progress) in Ethan's 
thesis here: https://github.com/ebuchman/thesis/raw/master/Buchman_Ethan_201606_MAsc.pdf. 
I've also asked the Tendermint people whether there are different or better reasons than 
the ones I've given, so can update here if there is.

> https://support.erisindustries.com/helpdesk/tickets/325

## 3. Duplicate peer error

**QUESTION:** I am trying to deploy a service to three nodes.  I deployed a contract 
at one of the nodes.  I then pulled the service to the other two nodes.  
I am able to see that all three nodes are connected via the information 
at `$host:46657/net_info`. However, the second and third node that 
I start do not get the new contract on their chains.  
When looking at the chain logs, I see the following

```
Ignoring peer module=p2p error="Duplicate peer"
```

**ANSWER:** That error can safely be ignored. The first thing to check is that all 
three nodes have the same genesis file. You can do with with 
`eris chains cat chainName genesis`. To be even easier you can pipe 
that into a hashing function just to make sure that all three 
have the same hash. The configuration you're using would likely 
be OK as you'd have one validator and the other two nodes could come and go.

> https://support.erisindustries.com/helpdesk/tickets/312

## 4. Pulling from Github error because of a firewall

**QUESTION:** I'm trying to install Eris at Ubuntu Wily and after everything, I try to do `eris init` 
and after all the images have been downloaded from `quay.io`' I'm getting an error
```
Get https://raw.githubusercontent.com/eris-ltd/eris-services/master/bigchaindb.toml: 
dial tcp: lookup raw.githubusercontent.com on 127.0.1.1:53: no such host
```

**ANSWER:** As you're behind proxy, you can pull the files (`git clone`) manually. 

* https://github.com/eris-ltd/eris-services → `~/.eris/services`
* https://github.com/eris-ltd/eris-actions → `~/.eris/actions`
* https://github.com/eris-ltd/eris-chains → `~/.eris/chains`
* https://github.com/eris-ltd/eris-chains/blob/master/default.toml → `~/.eris/chains`

And the `~/.eris/eris.toml` file (its contents):
```
IpfsHost = "http://0.0.0.0"
CompilersHost = "https://compilers.eris.industries"
CrashReport = "bugsnag"
Verbose = false     
```

> https://support.erisindustries.com/helpdesk/tickets/311

> https://support.erisindustries.com/helpdesk/tickets/300

## 5. Eris consensus algorithms

**QUESTION:** On your Wiki you are mentioned about using different Consensus Algorithms. 
There are "Tendermint" (as primary), "PBFT" and "Deposit based PoS". 

Could you, please, clarify me how it works together? Can I select one which more convenient for me? Or 
they work together at the same time?

**ANSWER:** We are not yet to the point of having fully swappable consensus 
engines although that is on our roadmap. Tendermint is a PBFT, 
Deposit-Based Proof of Stake system, so the second items 
were descriptive of Tendermint rather than being additional options.

if you want to get into the details of Tendermint v. PBFT 
there are subtle differences of course. But they are close 
enough algorithms that come from the same lineage that we 
are OK with the verbiage as it stands for most purposes.

The Eris framework utilizes a range of chain designs, 
although the Tendermint consensus engine is the 
one most of our users utilize.

> https://support.erisindustries.com/helpdesk/tickets/296

## 6. Hosted compiler server issues

**QUESTIONS:** I am following the tutorial to deploy `idi.sol` to `simplechain`, 
but when I run this command `eris pkgs do --chain simplechain --address $addr`,
I get this error:
```
Performing action. This can sometimes take a wee while
Executing Job                                 defaultAddr
Executing Job                                 setStorageBase
Executing Job                                 deployStorageK
failed to send HTTP request Post https://compilers.eris.industries:10114/compile: dial tcp 188.166.90.67:10114: i/o timeout
Error compiling contracts
Post https://compilers.eris.industries:10114/compile: dial tcp 188.166.90.67:10114: i/o timeout
```

**ANSWER 1:** Our compilers server is live. Make sure that you are not behind a firewall and do not have any issues connecting to that URL and after rerunning everything should work properly.

**ANSWER 2:** Our compliler server is not live. Run the command `eris pkgs do --chain simplechain --address $addr --compiler https://compilers.eris.industries:10114`

**ANSWER 3:** You're behind a proxy / firewall:

Edit the file `~/.eris/services/compilers.toml` to use the `quay.io/eris/compilers:0.11.4` image
instead of the `quay.io/eris/compilers` image then:
```
docker pull quay.io/eris/compilers:0.11.4
eris services start compilers
comp_ip=$(docker-machine ip eris)
eris pkgs do --chain simplechain --address $addr --compiler http://$comp_ip:9091
```

> https://support.erisindustries.com/helpdesk/tickets/286

> https://github.com/eris-ltd/eris-compilers/issues/70#issuecomment-233919165

## 7. Buying Eris

**QUESTION:** I would like to discuss the **pricing schedule** for the Eris Platform.  
Is there a customer service phone number that I can call?

**ANSWER:** We don't have a pricing structure as our platform is free and open source. 
Just visit https://docs.erisindustries.com/tutorials/getting-started/ to get started.

> https://support.erisindustries.com/helpdesk/tickets/277

## 8. Mining Eris

**QUESTION:** After reading your documentation, I created my chain 
and tried to deploy my contract. How do I mine?

**ANSWER:** Proof of stake chains, such as Eris chains, **do not mine**. Rather they 
**validate**. In order to conduct their validation function they need to have >66% 
of the stake online and connected. 

> https://support.erisindustries.com/helpdesk/tickets/276

## 9. Port conflicts

**QUESTION:** I am trying to use  eris smart contract for test and learn. 
I came across the following error: 

```
Bind for 0.0.0.0:46657 failed: port is already allocated
```

**ANSWER:** When you start the second chain give it the `--publish` or `-p` flag 
which will select random ports to use or specify them explicitly with the
`--ports` flag.

> https://support.erisindustries.com/helpdesk/tickets/255

## 10. Eris transaction throughput

**QUESTION:** Could you tell me the maximum theoretical transaction processing capacity of Eris?

The result we got was 13 tx/sec and if transactions exceeded that number, docker process stopped.

**ANSWER:** That sounds like an issue with how you are testing. We routinely get much higher throughput than that. 

> https://support.erisindustries.com/helpdesk/tickets/337

> https://support.erisindustries.com/helpdesk/tickets/321

## 11. Accounts

We have a few questions that we could not answer yet from the documentation:

1. How to create accounts out of the genesis block? This is the case when 
a new bank joins the OTC Market consortium. We found the "create_account" 
permission documented here https://docs.erisindustries.com/documentation/eris-db-permissions/, 
but still have no hint on how to create accounts after the blockchain started.

2. How do we send bond transactions, in order to change 
the original stake assets? Again, this is the case when a 
new bank joins, we found the "bond" permission documented here https://docs.erisindustries.com/documentation/eris-db-permissions/, 
but we couldn't find how to use this feature.

3. We deployed a blockchain as described here https://docs.erisindustries.com/tutorials/advanced/chaindeploying/, 
therefore with a single point of failure, which is easy and good for testing. Is it possible to 
deploy a blockchain without a "peer sergent major"?

**ANSWER:**

1. There are two ways to create_accounts.

  a. The easy way to create a new account is to have an account which is present on the current chain send the account to be created a single token. 

  b. If tokens are meaningful within the system you are doing, you can also 
have an account which is present on the chain and possesses either 
create_account permissions (or root permissions) can formulate a 
create_account transaction to send to the chain. 

  At this time we have go tooling for formulating this transaction 
available via the [eris-pm tooling](https://docs.erisindustries.com/documentation/eris-pm/latest/jobs_specification/). 
The [javascript tooling](https://github.com/eris-ltd/eris-db.js) currently has an issue for exposing 
the permission transaction type, but that work is not yet complete. 
Finally, because these are just convenience wrappers around the RPC call, 
you can formulate the RPC call from any language. 
Here's our [tutorial](https://docs.erisindustries.com/tutorials/tool-specific/eris_by_curl/).

2. Currently the only way to send bond transactions is via 
the eris-pm tooling about or via the raw RPC calls as provided 
in the [eris-by-curl](https://docs.erisindustries.com/tutorials/tool-specific/eris_by_curl/) example. We'd love to accept a pull request 
which exposes this for other users of our javascript tooling.

3.  The peer sergeant major is simple an abstraction. In the config.toml 
for your eris chain you can update the `seed` field to be any node 
which is running the chain. So you can mix and match validators, 
simple peer servers as you wish. 

> https://support.erisindustries.com/helpdesk/tickets/261 