#!/usr/bin/env bash

# ---------------------------------------------------------------------------
# PURPOSE

# This script will test the eris. If it is given the "local" argument, then
# it will run against the local docker backend. If it is not given the local
# argument then it will run against a series of docker backends defined in
# arrays at the top of the script.
#
# What this script will do is first it will define what should be run, then
# it will make sure that it has access to the eris' test machine definition
# image files necessary to connect into the backends. Then it will run the
# test_tool and test_stack scripts against a set of backends. That set will
# either be a random element of a given array for a major docker version, or
# it will run against the entire suite of backends.

# ---------------------------------------------------------------------------
# REQUIREMENTS

# Docker installed locally

# ---------------------------------------------------------------------------
# USAGE

# test.sh [local]

# ----------------------------------------------------------------------------
# Set definitions and defaults

# Where are the Things
base=github.com/eris-ltd/eris-cli
if [ "$CIRCLE_BRANCH" ]
then
  repo=${GOPATH%%:*}/src/github.com/${CIRCLE_PROJECT_USERNAME}/${CIRCLE_PROJECT_REPONAME}
  circle=true
else
  repo=$GOPATH/src/$base
  circle=false
fi
branch=${CIRCLE_BRANCH:=master}
branch=${branch/-/_}
if [[ "$branch" == "" ]]
then
  # get from git
  branch=`git rev-parse --abbrev-ref HEAD`
fi

# Define now the tool tests within the Docker container will be booted from docker run
entrypoint="/home/eris/test_tool.sh"
testimage=quay.io/eris/eris
testuser=eris
remotesocket=2376
localsocket=/var/run/docker.sock
dm_path=".docker/machine"
strt=`pwd`

# Adds the results for a particular box to the machine_results array
#   which is displayed at the end of the tests.
log_results() {
  if [ "$test_exit" -eq 0 ]
  then
    machine_results+=("$machine is Green!")
  else
    machine_results+=("$machine is Red.  :(")
  fi
}

# ----------------------------------------------------------------------------
# Define how tests will run

runTests(){
  if [ $? -ne 0 ]; then return 1; fi

  echo "Starting Eris Docker container."
  if [ "$circle" = true ]
  then
    docker run --volume $HOME/$dm_path:/home/$testuser/$dm_path --entrypoint $entrypoint -e MACHINE_NAME=$machine -p $remotesocket --user $testuser $testimage:$1
  else
    docker run --rm --volume $HOME/$dm_path:/home/$testuser/$dm_path --entrypoint $entrypoint -e MACHINE_NAME=$machine -p $remotesocket --user $testuser $testimage:$1
  fi

  # logging the exit code
  test_exit=$(echo $?)
  log_results
}

# ---------------------------------------------------------------------------
# Get the things build and dependencies turned on

echo "Hello! I'm the testing suite for eris."
echo ""
echo "Building eris in a docker container."
cd $repo
export testimage
export repo
tests/build_tool.sh 1>/dev/null
if [ $? -ne 0 ]
then
  echo "Could not build eris. Debug via by directly running [`pwd`/tests/build_tool.sh]"
  exit 1
fi

echo ""
echo "Getting machines sorted."
cd $repo/tests/machines
machines=( $(go run setup.go) )
if [ $? -ne 0 ]
then
  docker-machine rm $(docker-machine ls --filter -q)
  echo "Failure making machine(s). Exiting Tests."
  printf '%s\n' "${machines[@]}"
  exit 1
fi

# ---------------------------------------------------------------------------
# Go!

echo ""
for machine in "${machines[@]}"
do
  if [[ "$machine" == *1.8* ]]
  then
    runTests "docker18"
  else
    runTests $branch
  fi
done

# ---------------------------------------------------------------------------
# Cleaning up

for machine in "${machines[@]}"
do
  docker-machine rm --force $machine
done
echo ""
echo ""
echo "Your summary good human...."
printf '%s\n' "${machine_results[@]}"
echo ""
echo ""
echo "Done. Exiting with code: $test_exit"
cd $strt
exit $test_exit
