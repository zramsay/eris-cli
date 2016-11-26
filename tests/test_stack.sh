#!/usr/bin/env bash

# ---------------------------------------------------------------------------
# PURPOSE

# This script will test the eris stack and the connection between eris cli
# and eris pm. **Generally, it should not be used in isolation.**

# ---------------------------------------------------------------------------
# REQUIREMENTS

# Docker installed locally
# Docker-Machine installed locally (if using remote boxes)
# eris' test_machines image (if testing against eris' test boxes)
# Eris installed locally

# ---------------------------------------------------------------------------
# USAGE

# test_stack.sh

# ----------------------------------------------------------------------------
# Set definitions and defaults

# Where are the Things
start=`pwd`
base=github.com/eris-ltd/eris-cli
repo=$GOPATH/src/$base
if [ "$CIRCLE_BRANCH" ] # TODO add windows/osx
then
  repo=${GOPATH%%:*}/src/github.com/${CIRCLE_PROJECT_USERNAME}/${CIRCLE_PROJECT_REPONAME}
  ci=true
elif [ "$TRAVIS_BRANCH" ]
then
  ci=true
  osx=true
elif [ "$APPVEYOR_REPO_BRANCH" ]
then
  ci=true
  win=true
else
  ci=false
fi

export ERIS_PULL_APPROVE="true"
export ERIS_MIGRATE_APPROVE="true"
export SKIP_BUILD="true"

epm=eris-pm
epm_repo=https://github.com/eris-ltd/$epm.git
epm_dir=$repo/../$epm
epm_branch=${EPM_BRANCH:=master}

# ----------------------------------------------------------------------------
# Utility functions

check_and_exit() {
  if [ $test_exit -ne 0 ]
  then
    cd $start
    exit $test_exit
  fi
}

# ----------------------------------------------------------------------------
# Run ECM tests

tests/test_chains_make.sh
test_exit=$?
check_and_exit
cd $start

## ----------------------------------------------------------------------------
## Get EPM
#
#echo
#if [ -d "$epm_dir" ]; then
#  echo "eris-pm present on host; not cloning"
#  cd $epm_dir
#else
#  echo -e "Cloning eris-pm to:\t\t$epm_dir:$epm_branch"
#  git clone $epm_repo $epm_dir &>/dev/null
#  cd $epm_dir 1>/dev/null
#  git checkout origin/$epm_branch &>/dev/null
#fi
#echo
#
## ----------------------------------------------------------------------------
## Run EPM tests
#
#tests/test.sh
#test_exit=$?
#check_and_exit
#cd $start

# ----------------------------------------------------------------------------
# Cleanup
cd $start
exit $test_exit
