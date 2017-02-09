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

declare -a MACHINES
declare -a MACH_RESULTS=()

# Define now the tool tests within the Docker container will be booted from docker run
entrypoint="/home/eris/test_tool.sh"
testimage=quay.io/eris/eris
testuser=eris
remotesocket=2376
hostsocket=6732
dm_path=".docker/machine"
strt=`pwd`

# ----------------------------------------------------------------------------
# Utility functions

# this function is to provide simple output to stdOut during the sometimes
# long-ish machine building process. If the timing of this function is
# changed it should be harmonized with the timing of the `timeOutTicker`
# of setup.go
sleeper() {
  sleep 15
  sp="...."
  sc=0
  ticks=0
  # 15 minutes + standard CI 10 minutes will suffice
  until [ $ticks -eq 15 ]
  do
    printf "\b${sp:sc++:1}"
    ((sc==${#sp})) && sc=0
    sleep 60
    ticks=$((ticks + 1))
  done
}

# ----------------------------------------------------------------------------
# Build eris in a docker images

# build docker images; runs as background process
build_eris() {
  echo "Building eris in a docker container."
  cd $repo
  export BRANCH=$1
  tests/build_tool.sh &>/dev/null
  if [ $? -ne 0 ]
  then
    exit 1
  fi
  exit 0
}

# ensure eris image builds; machines built; and machines can be connected to
check_build() {
  if [ "$build_result" -ne 0 ] && [ -z $1 ]
  then
    echo "Could not build eris image. Rebuilding eris image."
    sleep 5
    build_eris $BRANCH &
    wait $!
    build_result=$?
    check_build "rebuild"
  elif [ "$build_result" -ne 0 ] && [ ! -z $1 ]
  then
    echo "Failure building eris image. Debug via by directly running [`pwd`/tests/build_tool.sh]. Exiting tests."
    remove_machines
    exit 1
  fi

  if [ "$machi_result" -ne 0 ] && [ -z $1 ]
  then
    echo "Could not make machines. Rebuilding machines."
    remove_machines
    sleep 5
    sort_machines
    machi_result=$?
    check_build "rebuild"
  elif [ "$machi_result" -ne 0 ] && [ ! -z $1 ]
  then
    echo "Failure making machines. Exiting tests."
    remove_machines
    exit 1
  fi

  check_machines

  if [ $? -ne 0 ] && [ -z $1 ]
  then
    echo "Could not connect to machine(s). Rebuilding machines."
    clear_machine
    remove_machines
    sleep 5
    sort_machines
    machi_result=$?
    check_build "rebuild"
  elif [ $? -ne 0 ] && [ ! -z $1 ]
  then
    echo "Failure connecting to machines. Exiting tests."
    remove_machines
    exit 1
  fi

  clear_machine
  echo "Setup and checks complete."
}

# ----------------------------------------------------------------------------
# Machine management functions

# make the machines
sort_machines() {
  echo "Getting machines sorted."
  sleeper &
  ticker=$!
  cd $repo/tests/machines
  if [ ! -z $1 ]
  then
    MACHINES=( $1 )
  else
    if [ "$ci" = true ]
    then
      go run setup.go 1>/dev/null
      setup_result=$?
      MACHINES=( $(docker-machine ls -q) )
    else
      MACHINES=( $(go run setup.go) )
      setup_result=$?
    fi
  fi
  kill $ticker
  wait $ticker 2>/dev/null
  if [ "$setup_result" -ne 0 ]
  then
    return 1
  fi
  cd $repo
  echo
  echo "Machines sorted."
  return 0
}

# check that we can connect to a machine
check_machines() {
  for machine in "${MACHINES[@]}"
  do
    if [ ! -e "$HOME/$dm_path/machines/$machine/server.pem" ]
    then
      return 1
    fi
  done
  return 0
}

# remove env vars
clear_machine() {
  unset DOCKER_TLS_VERIFY
  unset DOCKER_HOST
  unset DOCKER_CERT_PATH
  unset DOCKER_MACHINE_NAME
}

# test against a single machine; note this should **only** get called by circle. it is only used for
# non current docker versions. current versions docker versions do not utilize this DinD technique
# but rather test in the context of the "host" outer environment against a set backend.
test_tool_in_docker() {
  echo "Starting Eris Docker container."
  if [ "$ci" = true ]
  then
    docker run --name test_tool --volume $HOME/$dm_path:/home/$testuser/$dm_path --entrypoint $entrypoint -e MACHINE_NAME=$machine -p $hostsocket:$remotesocket --user $testuser $testimage:$1 &> $CIRCLE_ARTIFACTS/$1.log
  else
    docker run --name test_tool --rm --volume $HOME/$dm_path:/home/$testuser/$dm_path --entrypoint $entrypoint -e MACHINE_NAME=$machine -p $hostsocket:$remotesocket --user $testuser $testimage:$1 &> $CIRCLE_ARTIFACTS/$1.log
  fi
}

# Adds the results for a particular box to the MACH_RESULTS array
#   which is displayed at the end of the tests.
log_machine() {
  if [ "$1" -eq 0 ]
  then
    MACH_RESULTS+=( "$machine is Green!" )
  else
    MACH_RESULTS+=( "$machine is Red.  :(" )
  fi
}

# remove the machines
remove_machines() {
  if [ "$ci" = true ]
  then
    docker-machine rm --force $(docker-machine ls -q)
  else
    for machine in "${MACHINES[@]}"
    do
      docker-machine rm --force $machine
    done
  fi
  MACHINES=()
}

# ----------------------------------------------------------------------------
# Define how tests will run

setup_tests() {
  echo "Hello! I'm the testing suite for eris."
  echo
  if [ $linux ]
  then
    build_eris $BRANCH &
    build_result=$!
  else
    echo "Skipping docker build."
    build_result=$!
  fi
  sort_machines $1
  machi_result=$?
  wait $build_result
  build_result=$?
  if [ -z $1 ]
  then
    check_build
  fi
}

perform_tests() {
  echo
  for machine in "${MACHINES[@]}"
  do
    docker_cur_machine=$machine
    export MACHINE_NAME=$machine && tests/test_tool.sh
    test_exit=$?
  done

  machine=$docker_cur_machine
  log_machine $test_exit
}

cleanup_tests(){
  remove_machines
  echo
  echo
  echo "Your summary good human...."
  printf '%s\n' "${MACH_RESULTS[@]}"
  cd $strt
  exit $test_exit
}

# ---------------------------------------------------------------------------
# Get the things build and dependencies turned on

setup_tests $1

# ---------------------------------------------------------------------------
# Go!

perform_tests

# ---------------------------------------------------------------------------
# Clean up

cleanup_tests
