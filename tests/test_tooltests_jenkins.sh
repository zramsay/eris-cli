#!/usr/bin/env bash

# ---------------------------------------------------------------------------
# PURPOSE

# This script will test the eris tool itself, including its packages.

# ---------------------------------------------------------------------------
# REQUIREMENTS

# Go installed locally
# Docker installed locally
# Monax installed locally

# ---------------------------------------------------------------------------
# USAGE

# test_gotests_jenkins.sh

# ---------------------------------------------------------------------------
# Defaults and variables
start=`pwd`
job_name="$JOB_NAME-$BUILD_NUMBER"
test_exit=0

# ---------------------------------------------------------------------------
# Define the tests and passed functions
tests() {
  run_test initialize
  run_test util
  run_test config
  run_test loaders
  run_test perform
  run_test data
  run_test files
  run_test services
  run_test chains
  run_test keys
  run_test pkgs
  run_test clean
}

# ---------------------------------------------------------------------------
# Local test utility functions
run_test() {
  go test -cover -timeout 20m ./$1/... && passed $1
  if [ $? -ne 0 ]; then test_exit=1; fi
}

# ---------------------------------------------------------------------------
# Globaly utility functions
checks() {
  if [ "$CLI_REPO" = "" ]
  then
    echo "Cannot run without CLI_REPO being set"
    exit 1
  fi
  if [ "$MONAX_CLI_TESTS_PORT" = "" ]
  then
    echo "Cannot run without MONAX_CLI_TESTS_PORT being set"
    exit 1
  fi
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
  eris clean --yes --containers --images --scratch --dir
  eris version
  eris init --yes
}

passed() {
  if [ $? -eq 0 ]
  then
    echo
    echo "*** Congratulations! *** $1 Package Level Tests Have Passed for job: $job_name"
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
    echo "Congratulations! All Package Level Tests Passed."
    echo "Job: $job_name is green."
    echo
  else
    echo
    echo "Boo :( A Package Level Test has failed."
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
  cd $CLI_REPO
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
