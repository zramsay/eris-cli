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
if [ "$CIRCLE_BRANCH" ]
then
  repo=${GOPATH%%:*}/src/github.com/${CIRCLE_PROJECT_USERNAME}/${CIRCLE_PROJECT_REPONAME}
  circle=true
else
  repo=$GOPATH/src/$base
  circle=false
fi

epm_dir=$repo/../eris-pm
epm_test_dir=$repo/../eris-pm/tests
epm_branch=${EPM_BRANCH:=master}

# ----------------------------------------------------------------------------
# Get EPM

if [ -d "$epm_test_dir" ]; then
  cd $epm_test_dir
else
  git clone https://github.com/eris-ltd/eris-pm.git $epm_dir 1>/dev/null
  cd $epm_test_dir 1>/dev/null
  git checkout origin/$epm_branch &>/dev/null
fi

# ----------------------------------------------------------------------------
# Run EPM tests

./test.sh
test_exit=$?

# ----------------------------------------------------------------------------
# Cleanup

cd $start
exit $test_exit
