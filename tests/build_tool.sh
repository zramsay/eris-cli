#!/bin/bash

release_maj="0.10"
release_min="0.10.1"
branch=${CIRCLE_BRANCH:=master}
branch=${branch/-/_}

start=`pwd`
cd $repo

if [[ "$branch" = "master" ]]
then
  docker build -t $testimage:latest .
  docker tag -f $testimage:latest $testimage:$release_maj
  docker tag -f $testimage:latest $testimage:$release_min
else
  docker build -t $testimage:$branch .
fi

cd $start
