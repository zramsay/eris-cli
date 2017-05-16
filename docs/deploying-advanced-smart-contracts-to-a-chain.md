---

layout: single
title: "Tutorials | Deploying Advanced Smart Contracts To A Chain"
aliases:
  - /docs/deploying-advanced-smart-contracts
menu:
  tutorials:
    weight: 5

---

## Introduction

<div class="note">
{{% data_sites rename_docs %}}
</div>

For this tutorial, we are going to work with multiple contracts. All the base contract does is get and sets a value, but we'll add some layers to the contract which will satisfy common patterns smart contract writers adopt.

### Contracts Strategy

We are going to use a very simple `get` / `set` contract which sets a variable and gets that same variable. It is about the easiest interactive contract one can imagine and as such we will use that for showing how to work with the Monax platform.

### Make a Chain

Let's make a chain with a few keys on it.

{{< insert_contents 1 "/docs/contracts_deploying_adv/test.sh" 3 10 >}}

Check that it is running with `monax ls`.

## Let's make a more advanced get-set contract sequence.

The first thing we're going to do is to add a very simple contract.

{{< insert_contents 2 "/docs/contracts_deploying_adv/GSFactory.sol" >}}

Now you'll make a file in this directory. Let's assume that is called `GSFactory.sol` and has the following contents displayed above.

This is a slightly more advanced set of contracts than that we used in the [getting started tutorial](/docs/getting-started). Also, now we have multiple contracts we are going to handle.

What do these contracts do? Well, they aren't terribly interesting we know. The first contract, the `GSContract`, merely `gets` and `sets` a value which is an unsigned integer type. The second contract, the `GSFactory`, merely makes a new `GSContract` when `create` is called or it returns the address of the most recent contract created when `getLast` is called.

## Fixup your epm.yaml

Next we need to make an epm.yaml. It should look like this:

{{< insert_contents 3 "/docs/contracts_deploying_adv/epm.yaml" >}}

Now. What does this file mean? Well, this file is the manager file for how to deploy and test your smart contracts. The package manager invoked by `monax pkgs do` will read this file and perform a sequence of `jobs` with the various parameters supplied for the job type. It will perform these in the order they are built into the yaml file.

So let's go through them one by one and explain what each of these jobs are doing. For more on using various jobs [please see the jobs specification](/docs/specs/jobs_specification).

#### Job 1: Deploy Job

This job will compile the `GSFactory.sol` contracts using the compiler service (or run your own locally; which will be covered later in this tutorial). But which contract(s) will get deployed even though they both are in the contract? When we have more than one contract in a file, we tell the package manager which one it should deploy with the `instance` field.

Here we are asking the package manager to deploy `all` of the contracts so that we will have an ABI for the `GSContract` address. This is something important to understand about Factory contracts. Namely that at some point you will have to deploy a "fake" contract to your chain so that the ABI for it is properly saved to the ABI folder.

#### Job 2: Call Job

This job will send a call to the contract. The package manager will automagically use the abi's produced during the compilation process and allow users to formulate contracts calls using the very simple notation of `functionName` `params`.

In this job we are explictly using the `abi` field, which is optional. ABI's are the encoding scheme for how we are able to "talk" to our contracts. When the package manager compiles contracts it will compile their ABIs and save them in the ABI folder (which by default is `./abi` where `./` is wherever your epm.yaml file is). It will save the files using both the name of the contract as well as the address of the contract which is deployed onto the chain.

We explicitly tell the package manager in this call to use the GSFactory ABI. This ABI will be saved as above as the same name of the contract(s) which the package manager finds.

Finally, it is waiting on the call to be sunk into a block before it will proceed.

#### Job 3: Query Contract Job

This job is going to send what are alternatively called `simulated calls` or just `queries` to an accessor function of a contract. In other words, these are `read` transactions. We're selecting the abi to use here based on the job result paradigm. As stated above, ABI's are saved both as the names of the contracts when they are deployed but also as the address of the deployed contract. Since we only deployed one contract our ABI directory will have three files in it: `GSContract`, `GSFactory`, `B995CBBFA3BA0E7DFB0293BA008E0E98A75A53E3` (or whatever the address of the contract was). In this job we are using the result of the `deploy` job.

#### Job 4: Assert Job

This job checks that the contract last deployed matches the return from the create.

#### Job 5: Call Job

This job will send a call to the contract. We explicitly tell the package manager in this call to use the GSContract ABI.

#### Job 6: Query Contract Job

This job gets the value which was set in Job 7

#### Job 7: Assert Job

This job checks that the get and set match.

## Deploy (and Test) The Contract

Now, we are ready to deploy this world changing contract.

{{< insert_contents 4 "/docs/contracts_deploying_adv/test.sh" 11 12 >}}

Note that here we used both the `--address` flag to set the address which we would be using for the jobs and we also set the `setStorageBase` from a flag rather than from a job.

That's it! Your contract is all ready to go. You should see the output in `jobs_output.json` which will have the transaction hash of the transactions as well as the address of the deployed `GSFactory.sol` contract. You can also see your ABI folder:

{{< insert_contents 5 "/docs/contracts_deploying_adv/test.sh" 13 14 >}}

## The compiler?

Where are the contracts compiling? By default they are compiled using a microservice which is automagically turned on when running `monax pkgs do` and subsequently removed. If you'd like to use the remote compiler, specify its URL with the `--remote-compiler` flag.

## Clean Up

Let's clean up after ourselves

{{< insert_contents 6 "/docs/contracts_deploying_adv/test.sh" 15 16 >}}


## [<i class="fa fa-chevron-circle-left" aria-hidden="true"></i> All Tutorials](/docs/)
