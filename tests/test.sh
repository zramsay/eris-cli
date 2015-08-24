#!/usr/bin/env bash
set -e

start=`pwd`
branch=${ERISDB_BUILD_BRANCH:=master}
base=github.com/eris-ltd/eris-cli
repo=$GOPATH/src/$base

cd $repo

passed() {
  echo ""
  echo ""
  echo "$1 Passed."
  echo ""
  echo ""
}

echo "Testing Shall Commense."
echo ""
echo ""
docker version
echo ""
echo ""

# The first run of tests expect ipfs to be running
# eris init
eris services start ipfs
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

echo "Congratulations! All Tests Passed. We're Green"
echo ""
echo ""

cd $start