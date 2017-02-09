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
base=github.com/eris-ltd/eris
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

time tests/test_chains_make.sh
test_exit=$?
check_and_exit
cd $start

# ----------------------------------------------------------------------------
# Run EPM tests

time tests/test_jobs.sh
test_exit=$?
check_and_exit
cd $start

# ----------------------------------------------------------------------------
# Cleanup
cd $start
exit $test_exit
