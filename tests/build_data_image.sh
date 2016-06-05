#!/usr/bin/env bash
# ----------------------------------------------------------
# PURPOSE

# This is the build script for eris data image.i

# ----------------------------------------------------------
# REQUIREMENTS

# docker installed locally

# ----------------------------------------------------------
# USAGE

# build_data_image.sh

# ----------------------------------------------------------
# Set defaults

if [ "$CIRCLE_BRANCH" ]
then
  repo=`pwd`
else
  repo=$GOPATH/src/github.com/eris-ltd/eris-cli
fi
branch=${CIRCLE_BRANCH:=master}
branch=${branch/-/_}
testimage=${testimage:="quay.io/eris/data"}

release_min=$(cat $repo/version/version.go | tail -n 1 | cut -d \  -f 4 | tr -d '"')
release_maj=$(echo $release_min | cut -d . -f 1-2)

# ---------------------------------------------------------------------------
# Go!
cd $repo/data

if [[ "$branch" = "master" ]]
then
  docker build -t $testimage:latest .
  docker tag $testimage:latest $testimage:$release_maj
  docker tag $testimage:latest $testimage:$release_min
else
  docker build -t $testimage:$release_min .
fi
