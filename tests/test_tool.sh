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
  echo "Starting Machine."
  if [[ "$machine" != "eris" ]] # "eris" should be used when CI testing against OSX and Windows.
  then
    docker-machine start $machine 1>/dev/null
  fi
  until [[ $(docker-machine status $machine) == "Running" ]] || [ $ping_times -eq 10 ]
  do
     ping_times=$[$ping_times +1]
     sleep 3
  done
  if [[ $(docker-machine status $machine) != "Running" ]]
  then
    echo "Could not start the machine. Exiting this test."
    exit 1
  else
    echo "Machine Started."
    if [[ "$machine" != "eris" ]]
    then
      docker-machine regenerate-certs -f $machine 2>/dev/null
    fi
  fi
  sleep 5
  echo "Connecting to Machine."
  eval "$(docker-machine env $machine)" &>/dev/null
  echo "Connected to Machine."
  echo ""
  clear_stuff
}

setup_machine() {
  export ERIS_PULL_APPROVE="true" #because init now pulls images
  
  if [[ $machine != "eris-test-local" ]]
  then
    eris init --yes --pull-images=true --testing=true
    echo
    eris_version=$(eris version --quiet)
    #pull_images
    #echo "Image Pulling Complete."
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

#done by init
#pull_images() {
#  images=( "quay.io/eris/base" "quay.io/eris/data" "quay.io/eris/ipfs" "quay.io/eris/keys" "quay.io/eris/erisdb:$eris_version" "quay.io/eris/epm:$eris_version" )
#  for im in "${images[@]}"
#  do
#    echo -e "Pulling image =>\t\t$im"
#    docker pull $im 1>/dev/null
    # Async // parallel pulling not working consistently.
    #   see: https://github.com/docker/docker/issues/9718
    # this is fixed in docker 1.9 ONLY. So when we deprecate
    # docker 1.8 we can move to asyncronous pulling
    # docker pull $im 1>/dev/null &
    # set_procs
#  done
  # wait_procs
#}

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

  # For testing we want to override the Greg Slepak required ask before pull ;)
  export ERIS_PULL_APPROVE="true"

  # The first run of tests expect ipfs to be running
  eris services start ipfs
  if [ $? -ne 0 ]; then return 1; fi
  ERIS_IPFS_HOST="http://$(docker inspect --format='{{.NetworkSettings.IPAddress}}' eris_service_ipfs_1)"
  if [ $? -ne 0 ]; then return 1; fi
  export ERIS_IPFS_HOST
  sleep 5 # give ipfs node time to boot

  # Start the first series of tests
  go test ./initialize/...
  passed Initialize
  if [ $? -ne 0 ]; then return 1; fi
  go test ./perform/...
  passed Perform
  if [ $? -ne 0 ]; then return 1; fi
  go test ./util/...
  passed Util
  if [ $? -ne 0 ]; then return 1; fi
  go test ./data/...
  passed Data
  if [ $? -ne 0 ]; then return 1; fi
  go test ./files/...
  passed Files
  if [ $? -ne 0 ]; then return 1; fi
  go test ./config/...
  passed Config
  if [ $? -ne 0 ]; then return 1; fi
  go test ./keys/...
  passed Keys
  if [ $? -ne 0 ]; then return 1; fi

  # The second series of tests expects ipfs to not be running
  eris services stop ipfs -frx
  unset ERIS_IPFS_HOST
  if [ $? -ne 0 ]; then return 1; fi

  # Start the second series of tests
  go test ./services/... -timeout 720s # switch FROM me if needing to debug
  # cd services && go test && cd .. # switch to me if needing to debug
  passed Services
  if [ $? -ne 0 ]; then return 1; fi
  go test ./chains/... # switch FROM me if needing to debug
# cd chains && go test && cd .. # switch TO me if needing to debug
  passed Chains
  if [ $? -ne 0 ]; then return 1; fi
  go test ./actions/...
  passed Actions
  if [ $? -ne 0 ]; then return 1; fi
  go test ./contracts/...
  passed Contracts
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
  set +e
  docker rm $(docker ps -a -q) &>/dev/null
  docker rmi -f $(docker images -q) &>/dev/null
  set -e
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

set -e
echo "Hello! The marmots will begin testing now."
if [[ "$machine" == "eris-test-local" ]]
then
  announce
else
  announce
  connect
fi
set +e

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
