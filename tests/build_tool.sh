#!/bin/bash

if [ "$CIRCLE_BRANCH" ]
then
  repo=`pwd`
else
  repo=$GOPATH/src/github.com/eris-ltd/eris-cli
fi
branch=${CIRCLE_BRANCH:=master}
branch=${branch/-/_}
testimage=${testimage:="quay.io/eris/eris"}

release_min=$(cat $repo/version/version.go | tail -n 1 | cut -d \  -f 4 | tr -d '"')
release_maj=$(echo $release_min | cut -d . -f 1-2)

start=`pwd`
cd $repo

if [[ "$branch" = "master" ]]
then
  docker build -t $testimage:docker18 -f tests/Dockerfile-1.8 .
  docker build -t $testimage:latest .
  docker tag -f $testimage:latest $testimage:$release_maj
  docker tag -f $testimage:latest $testimage:$release_min
  docker tag -f $testimage:latest $testimage:master
else
  docker build -t $testimage:docker18 -f tests/Dockerfile-1.8 .
  docker build -t $testimage:$branch .
fi

cd $start
