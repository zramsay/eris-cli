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

# test_linux.sh [machine]

# ----------------------------------------------------------------------------
# Set definitions and defaults

# Where are the Things
base=github.com/eris-ltd/eris
repo=$GOPATH/src/$base
if [ "$TRAVIS_BRANCH" ]
then
  ci=true
  osx=true
elif [ "$APPVEYOR_REPO_BRANCH" ]
then
  ci=true
  win=true
else
  ci=false
fi

BRANCH=${CIRCLE_BRANCH:=master}
BRANCH=${BRANCH/-/_}
BRANCH=${BRANCH/\//_}

# Define now the tool tests within the Docker container will be booted from docker run
entrypoint="$GOPATH/src/github.com/eris-ltd/eris/tests/test_tool.sh"
testimage=quay.io/eris/eris
testuser=eris
remotesocket=2376
hostsocket=6732
dm_path=".docker/machine"
script="docker.sh"
strt=`pwd`
echo $strt
#import docker machine logic

source $repo/tests/machines/docker_machine.sh

# ----------------------------------------------------------------------------
# Define how tests will run

setup_tests() {
  echo
  echo "Hello! I'm marmot that sets up the docker machine."
  echo
  if [ $linux ]
  then
    build_eris $BRANCH &
    build_result=$!
  else
    echo "Skipping docker build."
    build_result=$!
  fi
  machi_result=$?
  wait $build_result
  build_result=$?
  check_build
}

perform_tests() {
  echo
  sh "$entrypoint"
  test_exit=$?
  if [ "$DOCKER_MACHINE" = true ]
  then
    log_machine $test_exit
  fi
}

cleanup_tests(){
  remove_machine
  echo
  echo
  echo "Your summary good human...."
  printf '%s\n' "${MACH_RESULTS[@]}"
  cd $strt
  exit $test_exit
}

# ---------------------------------------------------------------------------
# Get the things build and dependencies turned on
echo "Hello! I'm the testing suite for eris."

if [ "$DOCKER_MACHINE" = true ]
then
  setup_tests
fi
# ---------------------------------------------------------------------------
# Go!

perform_tests

# ---------------------------------------------------------------------------
# Clean up

cleanup_tests
