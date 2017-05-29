#!/usr/bin/env bash
# ----------------------------------------------------------
# PURPOSE

# This is the build script for monax base image.

# ----------------------------------------------------------
# REQUIREMENTS

# docker installed locally

# ----------------------------------------------------------
# USAGE

# build_images.sh

# ----------------------------------------------------------
# Set defaults

if [ "$CIRCLE_BRANCH" ]
then
  repo=`pwd`
else
  repo=$GOPATH/src/github.com/monax/monax
fi

release_min=$(cat $repo/version/version.go | tail -n 1 | cut -d \  -f 4 | tr -d '"')
release_maj=$(echo $release_min | cut -d . -f 1-2)

# ---------------------------------------------------------------------------
# Go!
for name in "base" "build" "data"
do
  image=quay.io/monax/$name
  build_dir=$repo/tests/$name

  mkdir -p $build_dir
  cd $build_dir

  cp $repo/docker/x86/$name/Dockerfile .
  echo
  echo "Building: $image:$release_maj"
  echo
  docker build -t $image:$release_maj .

  rm -rf $build_dir
done
