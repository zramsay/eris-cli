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

ecm=eris-cm
ecm_repo=https://github.com/eris-ltd/$ecm.git
ecm_dir=$repo/../$ecm
ecm_test_dir=$repo/../$ecm/tests
ecm_branch=${ECM_BRANCH:=master}

epm=eris-pm
epm_repo=https://github.com/eris-ltd/$epm.git
epm_dir=$repo/../$epm
epm_test_dir=$repo/../$epm/tests
epm_branch=${EPM_BRANCH:=master}

# ----------------------------------------------------------------------------
# Get ECM

if [ -d "$ecm_test_dir" ]; then
  cd $ecm_test_dir
else
  git clone $ecm_repo $ecm_dir 1>/dev/null
  cd $ecm_test_dir 1>/dev/null
  git checkout origin/$ecm_branch &>/dev/null
fi

# ----------------------------------------------------------------------------
# Run ECM tests

export ERIS_PULL_APPROVE="true"
eris init --yes --pull-images=true --testing=true
./test.sh
test_exit=$?
if [ $test_exit -ne 0 ]
then
  cd $start
  exit $test_exit
fi
cd $start

# ----------------------------------------------------------------------------
# Get EPM

if [ -d "$epm_test_dir" ]; then
  cd $epm_test_dir
else
  git clone $epm_repo $epm_dir 1>/dev/null
  cd $epm_test_dir 1>/dev/null
  git checkout origin/$epm_branch &>/dev/null
fi

# ----------------------------------------------------------------------------
# Run EPM tests

./test.sh
test_exit=$?
if [ $test_exit -ne 0 ]
then
  cd $start
  exit $test_exit
fi
cd $start

# ----------------------------------------------------------------------------
# Cleanup
exit $test_exit
