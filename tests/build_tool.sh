#!/bin/bash
set -e

start=`pwd`
base=github.com/eris-ltd/eris-cli
repo=$GOPATH/src/$base
if [ "$CIRCLE_BRANCH" ]
then
  repo=`pwd`
else
  repo=$GOPATH/src/github.com/eris-ltd/eris-cli
fi

testimage="quay.io/eris/eris"
release_min=$(cat version/version.go | tail -n 2 | head -n 1 | cut -d \  -f 4 | tr -d '"')
release_maj=$(echo $release_min | cut -d . -f 1-2)

cd $repo
if [[ "$BRANCH" = "master" ]]
then
  docker build -t $testimage:docker18 -f tests/Dockerfile-1.8 .
  docker build -t $testimage:latest .
  docker tag -f $testimage:latest $testimage:$release_maj
  docker tag -f $testimage:latest $testimage:$release_min
  docker tag -f $testimage:latest $testimage:master
else
  docker build -t $testimage:docker18 -f tests/Dockerfile-1.8 .
  docker build -t $testimage:$BRANCH .
fi
test_exit=$?
cd $start
exit $test_exit
