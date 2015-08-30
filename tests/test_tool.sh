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
cd $repo

# ---------------------------------------------------------------------------
# Define the tests and passed functions

packagesToTest() {

  # The first run of tests expect ipfs to be running
  eris services start ipfs
  if [ $? -ne 0 ]; then return 1; fi
  ERIS_IPFS_HOST="http://$(docker inspect --format='{{.NetworkSettings.IPAddress}}' eris_service_ipfs_1)"
  if [ $? -ne 0 ]; then return 1; fi
  export ERIS_IPFS_HOST
  sleep 3 # give node time to boot

  # Start the first series of tests
  cd perform && go test
  passed Perform
  if [ $? -ne 0 ]; then return 1; fi
  cd ../util && go test
  passed Util
  if [ $? -ne 0 ]; then return 1; fi
  cd ../data && go test
  passed Data
  if [ $? -ne 0 ]; then return 1; fi
  cd ../files && go test
  passed Files
  if [ $? -ne 0 ]; then return 1; fi
  cd ../config && go test
  passed Config
  if [ $? -ne 0 ]; then return 1; fi

  # The second series of tests expects ipfs to not be running
  eris services stop ipfs -rx
  if [ $? -ne 0 ]; then return 1; fi

  # Start the second series of tests
  cd ../services && go test
  passed Services
  if [ $? -ne 0 ]; then return 1; fi
  cd ../chains && go test
  passed Chains
  if [ $? -ne 0 ]; then return 1; fi
  cd ../actions && go test
  passed Actions
  if [ $? -ne 0 ]; then return 1; fi
  cd ../contracts && go test
  passed Contracts
  if [ $? -ne 0 ]; then return 1; fi
  # cd ../projects && go test
  # passed Projects
  # if [ $? -ne 0 ]; then return 1; fi
  # cd ../remotes && go test
  # passed Remotes
  # if [ $? -ne 0 ]; then return 1; fi
  cd ../commands && go test
  passed Commands
  if [ $? -ne 0 ]; then return 1; fi

  return 0
}

passed() {
  if [ $? -eq 0 ]
  then
    echo ""
    echo ""
    echo "Congratulations! $1 Package Level Tests Have Passed"
    echo ""
    echo ""
    return 0
  else
    return 1
  fi
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
  docker-machine start $machine
  until [[ $(docker-machine status $machine) == "Running" ]] || [ $ping_times -eq 5 ]
  do
     ping_times=$[$ping_times +1]
     sleep 3
  done
  if [[ $(docker-machine status $machine) != "Running" ]]
  then
    exit 1
  fi
  echo "Machine Started."
  echo "Connecting to Machine."
  eval "$(docker-machine env $machine)"
  echo "Connected to Machine."
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
  eris init -dp
else
  eris init -dp --machine $machine
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
      cd $1 && go test && passed $1
    fi
  else
    packagesToTest
  fi
fi
test_exit=$(echo $?)
set -e

if [ $test_exit -eq 0 ]
then
  echo "Congratulations! All Package Level Tests Passed."
  echo ""
else
  echo ""
  echo "Boo :( A Package Level Test has failed."
  echo ""
fi

# ---------------------------------------------------------------------------
# Cleaning up

if [[ $machine != "eris-test-local" ]]
then
  echo "Stopping Machine."
  docker-machine kill $machine
  echo "Machine Stopped."
fi

cd $start
exit $test_exit