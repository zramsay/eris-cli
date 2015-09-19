#!/usr/bin/env bash
set -e

# ---------------------------------------------------------------------------
# Defaults

base=github.com/eris-ltd/eris-cli
repo=$GOPATH/src/$base
ver=$APIVERSION
swarm=$SWARM
ping_times=0

# If an arg is passed to the script we will assume that only local
#   tests will be ran.
if [ $1 ]
then
  machine="eris-test-local"
else
  machine=$MACHINE_NAME
fi

start=`pwd`
declare -a images
declare -a checks

cd $repo

# ---------------------------------------------------------------------------
# Define the tests and passed functions

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

  # The second series of tests expects ipfs to not be running
  eris services stop ipfs -frx
  unset ERIS_IPFS_HOST
  if [ $? -ne 0 ]; then return 1; fi

  # Start the second series of tests
  go test ./services/...
  passed Services
  if [ $? -ne 0 ]; then return 1; fi
  go test ./chains/...
  # cd chains && go test
  passed Chains
  if [ $? -ne 0 ]; then return 1; fi
  # cd ..
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
  go test ./commands/...
  passed Commands

  # The final push....
  if [ $? -ne 0 ]; then return 1; fi
  return 0
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

set_procs() {
  checks[$1]=$!
}

wait_procs() {
  for chk in "${!checks[@]}"
  do
    wait ${checks[$chk]}
  done
}

pull_images() {
  images=( "eris/base" "eris/data" "eris/ipfs" "eris/keys" "eris/erisdb:$eris_version" )
  for im in "${images[@]}"
  do
    echo -e "Pulling image =>\t\t$im"
    docker pull $im 1>/dev/null &
    set_procs
  done
  wait_procs
}

# ---------------------------------------------------------------------------
# Go!

echo "Hello! The marmots will begin testing now."

if [[ "$machine" == "eris-test-local" ]]
then
  echo ""
  echo ""
  echo "Testing (locally) against"
  echo -e "\tDocker version:\t$ver"
  echo -e "\tIn Data Center:\t$swarm"
  echo -e "\tMachine name:\t$machine"
  echo ""
else
  echo ""
  echo ""
  echo "Testing against"
  echo -e "\tDocker version:\t$ver"
  echo -e "\tIn Data Center:\t$swarm"
  echo -e "\tMachine name:\t$machine"
  echo ""
  echo "Starting Machine."
  docker-machine start $machine &> /dev/null
  until [[ $(docker-machine status $machine) == "Running" ]] || [ $ping_times -eq 10 ]
  do
     ping_times=$[$ping_times +1]
     sleep 3
  done
  if [[ $(docker-machine status $machine) != "Running" ]]
  then
    echo "Could not start the machine. Exiting this test."
    exit 1
  fi
  sleep 5
  echo "Machine Started."
  echo "Connecting to Machine."
  eval "$(docker-machine env $machine)" &>/dev/null
  echo "Connected to Machine."
  echo ""
  echo "Clearing images and containers for tests."
  set +e
  docker rm $(docker ps -a -q) &>/dev/null
  docker rmi $(docker images -q) &>/dev/null
  set -e
  echo ""
fi

# Once machine is turned on, display docker information
echo ""
echo "Docker API Information"
echo ""
docker version
echo ""

# Init eris with debug flag to check the connection to docker backend
set +e
echo ""
echo "Checking the Eris <-> Docker Connection"
echo ""
if [[ $machine == "eris-test-local" ]]
then
  eris init -dp --yes
else
  eris init -dp --yes --machine $machine
  echo
  eris version
  echo
  eris_version=$(eris version | cut -d ':' -f2 | tr -d ' ')
  pull_images
  echo "Image Pulling Complete."
fi
passed Setup

# Perform package level tests run only if eris init ran without problem
if [ $? -eq 0 ]
then
  if [ $1 ]
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
fi
test_exit=$(echo $?)
set -e

# ---------------------------------------------------------------------------
# Clean up and report

if [[ $machine != "eris-test-local" ]]
then
  set +e
  echo "Cleaning up after ourselves."
  docker rm -f $(docker ps -a -q) &> /dev/null
  docker rmi -f $(docker images -q) &> /dev/null
  echo "Containers and Images cleanup complete."
  echo "Stopping Machine."
  docker-machine kill $machine
  echo "Machine Stopped."
  set -e
fi

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

cd $start
exit $test_exit
