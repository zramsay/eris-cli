---

layout: single
title: "Tutorials | Chain Making"
aliases:
  - /docs/chain-making
menu:
  tutorials:
    weight: 5

---

<div class="note">
   <em>Note: As of 2017, our product has been renamed from Eris to Monax. This documentation refers to an earlier version of the software prior to this name change (<= 0.16). Later versions of this documentation (=> 0.17) will change the <code>eris</code> command and <code>~/.eris</code> directory to <code>monax</code> and <code>~/.monax</code> respectively.</em>
</div>

## Introduction

There are typically two steps to making a permissioned blockchain (for less advanced users we say there are three but really there are two):

1. Make the necessary files
2. Instantiate the blockchain

We shall go through these in their logical order.

### Chain Design

To design our chain we need to, first, consider, *who* will get *what* permissions and *why*. It is outside the scope of this tutorial to outline all of the considerations which would come into play when thinking about creating a [permissioning system](/platform/db), but for the purposes of this tutorial, we will craft the genesis block to use the following paradigm:

* Administrators (these would be developers who had **full** control over the chain, but will **not** be validators on the chain);
* Validators (these will be set up as cloud instances and they will **only** be given validation permissions);
* Developers (who will have **partial** access to the more advanced featurs of the chain, such as the ability to update burrow's name registry and also create contracts on the chain); and
* Participants (these will have permissions to do most of the stuff necessary on the chain from a common participants point of view; they'll be able to send tokens and call contracts).

For the purposes of this tutorial, we will have (1) administrator, (7) validators, (3) developers, and (20) participants. This will require a total of 31 keys, and all of their specifics to be generated and added to the genesis block.

If you would like to understand all of the permissions which a `burrow` smart contract network is capable of providing, [please see its documentation](https://github.com/monax/burrow/blob/master/README.md).

## Step 1. Make the Necessary Files

If you have run through the chain making tool (`monax chains make myChain` with the `--wizard` flag) then you will have been introduced to the idea of account-types. In `monax`, we are not restrictive about what account-types you can use. We expose a wide variety of permissions which you can utilize to add a network level permissioning system to your network of `burrow` clients (see links above). This adds a large amount of complexity to the equation, however, and to simplify the use of permissions, we utilize a layer of abstraction which are `account types`. These account types are simply bundles of permissions and tokens which the `monax chains make` command uses to package up our files for us.

Let's first take a closer look at our account types:

```bash
cd ~/.monax/chains/account-types
ls
```

In this directory you will find a few `*.toml` files. These files each represent a different account type with its bundles of permissions. Let's see what they look like either open the file in your favorite text editor or:

```bash
cat root.toml
```

At the top of the file you will see the description of the account type and other narrative stuff which is consumed by the chain making wizard that is utilized by `monax chains make anotherChain --wizard`.

After the description sections you'll see the following lines:

```toml
default_number = 3
default_tokens = 9999999999
default_bond = 0
```

The first line `default_number` tells the monax chain maker by default how many of the `root` accounts to make. The `default_tokens` tells the monax chain maker how many tokens to give each key which is generated and given to this account type. The `default_bond` tells the monax chain maker by default how many tokens should be bonded by the key which is generated and given to this account type. When `default_bond` is zero, the monax chain maker will not add the account type's key(s) to the genesis.json as validators.

The third section of the toml file is the permissions table. This section looks something like this:

```toml
[perms]
root = 1
send = 1
call = 1
createContract = 1
createAccount = 1
bond = 1
name = 1
hasBase = 1
setBase = 1
unsetBase = 1
setBlobal = 1
hasRole = 1
addRole = 1
rmRole = 1
```

Where a field is `1`, the monax chain maker will turn that permission for the account type `on`; and where it is `0`, the monax chain maker will turn that permission for the account type `off`. To adjust the permissions for a default account type then edit any of the `~/.monax/chains/account-types/*.toml` files as you wish. After that, whenever you run the monax chain maker it will respect the changes to any of the fields.

You can also simply add new account types, which is what we're going to do next. Let's make a copy of the `developer.toml` file and edit it.

```bash
cp developer.toml adv_chain_developer.toml
```

Open `~/.monax/chains/account-types/adv_chain_developer.toml` in your favorite text editor. Make the following changes:

```toml
name = "AdvDeveloper"

...

bond = 1
hasBase = 1
```

What did those changes do? Well the first change should be obvious. For the second change we modified the permission to `bond` and to utilize the `hasBase` functionality from `0` (off) to `1` (on) for this account type. We are not going to use either of these permissions that we changed, this is only to demonstrate how we'd update the account types we're gonna use.

At this point once we're happy with the account types for our chain (feel free to look around at the other account types files if you like; but we're just going to use the defaults for the rest of this tutorial), then we can move on to the next step in the process.

Now we are goint to take a look at `monax`'s chain types feature.

```bash
cd ~/.monax/chains/chain-types
ls
```

In this directory is our chain types. Let's take a look at what a chain types file is:

```bash
cat simplechain.toml
```

Similarly to account types files, the chain types files start with some lines describing the chain type. Then there is a table for the account types (as well as some tables `monax chains` will be utilizing in future versions) which look like this:

```toml
[account_types]
Full = 1
Developer = 0
Participant = 0
Root = 0
Validator = 0
```

That table tells the monax chain maker that when the `--chain-type` flag is utilized how many of each of the account types to make. So let's make a copy of this file and add our own chain type.

```bash
cp simplechain.toml advchain.toml
```

Open `advchain.toml` in your favorite text editor and let's edit it to look like the following:

```toml
# This is a TOML config file.
# For more information, see https://github.com/toml-lang/toml

name = "advchain"

definition = ""

[account_types]
Full = 0
AdvDeveloper = 3
Developer = 0
Participant = 20
Root = 1
Validator = 7

[messenger]

[manager]

[consensus]

```

You can see that we have zeroed out `Full` (which is a root + validator account type useful in simplechain scenarios) and `Developer` and utilized our new account type `AdvDevelop` which we will make three (3) of. The rest of the account types will use the defaults.

Now. After that quick tour we are ready to make the chain.

```bash
cd ~/.monax/chains
monax chains make advchain --chain-type advchain
```

If it paused for a little while then returned you to your terminal that means it was successful. Let's check with:

```bash
ls
```

In your `~/.monax/chains` directory you should now have an `advchain` directory. Let's move into that directory.

```bash
cd advchain
ls
```

Your output should look something like this:

```irc
accounts.csv               advchain_participant_006  advchain_participant_018
accounts.json              advchain_participant_007  advchain_participant_019
addresses.csv              advchain_participant_008  advchain_root_000
advchain_advdeveloper_000  advchain_participant_009  advchain_validator_000
advchain_advdeveloper_001  advchain_participant_010  advchain_validator_001
advchain_advdeveloper_002  advchain_participant_011  advchain_validator_002
advchain_participant_000   advchain_participant_012  advchain_validator_003
advchain_participant_001   advchain_participant_013  advchain_validator_004
advchain_participant_002   advchain_participant_014  advchain_validator_005
advchain_participant_003   advchain_participant_015  advchain_validator_006
advchain_participant_004   advchain_participant_016  validators.csv
advchain_participant_005   advchain_participant_017
```

What are we looking at? Well, we're looking at a bunch of files and directories that we are going to use for starting our chain.

You should look through the files in this directory:

```bash
cat accounts.csv
```

That will output something that looks like this:

```csv
8E8A774435AC9F56B37DB4309ABF40CE9BE5E5BDD9D47C334FB5F3A1A8F2FEDA,9999999999,advchain_advdeveloper_000,14590,16383
4CE6468E8348424D55227F537F65154BA22A58D9F35B3066C9A9B03DE214D64D,9999999999,advchain_advdeveloper_001,14590,16383
D2C5837094738C73B29B95CE2E96E1D3B99EEE1E21D3AFB18DBDB2002E9B2BF2,9999999999,advchain_advdeveloper_002,14590,16383
62FC9E8C47C776321B71E396DD3F3C1BEF0BB70169E9BB2C894122531451DCFD,9999999999,advchain_participant_000,2118,16383
64BD863DD516254B291503306641BECB1EF2BFC27BBCA732C4C6A392B7CDE5D3,9999999999,advchain_participant_001,2118,16383
9080F5835B34EF05E2B6C02E27A6382118F90B51D8DA07CC15421DB9CC1BC329,9999999999,advchain_participant_002,2118,16383
189F5C04CB8151A91B64C805C3C64A2C947DD51E3FD7AE8CC66223BD9C0898D9,9999999999,advchain_participant_003,2118,16383
EA3448CA1FC8DAA696DB4E3D3B4D44C4D2B2C28D587C14A4AB3910138B59A946,9999999999,advchain_participant_004,2118,16383
64D247A9C0DD8BA506757086849256FE8A95CEA9914021C26C7926EA9B93446C,9999999999,advchain_participant_005,2118,16383
671CB5CE2E7B959AA82526C32C5249EE11E1FDA03D701B9E1E96DAB76D47EE8C,9999999999,advchain_participant_006,2118,16383
15E961A0DBCAD230CB4EF419637BD4F0C3869D39580C82D97235C8C91E71A769,9999999999,advchain_participant_007,2118,16383
896125B445EC682DB9509E672D8F4CFF7B9A761DAE11A0BEEF6EB8FCB7C84C1A,9999999999,advchain_participant_008,2118,16383
AA18F8EB0645FF1C400A8375D6E86AC1882EDEA87E03677D23F7BD12D4820431,9999999999,advchain_participant_009,2118,16383
5B6DF82BA53C4233938C277C1B1CA5A5E98FE6A0672C1470718485626036FA1E,9999999999,advchain_participant_010,2118,16383
982B4F5744ECAD949FD9741DC991D2D74F5586BA2F43EE3BEB1E89ABECCC4356,9999999999,advchain_participant_011,2118,16383
9A86EA853704B0AD3B1160CF8B73DD3F4BDD8C6B911FE6BD8D01B96B55D94A1B,9999999999,advchain_participant_012,2118,16383
F46FFB9321BC01B31393ACD0568A06124EE68D4780F3E412980EB1EFB765C667,9999999999,advchain_participant_013,2118,16383
583805A071C6BBD54BBEA7AB588ED65CDF6091EA1A1E81D09BE37722D4252083,9999999999,advchain_participant_014,2118,16383
CCF936A8468086AD2BD07BBA9A1EDAFCB6BF1AA40EEB99B2060A81B13ABBD902,9999999999,advchain_participant_015,2118,16383
FCEC03B2FDBF8F4ABBF67C024125A4A0C4974B87E72CA6A3498B4157596AD396,9999999999,advchain_participant_016,2118,16383
988DDB37CB56D9C5D4D1553D912B2EE9B308E266BCFA3BE9FB4BA5DA1E7BFE38,9999999999,advchain_participant_017,2118,16383
94746737F920A84831318AA782888D0F89A2D08DBDC84006E15B8B053B174396,9999999999,advchain_participant_018,2118,16383
B463365F86522FF4A3C31483CBB1376EB3147B180A3F47DA722BF3280F527EA8,9999999999,advchain_participant_019,2118,16383
FC500CD9311575B1B2D12B72B2F5B37C2E3EFBC9C57D4B4F27190C350E54427E,9999999999,advchain_root_000,16383,16383
A343EBBED1AA05AAB2FD2C3D377FF1A4F0986ECB364D3542050216BD72042311,9999999999,advchain_validator_000,32,16383
4C0DB0C7D3C44963DFCBDAB19271EE2F4F3CD52A7B925DAFBA0A0265E0BCF5BD,9999999999,advchain_validator_001,32,16383
734A10E769FD137B5F4B46423BA8F73DE099FB4A68AD768E933990B711E21325,9999999999,advchain_validator_002,32,16383
22E1D7439C881B7F891DA6E63F12287ECCF04F3F03382D8247DC05478A41C915,9999999999,advchain_validator_003,32,16383
46828EC5120AE670A2D4386780F5034110DFA51B80DA7BFAE01B34795558E190,9999999999,advchain_validator_004,32,16383
24E7BC129F9FDE6D6D31F749A0220F1D827186883B8ED1164273D666C9A3C350,9999999999,advchain_validator_005,32,16383
D1B95DC7AC13786DABE6BE2F6F5217A4276EDE942AC7EA6853DBA5A11E15641C,9999999999,advchain_validator_006,32,16383
```

These are the accounts that will get made on the chain. This csv can later be utilized by `monax chains make --known` to remake a genesis.json if needed. The form of this csv is:

```csv
publicKey,tokens,name,permission,setBase
```

You can see that, e.g., each of the validator nodes has the same `permission` number, that all the accounts have the same `setBase` and all of the tokens given match the defaults set up in the account types files.

**Temporary Hack**

Next let's look at the accounts.json

```bash
cat accounts.json
```

This file is useful for testing integration with `legacy-contracts.js`. Getting `legacy-contracts.js` fully integrated into `monax-keys` is on our roadmap for future releases but at this time it is still needed. As of the 0.16.0 release, this file will **not** have the required `privKey` field for `legacy-contracts.js`. You'll need to add the `--unsafe` flag to `monax chains make`.

**End Temporary Hack**

Now let's look at the addresses.json

```bash
cat addresses.csv
```

This file should be self-explanatory. It simply includes the `address` (which is a hashed version of the public key) and the `name`. This file is useful when combining monax chain maker with the package manager (`monax pkgs`) and for scripting interactions over a given chain.

Finally, let's look at the validators.csv

```bash
cat validators.csv
```

As with the accounts.csv, this is a file which can later be fed into `monax chains make --known` for recreation of the genesis.json. The file looks similar, but distinct from the accounts.json

```csv
A343EBBED1AA05AAB2FD2C3D377FF1A4F0986ECB364D3542050216BD72042311,9999999998,advchain_validator_000,32,16383
4C0DB0C7D3C44963DFCBDAB19271EE2F4F3CD52A7B925DAFBA0A0265E0BCF5BD,9999999998,advchain_validator_001,32,16383
734A10E769FD137B5F4B46423BA8F73DE099FB4A68AD768E933990B711E21325,9999999998,advchain_validator_002,32,16383
22E1D7439C881B7F891DA6E63F12287ECCF04F3F03382D8247DC05478A41C915,9999999998,advchain_validator_003,32,16383
46828EC5120AE670A2D4386780F5034110DFA51B80DA7BFAE01B34795558E190,9999999998,advchain_validator_004,32,16383
24E7BC129F9FDE6D6D31F749A0220F1D827186883B8ED1164273D666C9A3C350,9999999998,advchain_validator_005,32,16383
D1B95DC7AC13786DABE6BE2F6F5217A4276EDE942AC7EA6853DBA5A11E15641C,9999999998,advchain_validator_006,32,16383
```

The form of this csv is:

```csv
publicKey,tokensBonded,name,permission,set_base
```

Note that there are two main differences between the accounts.csv and the validators.csv. (1) Only the validator nodes are in the validator.csv file, and (2) the amount of tokens in the csv file is one less for the validator nodes. This means that all the validator nodes will be given `9999999999` tokens but will bond `9999999998` tokens (leaving them with `1` token unbonded).

Next let's look at the files in all these directories:

```bash
cd advchain_root_000
ls
```

In this directory you should see a `priv_validator.json`. This is the key that will be used by the burrow client. (Note, we are working on moving signing completely out of burrow and completely into monax-keys but this work is not yet finished.)

There is also a genesis.json file that is within the directory. Finally, there is a `config.toml` which, for any multi-node chain will need the `seeds` field filled in. Note that `monax chains make` has `--seeds-ip` field to fill the `seeds` field out automatically.

This directory contains the **minimum** necessary files to start a chain. As we will see soon, there is one file which is lacking to fully run *this* chain.

**N.B.** You will want to export your keys onto the host at this point so that you have them backed up. Run `monax keys export --all` and you'll see the keys on your host by running `monax keys ls` or looking in `~/.monax/keys/data`. 

## Step 2. Instantiate the Blockchain

With all the files made for us by the monax chain maker out we're ready to rock and roll.

Let's start the chain and use our root credentials!

```bash
monax chains start advchain --init-dir ~/.monax/chains/advchain/advchain_root_000
```

Boom. You're all set with your custom built, permissioned, smart contract-ified, blockchain. Except for one thing. This particular chain won't run out of the box though. Why? Because you'll need to deploy the validators and connect them to one another.

Let's take a look at the chain for a minute:

```bash
monax chains logs advchain -f
```

That command will `follow` the logs. To stop following the logs use `ctrl+c`. As you will see, nothing appears to be happening here. This is a feature not a bug.

### A Bit About Validators

Burrow uses the tendermint consensus engine under the hood (on our roadmap is to be able to provide burrow's comprehensive RPC and application manager portion over various consensus engines.

The tendermint consensus engine requires that >2/3 (not >=2/3 !!) of the bonded stake is present in a round of voting in order to add a block to the chain. When we only started one node on this chain, and very much unlike proof of work consensus engines, the chain will not progress by itself. This is because there was only one node on the network and it doesn't actually have any bonded stake. Remember we started the `advchain_root_000` node, which according to the genesis.json and validators.csv file has bonded no stake.

So how do we move this chain forward? Basically we have to start nodes which *do have* bonded stake and connect them together. When the bonded stake "present" on the network is > 2/3 of the total bonded stake then the chain will begin moving forward and blocks will be created.

This trips up a lot of folks when starting to work with proof of stake consensus engines, especially those who are coming from proof of work style consensus engines. In the context of this `advchain` we are making now. We have seven validators, all with equal stake. These could represent seven parties to a deal, two parties to a deal (with control over the validators split amongst the parties) or anything in between. In order for the network to move forward (add blocks to the end of the chain), then >= 5 of the validator nodes will need to be available to the network.

Note that it is **not** the "number of nodes" on the network which matters. Rather it is the **amount bonded** that matters. However for pilots and proofs of concept, we recommend giving validators the same number of tokens bonded and then the "number of nodes" can be used as a proxy for the amount of bonded stake which is present.

So, instead of talking about validators, let's get this chain "turned on"!

But before we do that, let's actually remove the chain for now so it doesn't get in our way.

```bash
eris chains rm -xfd advchain
```


## [<i class="fa fa-chevron-circle-left" aria-hidden="true"></i> All Tutorials](/docs/)
