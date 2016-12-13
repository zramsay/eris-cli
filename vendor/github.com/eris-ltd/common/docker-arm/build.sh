#!/usr/bin/env bash

declare -a checks

repo_base="quay.io/eris"
tag="arm"

dep=(
  "armbuild/alpine:3.3"
  "armv7/armhf-ubuntu:14.04"
)

tobuild=(
  "base"
  "data"
  "ipfs"
  "btcd"
  "ubuntu"
  "tools"
  ##"geth"
  "node"
  "commonform"
  ##"bitcoincore"
  ##"bitcoinclassic"
  ##"zcash"
)

tobuildscript=(
  "keys"
  ##"compilers"
)

pull_deps() {
  for d in "${dep[@]}"
  do
    echo "Pulling => $d"
    echo ""
    echo ""
    docker pull $d
    echo ""
    echo ""
    echo "Finished Pulling."
  done
}

build_and_push() {
  ele=$1
  # -$- Build options such as --no-cache -$- 
  if [ $# -gt 1 ]; then
    shift
    build_opts=$@
  fi

  echo "Building => $repo_base/$ele:$tag"
  echo ""
  echo ""
  docker build $build_opts -t $repo_base/$ele:$tag $ele 1>/dev/null
  echo ""
  echo ""
  echo "Finished Building."
  echo "Pushing => $ele:$tag"
  echo ""
  echo ""
  docker push $repo_base/$ele:$tag 1>/dev/null
  echo "Finished Pushing."
}

buildscript_and_push() {
  ele=$1
  echo "Building => $repo_base/$ele:$tag"
  echo ""
  echo ""
  cd $ele
  ./build.sh
  cd ..
  echo ""
  echo ""
  echo "Finished Building."
  echo "Pushing => $ele"
  echo ""
  echo ""
  docker push $repo_base/$ele:$tag 1>/dev/null
  echo "Finished Pushing."
}

pull_deps

for ele in "${tobuild[@]}"
do
  set -e
  build_and_push $ele $@
  set +e
done

for ele in "${tobuildscript[@]}"
do
  set -e
  buildscript_and_push $ele
  set +e
done
