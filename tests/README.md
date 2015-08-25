# Eris CLI Tool Testing Philosophy

This is a hard tool to test. In order to clearly define (and limit) the testing of this this tool we must first understand what the tests will attempt to achieve. Goals which are outside of this realm will be considered as time and resources become available to do so (this is an open source project afterall).

# Goals of Eris CLI Testing Suite

* test the tool itself
* test minimum viable stack connection and sequencing
* test the tool against a multitude of docker-engine APIs
  * docker 1.7.1 (implemented)
  * docker 1.8.0 (implemented)
  * docker 1.8.1 (implemented)
* test the tool from a multitude of host environments
  * debian (not implemented)
  * rhel (not implemented)
  * ubuntu (circle CI provides natively)
  * OSX (not implemented)
  * windows (not implemented)

## Goal 1: Test The Tool Itself

Testing the tool itself is performed from the `tests/test_tool.sh` script. This script should only concern itself with the mechanisms for testing the tool.

## Goal 2: Test The Minimum Viable Stack Connection and Sequencing

TODO (will fill in as we collaboratively build [eris-dl](https://github.com/eris-ltd/eris-dl)).

## Goal 3: Test the Tool Against a Multitue of docker-engine APIs

Testing at this level is performed from the `tests/test.sh` script. This script should be run from each host environment.

## Goal 4: Test the Tool From a Multitude of Host Environments

Currently hosts must be manually built. There is a rudimentary setup script (for Ubuntu) in `tests/host_provision.sh`. It could be fleshed out. Eventually this should be turned into one of the cloud_init happy yamls.

# How To Use It

## Step 1: Define The Hosts (or, Machines)

Hosts should have docker built and available locally in a relatively "clean" OS environment (what "clean" means will be dependent on the host itself).

The set of things which connect into the host (which docker calls a "machine"), including things like the ssh keys, the ssl certs, etc., should be cleanly defined and to the greatest extent possible portable.

What I have done is I have packaged these machine connection definitions into a docker container we keep in a private repository. Billings (our build robot) has access to the correct repository and manages the acquisition of these machine connections (which we would not expose). There is a further layer of protection here in that the machine definitions rely generally on an API keyset to the hosting provider (in our case Digital Ocean) which we are able to place into an Environment Variable into Circle (or another place should that become a challenge) but that is unneccessary usually as we generally just want to be able to power on or off a set of docker "backends" to talk to (more on that later).

Machines should be given a naming convention. The eris naming convention will be as follows:

```
machine=eris-test-$swarm-$ver
```

This convention makes it much easier to pass machine on or off instructions around the testing suite as necessary.

When the test suite boots up (meaning the full `tests/test.sh` is ran), the suite will get the machine connections container, make sure that is accessible to the main portion of the test, which focuses on running the eris tool (step 3) tests against the definitive backends list (step 2) for that eris <-> host connection.

Circle takes care of this for ubuntu. But we still will have to figure out how to broaden this. This connection should be local to the machine, as we test the tool <-> docker connection in the next step.

The test suite will turn on and off docker-machines as needed.

## Step 2: Define The APIs (or, Engines)

On the machines run engines. For our purposes, the engine we're talking about is the docker-engine API. As different people will have different API structures, we need to test against a range of API backends. This array of backends is kept definitively in the `tests/test.sh`.

Before running the tool's tests (step 3) against a specific API backend, the test suite will `docker-machine start` that machine. This is generally done over SSH connection using the SSH keys kept in the test_machines image which the tool suite will make sure is available locally when it performs `docker run erisindustres/test_machines`. The eris cli docker container has docker-machine built into it so that it can work with the machine definitions.

Finally, the test suite will not exit itself when the tool tests (step 3) exit as it assumes it will be looping through an array.

To run from eris connected to docker on the host (instead of connecting to APIs on the ) the `tests/test.sh` can take an additional argument. If `tests/test.sh local` is called. Then the tool will only be tested against the hosts' docker connection. If `tests/test.sh all` is called then the tool will be tested against both the local docker connection as well as the set of API remote machines. By default, the tool will only test against the API remote machines as the host connections *really should* only matter for linux boxes. Windows and OSX both use docker-machine which has the same exact interface as is used to test the APIs.

## Step 3: Test The Tool

The tool tests are managed via the `tests/test_tool.sh` file. This file should generally be ran from inside a container. This is why the test master script will build it and then run that script from inside the container.

The `tests/test_tool.sh` will run through the go tests in the package. It shows were the package level testing is done.

`local` can be passed to the test_tool to run the full test_tool.sh suite *outside* of containers.

Any single package name can be given to the script. So, for example to run the full package tests for the chains package, you would `tests/test_tool.sh chains`. To get to the unit tests within the package will require go test.

**N.B.** -- The difference between `tests/test.sh local` and `tests/test_tool.sh local` is that the former (`tests/test.sh`) will test the tool *inside* a container. The latter will test the tool on the host (or, *outside* a container).

## Step 4: Test the Tool's Packages

To document.

Generally you can increase the visibility by changing the logLevel in the start up script. Be default (e.g., when you PR) it should be `0`.

# Tips

Get inside the container:

```
docker run -it --rm --entrypoint "/bin/bash" -e MACHINE_NAME="eris-test-local" -v /var/run/docker.sock:/var/run/docker.sock --user eris eris/eris
```

To manually run the tool's test script inside the container...

```
docker run --rm --entrypoint "/home/eris/test_tool.sh" -e MACHINE_NAME="eris-test-local" -v /var/run/docker.sock:/var/run/docker.sock --user eris eris/eris
```

# ToDos

* Figure out what breaks when OSX is 'local'
* Figure out what to have circle report (relates to next)
* Get the metrics on the array of hosts and api versions into a consumable form
* Figure out the deployment paths for build artifacts
* moar package level testing
* finalize the dapp level testing