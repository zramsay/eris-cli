#!/usr/bin/env bash
# ----------------------------------------------------------
# PURPOSE

# This is the build script for the eris stack. It will
# build the tool into docker containers in a reliable and
# predicatable manner.

# ----------------------------------------------------------
# REQUIREMENTS

# docker installed locally

# ----------------------------------------------------------
# USAGE

# build_tool.sh

# ----------------------------------------------------------

NAME=eris-pm
IMAGE=quay.io/eris/pm

set -e

if [ "$JENKINS_URL" ] || [ "$CIRCLE_BRANCH" ]
then
  REPO=`pwd`
  CI="true"
else
  REPO=$GOPATH/src/github.com/eris-ltd/$NAME
fi

release_min=$(cat $REPO/version/version.go | tail -n 1 | cut -d \  -f 4 | tr -d '"')
release_maj=$(echo $release_min | cut -d . -f 1-2)

# Build
docker build -t $IMAGE:build $REPO
docker run --rm --entrypoint cat $IMAGE:build /usr/local/bin/$NAME > $REPO/$NAME
docker build -t $IMAGE:$release_min -f Dockerfile.deploy $REPO

# Cleanup
rm $REPO/$NAME
if [ "$CI" ]
then
  docker rmi $IMAGE:build
fi

# Extra Tags
if [[ "$branch" = "master" ]]
then
  docker tag -f $IMAGE:$release_min $IMAGE:$release_maj
  docker tag -f $IMAGE:$release_min $IMAGE:latest
fi

if [ "$CIRCLE_BRANCH" ]
then
  docker tag -f $IMAGE:$release_min $IMAGE:latest
fi
