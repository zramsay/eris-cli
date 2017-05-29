---

layout: single
title: "Tutorials | Getting Started"
aliases:
  - /docs/getting-started
menu:
  tutorials:
    weight: 5

---

<div class="note">
{{% data_sites rename_docs %}}
</div>

`monax` is the CLI ecosystem application platform built by Monax.

There are four steps need to get moving with Monax:

1. **Install** the platform.
2. **Roll** the blockchain base for your ecosystem application.
3. **Deploy** your ecosystem application using smart contract templates and a simple, web-based user interface.
4. **Integrate** your ecosystem application with a web server or other microservices.

## Step 1. Install the Monax Platform

**Dependencies**: `monax` has 2 dependencies: [Docker](https://www.docker.com/) and for macOS and Windows *only* [Docker Machine](https://docs.docker.com/machine/). Docker is a run anywhere container solution which makes development, deployment, testing, and running of ecosystem applications a breeze and Docker Machine allows you to run Docker on remote machines. We do not currently support Docker for Mac/Windows as they are still in beta.

Currently we consider the most workable setup to be (what our tests consider authoritative) with these operating system and dependencies' versions:

* Host OS = {{< data_coding authoritative_os >}}
* Docker = {{< data_coding docker_auth >}}
* Docker Machine = {{< data_coding docker_machine_auth >}}

We are working steadily toward making `monax` available for a wide variety of host environments.

At the current time, `monax` requires `docker version` >= {{< data_coding docker_min >}} and `docker-machine version` >= {{< data_coding docker_machine_min >}}.
We do not test against older versions of Docker and Docker Machine: `monax` may still work against earlier versions and we can make no guarantees of usability there.

### Linux

Please see the [Docker](https://docs.docker.com/installation/) documentation for how to install it for your Linux distribution.

**Essential**! After you install Docker, you must make sure that the user you are using to develop with `monax` has access to the Docker socket (which is accessible via the `docker` Linux usergroup). When you are logged in as the user you can do this:

```bash
sudo usermod -a -G docker $USER
```

That command will add the current user to the `docker` group which will mean that Docker will not need to be called from `sudo`. After you run that command, then please log out of the current shell and open a new shell. After that `monax` will then be able to connect to Docker.

Make sure that everything is set up with Docker by running (you shouldn't see any errors in the command's output):

```bash
docker version
```

**Note** you will need to make sure that you perform the above command for the *user* which will be running Monax.

If you've also chosen to install Docker Machine, please follow [these](https://docs.docker.com/machine/install-machine/#installing-machine-directly) instructions to install Docker Machine and [these](https://www.virtualbox.org/wiki/Linux_Downloads) to install VirtualBox; then create an Monax virtual machine and run the (`eval`) command:

```bash
docker-machine create -d virtualbox monax
eval $(docker-machine env monax)
```

**Note** Installation of VirtualBox is not a prerequisite, because you may choose to create your virtual machine on Amazon AWS cloud or DigitalOcean, but VirtualBox is what most people use alongside Docker Machine and what we recommend for the Docker Machine setup.

Proceed to one of the package or binary installations below to install the `monax` binary then finalize your setup by running.

```bash
monax init
```

`monax init` will be downloading a few Docker images which may take a few minutes.

#### Debian Package Installation

We have `apt-get` support for most current versions of Ubuntu and Debian Linux:

```bash
{{< data_coding apt >}}
```

#### RPM Package Installation

We have RPM support for most current versions of Fedora, CentOS, and RHEL:

```bash
{{< data_coding yum >}}
```

#### Binary Installation

Alternatively, you can download a release binary for the latest [Release](https://github.com/monax/monax/releases). Make sure you put the binary under one of the paths in the `$PATH` variable and that it has executable permissions:

```bash
curl -L https://github.com/monax/monax/releases/download/v0.16.0/monax_0.16.0-linux-amd64 > monax
chmod +x monax
```

### macOS

We **highly recommend** that you utilize [Homebrew](https://brew.sh) to install `monax`. Docker, Docker Machine, VirtualBox, and `monax` binary will be properly installed with:

```bash
{{< data_coding brew >}}
```

If you are not a `brew` user then please install Docker, Docker machine, and VirtualBox by installing [Docker Toolbox](https://www.docker.com/products/docker-toolbox) and Monax binary from the [Release](https://github.com/monax/monax/releases) page. Make sure you put the binary under one of the paths in your `$PATH` variable and it has executable permissions:

```bash
curl -L https://github.com/monax/monax/releases/download/v0.16.0/monax_0.16.0_darwin_amd64 > monax
chmod +x monax
```

If you don't want to utilize Docker Toolbox, you can install those manually: follow [these](https://docs.docker.com/installation/) instructions to install Docker, [these](https://docs.docker.com/machine/install-machine/#installing-machine-directly) to install Docker Machine, and [these](https://www.virtualbox.org/wiki/Downloads) to install VirtualBox.

If you have chosen not to use Docker Toolbox at all, you need to create an Monax virtual machine and run the (`eval`) command:

```bash
docker-machine create -d virtualbox monax
eval $(docker-machine env monax)
```

**N.B.** At this time Docker for Mac (DFM) and Docker for Windows (DFW), which are still in beta, are not currently supported.

Finalize your setup by running:

```bash
monax init
```

`monax init` will be downloading a few Docker images which may take a few minutes.

### Windows

Install Docker, Docker Machine, and VirtualBox by downloading the [Docker Toolbox](https://www.docker.com/products/docker-toolbox) and Monax binary from the [Release](https://github.com/monax/monax/releases) page.
Make sure you put the binary under one of the paths in your `%PATH%` variable.

If you don't want to utilize Docker Toolbox, you can install those manually: follow [these](https://docs.docker.com/installation/) instructions to install Docker, [these](https://docs.docker.com/machine/install-machine/#installing-machine-directly) to install Docker Machine, and [these](https://www.virtualbox.org/wiki/Downloads) to install VirtualBox.

(You'll want to run `monax` commands either from `git bash` or from the `Docker Quickstart Terminal`, a part of Docker Toolbox. If you prefer to use the `cmd` as your shell, you still can: every command should work as expected, though all the tutorials will assume that you are using the `Docker Quickstart Terminal` and are structured to support **only** that environment.)

If you have chosen not to use Docker Toolbox at all and use `cmd` as your shell, you need to create an Monax virtual machine:

```bash
docker-machine create -d virtualbox default
```

and create a script `setenv.bat` with these contents to be run before your every session with Monax:
```cmd
@echo off

FOR /f "tokens=*" %%i IN ('"docker-machine.exe" env default') DO %%i
```

**Note** -- At this time Docker for Windows (DFW), which is still in beta, is not currently supported.

Finalize your setup by running:

```bash
monax init
```

`monax init` will be downloading a few Docker images which may take a few minutes.

### ARM Installation (IoT devices)

Although we once supported IoT installations, this has been temporarily disabled while the platform undergoes further consolidation. See [this issue](https://github.com/monax/monax/issues/1088) for more details on progress. See also the [deprecated ARM installation tutorial](/docs/deprecated/install-arm).

### Building From Source

If you would like to build from source [see our documentation](/docs/install-source).

### Troubleshooting Your Install

If you have any errors which arise during the installation process, please see our [trouble shooting page](/docs/install-troubleshooting) or join [The Marmot Den](https://slack.monax.io) to ask for help.

## Step 2: Roll Your Own Blockchain in Seconds

If you want to create your blockchain it is two commands:

```bash
monax chains make test_chain
monax chains start test_chain --init-dir ~/.monax/chains/test_chain/test_chain_full_000
```

That `test_chain` can be whatever name you would like it to be. These two commands will create a permissioned, smart contract enabled blockchain suitable for testing.

To check that your chain is running type (running chains have a `*` symbol next to them rather than a `-`):

```bash
monax ls
```

You can peek at chain's logs with these commands (`-f` for "follow"):

```bash
monax chains logs test_chain
monax chains logs -f test_chain
```

Note: although your chain may be "running" (i.e., has been started and has a docker container that is "ON", it is possible that you chain is not making blocks and thus will be useless for deploying contracts. After running the above command, ensure you chain is indeed making blocks. Stay tuned for an `monax chains info/status` command.

Stop your chain:

```bash
monax chains stop test_chain
```

Remove your chain (`-f` to force remove a running chain, `-x` to remove the chain's separate data container which it writes to, and `-d` to remove the (local)  chain directory entirely):

```bash
monax chains rm -xfd test_chain
```

Obviously, you will want an ability to make chains which you properly parameterize. As such you can always type:

```bash
monax chains make --help
```

That's it! Your chain is rolled!

Let's remove all of the monax "stuff" before we move on to the next portion of the tutorials:

```bash
monax clean -yx
```

### Step 2.a: Advanced Chain Making

**Note:** If you'd like to get right into deploying contracts and building your ecosystem application, jump to Step 3 below.

Blockchains are meant to be trustless, and that means everyone generates their own keys. Validators and any other accounts to be included at the inception of a chain must be included in the `genesis.json` file. This is done using the `--known` flag for `monax chains make`. See our [known chain making tutorial](/docs/known-chain-making) for more information. For the purposes of this tutorial, however, we'll be using a simplechain with one account.

To learn about the account types paradigm, try the chain making wizard:

```bash
monax chains make toRemoveLater --wizard
```

This will drop you into an interactive, command line wizard. Follow the text and the prompts to chain making bliss. Since we're going to throw this chain away later you can just press "Enter" at each of the prompts or you can change the variables and get a feel for the wizard.

Once the wizard exits let's take a look at what was created:

```bash
ls ~/.monax/chains/toRemoveLater
```

You should see three `*.csv` files and a bunch of directories. Let's look in one of those directories:

```bash
ls ~/.monax/chains/toRemoveLater/toremovelater_full_000
```

In that directory you should see a `genesis.json`, a `priv_validator.json` and a `config.toml`. The marmots call these a "bundle" as generally they are what is needed to get a chain going.

What about those `csv` files? There should be three of them. Let's take a look:

```bash
cat ~/.monax/chains/toRemoveLater/accounts.csv
cat ~/.monax/chains/toRemoveLater/validators.csv
cat ~/.monax/chains/toRemoveLater/addresses.csv
```

The first two files can be used later to create a new genesis.json if the actual json gets lost. One of the things about this tooling is that it **creates** the keys for you. That is helpful in some circumstances. For production/consortium chains this is not appropriate. See the [known chain making tutorial](/docs/known-chain-making) for more info.

The `monax chains make` tool comes with advanced account type and chain type definition capabilities. More information on complex chain making is included in our [advanced chain making tutorial](/docs/chain-making).

The last file is the `addresses.csv` file which is another artifact of the chain making process. It simply has the addresses and the "names" of the nodes. We find it useful when scripting out complex interactions and it is simply a reference file along the lines of `addr=$(cat $chain_dir/addresses.csv | grep $name | cut -d ',' -f 1)`.

OK, enough playing around let's get serious! Cleaning after our previous experiment:

```bash
monax clean -yx
```

Per the above and after our review of the account types, we know we want to have two Root account types and one Full account type for our new chain. So let's get to business.

```bash
chain_dir=$HOME/.monax/chains/firstchain
chain_dir_this=$chain_dir/firstchain_full_000
```

That will just create a few variables we'll be using in the future. Now, we're ready.

```bash
monax chains make firstchain --account-types=Root:2,Full:1 --unsafe
```

That's it! Let's double check the files to make sure we are squared away.

```bash
ls $chain_dir
ls $chain_dir_this
```

You'll see a `genesis.json`, a `priv_validator.json` and a `config.toml` in `$chain_dir_this`.

#### Step 2.a.3: Instantiate the Blockchain

With all the files prepared we're ready to rock and roll.

```bash
monax chains start firstchain --init-dir $chain_dir_this
```

Check that the chain is running with:

```bash
monax ls
```

You'll see something like:

```bash
CHAIN        ON     VERSION
firstchain   *      0.16.0
```

Note: You can see more information with `monax ls --all`.

To see the logs of the chain:

```bash
monax chains logs firstchain
```

To turn off the chain:

```bash
monax chains stop firstchain
```

Boom. You're all set with your custom built, permissioned, smart contract-ified, chain.

You start your chain up again for the next step:

```
monax chains start firstchain
```

or *remove everything* with:

```bash
monax clean -yx
```

If anything went wrong, please see our trouble shooting guide -> [^1], [^2], [^3], [^4]

## Step 3: Deploy your ecosystem application using smart contract templates

In general we are going to take three steps in order to get our contracts deployed to the blockchain:

1. Write a simple contract
2. Make sure your application package has the proper information
3. Deploy the contracts

#### Contracts Strategy

We are going to use a very simple `get` / `set` contract which sets a variable and gets that same variable. It is about the easiest interactive contract one can imagine and as such we will use that for showing how to work with the Monax platform.

### Step 3.1: Make A Contract for Idi

The first thing we're going to do is to add a very simple contract.

```bash
cd ~/.monax/apps
mkdir idi
cd idi
```

Now you'll make a file in this directory. Let's assume that is called `idi.sol` and has the following contents

{{< insert_contents 1 "/docs/contracts_simple_idi/idi.sol" >}}

What does this contract do? Well, it isn't very interesting, we know. It merely `gets` and `sets` a value which is an unsigned integer type.

### Step 3.2: Fixup your epm.yaml

Next we need to make an `epm.yaml` and make it look something like this:

{{< insert_contents 2 "/docs/contracts_simple_idi/epm.yaml" >}}

Now, what does this file mean? Well, this file is the manager file for how to deploy and test your smart contracts. The package manager invoked by `monax pkgs do` will read this file and perform a sequence of `jobs` with the various parameters supplied for the job type. It will perform these in the order they are built into the yaml file. So let's go through them one by one and explain what each of these jobs are doing. For more on using various jobs [please see the jobs specification](/docs/specs/jobs_specification).

#### Job 1: Set Job

The `set` job simply sets a variable. The package manager includes a naive key value store which can be used for pretty much anything.

#### Job 2: Deploy Job

This job will compile and deploy the `idi.sol` contract using the local compiler service.

#### Job 3: Call Job

This job will send a call to the contract. The package manager will automagically utilize the abi's produced during the compilation process and allow users to formulate contract calls using the very simple notation of `functionName` `params`. The package manager also allows for variable expansion.

So what this job is doing is this. The job is pulling the value of the `$setStorageBase` job (the package manager knows this because it resolved `$` + `jobName` to the result of the `setStorageBase` job) and replacing that with the value, which is `5`. Then it will send that `5` value to the `set` function of the contract which is at the `destination` that is the result of the `deployStorageK` job; in other words the result of Job 3. For more on variables in the package manager, please see the [variables specification](/docs/specs/variable_specification).

#### Job 4: Query Contract Job

This job is going to send what are alternatively called `simulated calls` or just `queries` to an accessor function of a contract. In other words, these are `read` transactions. Generally the `query-contract` is married to an accessor function (such as `get` in the `idi.sol` contract). Usually accessor, or read only functions, in a solidity contracts are denoted as a `constant` function which means that any call sent to the contract will not update the state of the contract.

The value returned from a `query-contract` job then is usually paired with an assert.

#### Job 5: Assert Job

In order to know that things have deployed or gone through correctly, you need to be able to assert relations. The package manager provides you with:

* equality
* non-equality
* greater than or equals (for integers & unsigned integers values only)
* greater than (for integers & unsigned integers values only)
* less than or equals (for integers & unsigned integers values only)
* less than (for integers & unsigned integers values only)

Relations can use either `eq` `ne` `ge` `gt` `le` `lt` syntax, or, in the alternative they can use `==` `!=` `>=` `>` `<=` `<` syntax in the relation field. This is similar to Bash. To make this more explicit we have chosen in the above `epm.yaml` to use the `eq` syntax, but feel free to replace with `==` syntax if you want.

Both the `key` and the `val` (which in other testing frameworks are the `given` and `expect`ed in an assert function) use variable expansion to compare the result of what was supposed to be sent to the `setStorageBase` job (which should have been sent to and stored in the contracts' storage) with what was received from the `queryStorage` job (which in turn called the `get` function of the contract).

### Step 3.3: Deploy (and Test) The Contract

See the Step 2 above if you need to review the chain making process. This series of commands assumed you followed that tutorial and continued here after `monax chains stop firstchain`.

First, let's get our chain turned back on.

```bash
monax ls
```

If it's on, you'll see:

```
CHAIN        ON    VERSION
firstchain  *      0.16.0
```

Whereas if it has been stopped, the `ON` field will have `-` rather than `*`. The same logic applies to services.

If `firstchain` is not running, then turn it on with:

```bash
monax chains start firstchain
```

or make a new chain if firstchain no longer exists.

Now, we are ready to deploy this world changing contract. Make sure you are in the `~/.monax/apps/idi` folder, or wherever you saved your `epm.yaml`. Note that this is a very common pattern in simple contract testing and development; namely to (1) deploy a contract; (2) send it some transactions (or `call`s); (3) query some results from the contract (or `query-contract`s); and (4) assert a result. As you get moving with contract development you will likely find yourself doing this a lot.

```bash
addr=$(cat $chain_dir/addresses.csv | grep firstchain_full_000 | cut -d ',' -f 1)
```

That will make sure we have available the address we would like to use to deploy the contracts. Now we're ready. If the above does not output an address then check your $chain_dir variable and also check that the `firstchain_full_000` variable exists in the addresses.csv.

```bash
monax pkgs do --chain firstchain --address $addr
```

You *should* be able to use any of the addresses you generated during the chainmaking tutorial since they all have the same permission levels on the chain (which, if you followed the simple tutorial are basically all public). If you are using this tutorial outside of the tutorial sequence then you can just give it the address that you'd like to use to sign transactions instead of the `grep firstchain_full_000` bash expansion.

(For those that do not know what is happening in that bash line: `cat` is used to "print the file" and "pipe" it into the second command; `grep` is a finder tool which will find the line which has the right name we want to use; the `cut` says split the line at the `,` and give me the first field).

Note that the package manager can override the account which is used in any single job and/or can set a default `account` job which will establish a default account within the yaml. We find setting the default account within the yaml to usually be counter-productive because others will not be able to easily use your yaml unless they have the same keys in their `monax-keys` (which we **never** recommend). For more on using accounts [please see the jobs specification](/docs/specs/jobs_specification).

Since we have a deployed contract on a running chain, please do take a look at the available options for deployment using the package manager with:

```bash
monax pkgs do --help
```

That's it! Your contract is all ready to go. You should see the output in `jobs_output.json` which will have the transaction hash of the transactions as well as the address of the deployed `idi.sol` contract. The job runner (`monax pkgs do`) can be leveraged for building and interacting with your custom application.


[^1]: If you get an error which looks something like this:

    ```irc
    Performing action. This can sometimes take a wee while
    Post http://chain:46657/status: dial tcp 172.17.0.3:46657: getsockopt: connection refused
    Container monax_interactive_monax_service_idi_tmp_deploy_1 exited with status 1
    ```

    That means that your chain is not started. Please start the chain and give the chain a second to reboot before rerunning the deploy command again. Ensure your chain is making blocks by running `monax chains logs` a few times. The block height should be increasing.

[^2]: If you get an error which looks something like this:

    ```irc
    open /home/monax/.monax/keys/data/1040E6521541DAB4E7EE57F21226DD17CE9F0FB7/1040E6521541DAB4E7EE57F21226DD17CE9F0FB7: no such file or directory
    Container 74a9dbf3d72a2f67e2280bc792e30c7b37fa57e3d04aeb348222f72448bdc84a exited with status 1
    ```

    What is this telling us? Well, it is telling us that it doesn't have the key in the `keys` container. So what you'll want to do is to update with one of the keys you have generated during the prior tutorials.

    To see what keys are currently on your key signing daemon do this:

    ```
    monax keys ls
    ```

    If you do not have any keys then please take the time to generate some keys as described in Step 2 of this tutorial.

[^3]: If you choose the wrong key then you'll get an error which will probably look something like this:

    ```irc
    Error deploying contract idi.sol: unknown account 03E3FAC131CC111D78B569CEC45FA42CE5DA8AD8
    Container edbae127e1a31f1f85fbe14359362f7943028e57dc5eec4d91a71df706f5240f exited with status 1
    ```

    This means that the account `03E3FAC131CC111D78B569CEC45FA42CE5DA8AD8` has not been registered in the `genesis.json`. The account which is not registered will be the same account you told [monax pkgs do] to use via the signing server (`monax-keys`).

    To "see" your `genesis.json` then do this:

    ```
    monax chains exec -it firstchain "cat /home/monax/.monax/chains/firstchain/firstchain_full_000/genesis.json"
    ```

    You can also see your `genesis.json` at `http://localhost:46657/genesis`. Note: replace `localhost` with the output of `docker-machine ip monax` if on OSX or Windows. See our [docker-machine tutorial](/docs/deprecated/using-docker-machine-with-eris) for more information.

[^4]: If the account you are trying to use has not been registered in the `genesis.json` (or, latterly, has not been given the appropriate [permissions](https://github.com/monax/burrow) via permission transactions) and been given the appropriate permissions, then it will not be able to perform the actions it needs to in order to deploy and test the contract. You'll want to make a new chain with the appropriate account types.

    Once you have the following sorted:

    1. The provided account parameter matches a key which is known to the signing daemon; and
    2. The provided account parameter matches an account in the `genesis.json` of a chain;

    Then you'll be ready to:

    ```bash
    monax pkgs do --chain firstchain --address ADDR
    ```

    Where `ADDR` in the above command is the address you want to use.

## [<i class="fa fa-chevron-circle-left" aria-hidden="true"></i> All Tutorials](/docs/)
