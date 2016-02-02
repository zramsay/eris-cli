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

ecm=eris-cm
ecm_repo=https://github.com/eris-ltd/$ecm.git
ecm_dir=$repo/../$ecm
ecm_branch=${ECM_BRANCH:=master}

epm=eris-pm
epm_repo=https://github.com/eris-ltd/$epm.git
epm_dir=$repo/../$epm
epm_branch=${EPM_BRANCH:=master}

mindy_repo=https://github.com/eris-ltd/mindy.git
mindy_dir=$repo/../mindy
mindy_branch=${MINDY_BRANCY:=master}

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
# Get ECM

echo
if [ -d "$ecm_dir" ]; then
  echo "eris-cm present on host; not cloning"
  cd $ecm_dir
else
  echo -e "Cloning eris-cm to:\t\t$ecm_dir:$ecm_branch"
  git clone $ecm_repo $ecm_dir &>/dev/null
  cd $ecm_dir 1>/dev/null
  git checkout origin/$ecm_branch &>/dev/null
fi
echo

# ----------------------------------------------------------------------------
# Run ECM tests

tests/test.sh
test_exit=$?
check_and_exit
cd $start

# ----------------------------------------------------------------------------
# Get EPM

echo
if [ -d "$epm_dir" ]; then
  echo "eris-pm present on host; not cloning"
  cd $epm_dir
else
  echo -e "Cloning eris-pm to:\t\t$epm_dir:$epm_branch"
  git clone $epm_repo $epm_dir &>/dev/null
  cd $epm_dir 1>/dev/null
  git checkout origin/$epm_branch &>/dev/null
fi
echo

# ----------------------------------------------------------------------------
# Run EPM tests

tests/test.sh
test_exit=$?
check_and_exit
cd $start

# ----------------------------------------------------------------------------
# Get Mindy

echo
if [ -d "$mindy_dir" ]; then
  echo "mindy present on host; not cloning:$mindy_branch"
  cd $mindy_dir
else
  echo -e "Cloning mindy to:\t\t$mindy_dir"
  git clone $mindy_repo $mindy_dir &>/dev/null
  cd $mindy_dir 1>/dev/null
  git checkout origin/$mindy_branch &>/dev/null
fi
echo

# ----------------------------------------------------------------------------
# Run Mindy tests

tests/test.sh
test_exit=$?
check_and_exit
cd $start

# ----------------------------------------------------------------------------
# Cleanup
cd $start
exit $test_exit
