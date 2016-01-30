#!/usr/bin/env bash

# ---------------------------------------------------------------------------
# PURPOSE

# **DEPRECATED**

# ---------------------------------------------------------------------------
# REQUIREMENTS

# Docker installed locally

# ---------------------------------------------------------------------------
# USAGE

# test_osx.sh

# ----------------------------------------------------------------------------
# Set definitions and defaults

# Where are the Things
base=github.com/eris-ltd/eris-cli
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

declare -a MACHINES
declare -a MACH_RESULTS=()

# Define now the tool tests within the Docker container will be booted from docker run
dm_path=".docker/machine"
strt=`pwd`

# ----------------------------------------------------------------------------
# Utility functions

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

check_build() {
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
}

# ----------------------------------------------------------------------------
# Machine management functions

# make the machines
sort_machines() {
  echo "Getting machines sorted."
  sleeper &
  ticker=$!
  cd $repo/tests/machines
  if [ "$ci" = true ]
  then
    go run setup.go 1>/dev/null
    MACHINES=( $(docker-machine ls -q) )
  else
    MACHINES=( $(go run setup.go) )
  fi
  kill $ticker
  wait $ticker 2>/dev/null
  cd $repo
  for machine in "${MACHINES[@]}"
  do
    if [ ! -e "$HOME/$dm_path/machines/$machine/server.pem" ]
    then
      return 1
    fi
  done
  echo
  echo "Machines sorted."
  return 0
}

# Adds the results for a particular box to the MACH_RESULTS array
#   which is displayed at the end of the tests.
log_machine() {
  if [ "$test_exit" -eq 0 ]
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
}

# ----------------------------------------------------------------------------
# Define how tests will run

run_test() {
  echo "Starting Eris Tests."
  tests/test_tool.sh
}

perform_tests() {
  echo
  for machine in "${MACHINES[@]}"
  do
    export MACHINE_NAME=$machine
    run_test
    test_exit=$?
    log_machine
  done
}

# ---------------------------------------------------------------------------
# Get the things build and dependencies turned on

echo "Hello! I'm the testing suite for eris."
echo
sort_machines
machi_result=$?

check_build

# ---------------------------------------------------------------------------
# Go!

perform_tests

# ---------------------------------------------------------------------------
# Cleaning up

remove_machines
echo
echo
echo "Your summary good human...."
printf '%s\n' "${MACH_RESULTS[@]}"
cd $strt
exit $test_exit
