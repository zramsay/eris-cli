#!/usr/bin/env bash
set -e

base=github.com/eris-ltd/eris-cli
repo=$GOPATH/src/$base

if [ $1 ]
then
  echo "Assuming locally."
  machine="eris-test-local"
else
  echo "Will use a machine."
  machine=$MACHINE_NAME
fi

start=`pwd`
cd $repo

passed() {
  echo ""
  echo ""
  echo "$1 Passed."
  echo ""
  echo ""
}

packagesToTest() {

  # The first run of tests expect ipfs to be running
  eris services start ipfs
  ERIS_IPFS_HOST="http://$(docker inspect --format='{{.NetworkSettings.IPAddress}}' eris_service_ipfs_1)"
  export ERIS_IPFS_HOST
  sleep 3
  cd perform && go test
  passed Perform
  cd ../util && go test
  passed Util
  cd ../data && go test
  passed Data
  cd ../files && go test
  passed Files
  cd ../config && go test
  passed Config

  # The second series of tests expects ipfs to not be running
  eris services stop ipfs -rx
  cd ../services && go test
  passed Services
  cd ../chains && go test
  passed Chains
  cd ../actions && go test
  passed Actions
  cd ../contracts && go test
  passed Contracts
  # cd ../projects && go test
  # passed Projects
  # cd ../remotes && go test
  # passed Remotes
  cd ../commands && go test
  passed commands

}

echo ""
echo ""
echo "Testing Shall Commense."
echo ""
echo ""

if [[ "$machine" == "eris-test-local" ]]
then
  echo ""
  echo "Connecting to (local) Machine: $machine"
  echo ""
else
  echo ""
  echo "Connecting to Machine: $machine"
  eval "$(docker-machine env $machine)"
  echo ""
fi

echo ""
echo "Docker API Information"
echo ""
docker version
echo ""

if [[ $machine == "eris-test-local" ]]
then
  echo ""
  eris init -d
else
  echo ""
  eris init -d --machine $machine
fi
passed Setup

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

echo "Congratulations! All Tests Passed. We're Green"
echo ""
echo ""

cd $start