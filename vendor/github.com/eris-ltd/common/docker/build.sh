#!/usr/bin/env bash

declare -a checks

repo_base="quay.io/eris"
tag="latest"

dep=(
  "ubuntu:14.04"
  # should match digest in base/Dockerfile (line 1)
  #"alpine@sha256:4b7f27ae8ce4ce6019ce41fa4275296f31b7b730b3eeb5fecf80f1b60959343d"
)

tobuild=(
  # these two images left stable for now until base (ubuntu) is fully deprecated.
  #"base"
  #"build"
  "ipfs"
  "btcd"
  "ubuntu"
  "tools"
  "eth"
  "geth"
  "node"
  "gulp"
  "commonform"
  "sunit_base"
  "embark_base"
  "bitcoincore"
  "bitcoinclassic"
  "parity"
  "openbazaar-server"
  "openbazaar-client"
  "zcash"
)

tobuildscript=(
  "compilers"
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
  echo "Building => $repo_base/$ele:$tag"
  echo ""
  echo ""
  docker build --no-cache -t $repo_base/$ele:$tag $ele 1>/dev/null
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
  echo "Building => $repo_base/$ele"
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
  docker push $repo_base/$ele 1>/dev/null
  echo "Finished Pushing."
}

pull_deps

for ele in "${tobuild[@]}"
do
  set -e
  build_and_push $ele
  set +e
done

for ele in "${tobuildscript[@]}"
do
  set -e
  buildscript_and_push $ele
  set +e
done
