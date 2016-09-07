#!/usr/bin/env bash

# ---------------------------------------------------------------------------
# PURPOSE

# This script will test the eris tool itself, including its packages, against
# a given docker backend. If it is started with the "local" argument then
# it will test against the local docker backend. If it is started with an
# argument which is not "local" then it will run the package tests for only
# that package.
#
# Generally, the script will start a given docker-machine backend, make sure
# that it can connect to that machine properly, then it will pull the required
# docker images, run the eris package level tests, then run the eris stack
# level tests, finally it will remove all of the docker containers and images
# so that everything is nice and clean and then shut down the docker-machine.

# ---------------------------------------------------------------------------
# REQUIREMENTS

# Docker installed locally
# Docker-Machine installed locally (if using remote boxes)
# eris' test_machines image (if testing against eris' test boxes)
# Eris installed locally

# ---------------------------------------------------------------------------
# USAGE

# test_tool.sh [local||package]

# ---------------------------------------------------------------------------
# Defaults

start=`pwd`
base=github.com/eris-ltd/eris-cli
repo=$GOPATH/src/$base

# If an arg is passed to the script we will assume that only local
#   tests will be ran.
if [ $1 ]
then
  machine="eris-test-local"
else
  machine=$MACHINE_NAME
fi

start=`pwd`
declare -a checks

cd $repo

export ERIS_PULL_APPROVE="true"
export ERIS_MIGRATE_APPROVE="true"

# ---------------------------------------------------------------------------
# Define the tests and passed functions

announce() {
  echo
  echo "Testing against"
  echo -e "\tMachine name:\t$machine"
  echo
}

connect(){
  echo "Connecting to Machine."
  eval "$(docker-machine env $machine)" &>/dev/null
  echo "Connected to Machine."
  echo
}

setup() {
  if [[ "$machine" == eris-test-win* ]]
  then
    mkdir $HOME/.eris
    touch $HOME/.eris/eris.toml
  fi

  echo "Checking the Host <-> Docker Connection"
  if [ $? -ne 0 ] && [ -z $1 ]
  then
    echo "Could not connect to Docker backend. Attempting to regenerate certificates."
    docker-machine regenerate-certs --force $machine
    connect
    setup "rebuild"
  elif [ $? -ne 0 ] && [ ! -z $1 ]
  then
    flame_out
  fi
  echo "Docker connection established"

  echo "Initializing eris (this may take a few moments)"
  eris init --yes --pull-images=true --testing=true &>/dev/null
  if [ $? -ne 0 ]
  then
    flame_out
  fi
  echo "Eris initialized."

  echo
  echo "Docker API Information"
  echo
  docker version
  if [ $? -ne 0 ]
  then
    flame_out
  fi

  echo
  echo "Checking the Eris <-> Docker Connection"
  echo
  eris version
  if [ $? -ne 0 ]
  then
    flame_out
  fi
}

packagesToTest() {
  if [[ "$SKIP_PACKAGES" != "true" ]]
  then
    go test ./initialize/... && passed Initialize
    if [ $? -ne 0 ]; then return 1; fi
    go test ./util/... && passed Util
    if [ $? -ne 0 ]; then return 1; fi
    go test ./config/... && passed Config
    if [ $? -ne 0 ]; then return 1; fi
    go test ./loaders/... && passed Loaders
    if [ $? -ne 0 ]; then return 1; fi
    go test ./perform/... && passed Perform
    if [ $? -ne 0 ]; then return 1; fi
    go test ./data/... && passed Data
    if [ $? -ne 0 ]; then return 1; fi
    go test ./files/... && passed Files
    if [ $? -ne 0 ]; then return 1; fi
    go test ./services/... && passed Services
    if [ $? -ne 0 ]; then return 1; fi
    go test -timeout=900s ./chains/... && passed Chains
    if [ $? -ne 0 ]; then return 1; fi
    go test ./keys/... && passed Keys
    if [ $? -ne 0 ]; then return 1; fi
    go test ./pkgs/... && passed Packages
    if [ $? -ne 0 ]; then return 1; fi
    # go test ./remotes/... && passed Remotes
    # if [ $? -ne 0 ]; then return 1; fi
    # go test ./apps/... && passed Apps
    # if [ $? -ne 0 ]; then return 1; fi
    # go test ./agent/... && passed Agent
    # XXX the agent test catches epm's error by running through a deploy
    # if [ $? -ne 0 ]; then return 1; fi
    go test ./clean/... && passed Clean
    if [ $? -ne 0 ]; then return 1; fi
  fi
  # The appveyor.yml and circle.yml currently use SKIP_PACKAGES and 
  # SKIP_STACK to parallelize Go and stack tests; otherwise this is 
  # here for faster test runs when needed.
  # Set either variable in your shell before calling `tests/test.sh` 
  # or `tests/test_tool.sh`
  if [[ "$SKIP_STACK" != "true" ]]
  then
    echo "Running Stack Tests"
    echo
    if [[ "$( dirname "${BASH_SOURCE[0]}" )" == "$HOME" ]]
    then
      $HOME/test_stack.sh && passed Stack
    else
      tests/test_stack.sh && passed Stack
    fi
  fi
}

passed() {
  if [ $? -eq 0 ]
  then
    echo
    echo "*** Congratulations! *** $1 Package Level Tests Have Passed on Machine: $machine"
    echo
    return 0
  else
    return 1
  fi
}

report() {
  if [ $test_exit -eq 0 ]
  then
    echo
    echo "Congratulations! All Package Level Tests Passed."
    echo "Machine: $machine is green."
    echo
  else
    echo
    echo "Boo :( A Package Level Test has failed."
    echo "Machine: $machine is red."
    echo
  fi
}

flame_out() {
  echo
  echo "Could not connect to setup for tests. Dumping information =>"
  echo
  ls -la $HOME/.docker/machine/machines/$machine/
  echo
  docker-machine ls
  echo
  docker-machine env $machine
  echo
  unset DOCKER_USER
  unset DOCKER_PASS
  env | grep -i "docker"
  echo
  docker version
  echo
  echo "Exiting. :("
  echo
  exit 1
}

# ---------------------------------------------------------------------------
# Go!
echo "Hello! The marmots will begin testing now."
if [[ "$machine" == "eris-test-local" ]]
then
  announce
else
  announce
  connect
  setup
fi
passed Setup

if [[ $machine == "eris-test-local" ]]
then
  if [[ $1 == "local" ]]
  then
    packagesToTest
  else
    go test ./$1/... && passed $1
  fi
else
  packagesToTest
fi
test_exit=$?

# ---------------------------------------------------------------------------
# Clean up and report

report
cd $start
exit $test_exit
