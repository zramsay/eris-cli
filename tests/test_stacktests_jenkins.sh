#!/usr/bin/env bash

# ---------------------------------------------------------------------------
# PURPOSE

# This script will test the eris stack and the connection between eris cli
# and eris pm.

# ---------------------------------------------------------------------------
# REQUIREMENTS

# Docker installed locally
# Eris installed locally
# git installed locally
# go installed locally

# ---------------------------------------------------------------------------
# USAGE

# test_stacktests_jenkins.sh

# ----------------------------------------------------------------------------
# Defaults and variables
start=`pwd`
job_name="$JOB_NAME-$BUILD_NUMBER"

cm=eris-cm
cm_repo=https://github.com/eris-ltd/$cm.git
cm_dir=$CLI_REPO/../$cm
cm_branch=${CM_BRANCH:=master}

pm=eris-pm
pm_repo=https://github.com/eris-ltd/$pm.git
pm=$CLI_REPO/../$pm
pm_branch=${PM_BRANCH:=master}

# ---------------------------------------------------------------------------
# Define the tests and passed functions
tests() {
  # ----------------------------------------------------------------------------
  # Get CM
  if [ -d "$cm_dir" ]; then
    echo "cm present on host; not cloning"
    cd $cm_dir
  else
    echo -e "Cloning cm to:\t\t$cm_dir:$cm_branch"
    git clone $cm_repo $cm_dir &>/dev/null
    cd $cm_dir 1>/dev/null
    git checkout origin/$cm_branch &>/dev/null
  fi
  echo

  # ----------------------------------------------------------------------------
  # Run CM tests
  tests/test.sh && passed CM
  if [ $? -ne 0 ]; then return 1; fi
  cd $start

  # ----------------------------------------------------------------------------
  # Get PM
  echo
  if [ -d "$pm" ]; then
    echo "pm present on host; not cloning"
    cd $pm
  else
    echo -e "Cloning pm to:\t\t$pm:$pm_branch"
    git clone $pm_repo $pm &>/dev/null
    cd $pm 1>/dev/null
    git checkout origin/$pm_branch &>/dev/null
  fi
  echo

  # ----------------------------------------------------------------------------
  # Run PM tests
  tests/test.sh && passed PM
  if [ $? -ne 0 ]; then return 1; fi
  cd $start
}

# ----------------------------------------------------------------------------
# Utility functions
checks() {
  if [ "$CLI_REPO" = "" ]
  then
    echo "Cannot run without CLI_REPO being set"
    exit 1
  fi
  # if [ "$ERIS_CLI_TESTS_PORT" = "" ]
  # then
  #   echo "Cannot run without ERIS_CLI_TESTS_PORT being set"
  #   exit 1
  # fi
}

enviro() {
  echo
  echo "Hello! The marmots will begin testing now."
  echo
  echo "Testing against"
  echo -e "\tSlave node:\t$NODE_NAME"
  echo -e "\tJob name:\t$JOB_BASE_NAME"
  echo -e "\tJob number:\t$BUILD_ID"
  echo -e "\tCLI branch:\t$CLI_BRANCH"
  echo
  go version
  echo
  docker version
  echo
  eris clean --yes --all
  eris version
  eris init --yes
}

passed() {
  if [ $? -eq 0 ]
  then
    echo
    echo "*** Congratulations! *** $1 Stack Level Tests Have Passed for job: $job_name"
    echo
    return 0
  else
    return 1
  fi
}

report() {
  if [ $test_exit -eq 0 ]
  then
    echo
    echo "Congratulations! All Stack Level Tests Passed."
    echo "Job: $job_name is green."
    echo
  else
    echo
    echo "Boo :( A Stack Level Test has failed."
    echo "Job: $job_name is red."
    echo
  fi
}

cleanup() {
  eris clean --yes --containers --images --scratch
}

# -------------------------------------------------------------------------
# Go!
main() {
  # run
  cd $ERIS_CLI_REPO
  checks
  enviro
  passed Env
  tests
  test_exit=$?

  # Clean up and report
  cleanup
  report
  cd $start
  exit $test_exit
}

main