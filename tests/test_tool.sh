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
ver=$APIVERSION
swarm=$SWARM
ping_times=0
regn_times=0
declare -a images

# If an arg is passed to the script we will assume that only local
#   tests will be ran.
if [ $1 ]
then
  machine="eris-test-local"
  swarm="local"
  ver=$(docker version --format="{{.Client.Version}}")
else
  machine=$MACHINE_NAME
fi

start=`pwd`
declare -a images
declare -a checks

cd $repo

export ERIS_PULL_APPROVE="true"

# ---------------------------------------------------------------------------
# Define the tests and passed functions

announce() {
  echo ""
  echo ""
  echo "Testing against"
  echo -e "\tDocker version:\t$ver"
  echo -e "\tIn Data Center:\t$swarm"
  echo -e "\tMachine name:\t$machine"
  echo ""
}

connect() {
  if [[ "$machine" != eris-test-osx* ]] && [[ "$machine" != eris-test-win* ]] && [[ "$machine" != "eris" ]]
  then
    echo "Starting Machine."
    docker-machine start $machine 1>/dev/null
    until [[ $(docker-machine status $machine) == "Running" ]] || [ $ping_times -eq 10 ]
    do
       ping_times=$[$ping_times+1]
       sleep 3
    done
    if [[ $(docker-machine status $machine) != "Running" ]]
    then
      echo "Could not start the machine. Exiting this test."
      echo
      early_exit
    else
      echo "Machine Started. Regenerating the Certificates."
      sleep 15
      until [ $regn_times -ge 10 ]
      do
        docker-machine regenerate-certs -f $machine &>/dev/null && break
        regn_times=$[$regn_times+1]
        sleep 3
      done
      if [ $regn_times -ge 10 ]
      then
        echo "There was an error connecting to the machine. Exiting test."
        echo
        early_exit
      fi
    fi
    connect_machine
    clear_stuff
  else
    connect_machine
  fi
}

early_exit(){
  docker-machine kill $machine &>/dev/null
  test_exit=1
  report
  cd $start
  exit $test_exit
}

connect_machine(){
  echo "Connecting to Machine."
  eval "$(docker-machine env $machine)" &>/dev/null
  echo "Connected to Machine."
  echo
}

setup_machine() {
  if [[ $machine != "eris-test-local" ]]
  then
    eris init --yes --pull-images=true --testing=true
    echo
    eris_version=$(eris version --quiet)
  fi
}

set_procs() {
  checks[$1]=$!
}

wait_procs() {
  for chk in "${!checks[@]}"
  do
    wait ${checks[$chk]}
  done
}

passed() {
  if [ $? -eq 0 ]
  then
    echo ""
    echo ""
    echo "*** Congratulations! *** $1 Package Level Tests Have Passed on Machine: $machine"
    echo ""
    echo ""
    return 0
  else
    return 1
  fi
}

packagesToTest() {
  go test ./initialize/...
  passed Initialize
  if [ $? -ne 0 ]; then return 1; fi
  go test ./util/...
  passed Util
  if [ $? -ne 0 ]; then return 1; fi
  go test ./config/...
  passed Config
  if [ $? -ne 0 ]; then return 1; fi
  go test ./perform/...
  passed Perform
  if [ $? -ne 0 ]; then return 1; fi
  go test ./data/...
  passed Data
  if [ $? -ne 0 ]; then return 1; fi
  if [[ "$machine" != eris-test-win* ]]
  then
    go test ./files/...
    passed Files
    if [ $? -ne 0 ]; then return 1; fi
    go test ./services/... # switch FROM me if needing to debug
    # cd services && go test && cd .. # switch to me if needing to debug
    passed Services
    if [ $? -ne 0 ]; then return 1; fi
    go test ./chains/... # switch FROM me if needing to debug
    # cd chains && go test && cd .. # switch TO me if needing to debug
    passed Chains
    if [ $? -ne 0 ]; then return 1; fi
  fi
  go test ./keys/...
  passed Keys
  if [ $? -ne 0 ]; then return 1; fi
  go test ./contracts/...
  passed Contracts
  if [ $? -ne 0 ]; then return 1; fi
  go test ./actions/...
  passed Actions
  if [ $? -ne 0 ]; then return 1; fi
  # go test ./projects/...
  # passed Projects
  # if [ $? -ne 0 ]; then return 1; fi
  # go test ./remotes/...
  # passed Remotes
  # if [ $? -ne 0 ]; then return 1; fi
  # The final push....
  go test ./commands/...
  passed Commands
  if [ $? -ne 0 ]; then return 1; fi

  # Now! Stack based tests
  if [[ "$( dirname "${BASH_SOURCE[0]}" )" == "$HOME" ]]
  then
    $HOME/test_stack.sh
  else
    tests/test_stack.sh
  fi
  passed Stack
  return $?
}

clear_stuff() {
  echo "Clearing images and containers."
  docker rm $(docker ps -a -q) &>/dev/null
  docker rmi -f $(docker images -q) &>/dev/null
  echo ""
}

turn_off() {
  echo "Cleaning up after ourselves."
  clear_stuff
  echo "Containers and Images cleanup complete."
  echo "Stopping Machine."
  set +e
  docker-machine kill $machine
  set -e
  echo "Machine Stopped."
}

report() {
  if [ $test_exit -eq 0 ]
  then
    echo ""
    echo "Congratulations! All Package Level Tests Passed."
    echo "Machine: $machine is green."
    echo ""
  else
    echo ""
    echo "Boo :( A Package Level Test has failed."
    echo "Machine: $machine is red."
    echo ""
  fi
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
fi

# Once machine is turned on, display docker information
echo ""
echo "Docker API Information"
echo ""
docker version
echo ""

# Init eris with debug flag to check the connection to docker backend
echo ""
echo "Checking the Eris <-> Docker Connection"
echo ""
eris version
echo
setup_machine
passed Setup

# Perform package level tests run only if eris init ran without problem
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

if [[ $machine != "eris-test-local" ]]
then
  turn_off
fi

report

cd $start
exit $test_exit
