#!/usr/bin/env bash

# ---------------------------------------------------------------------------
# PURPOSE

# This script will test the eris tool itself, including its packages.

# ---------------------------------------------------------------------------
# REQUIREMENTS

# Go installed locally
# Docker installed locally
# Eris installed locally

# ---------------------------------------------------------------------------
# USAGE

# test_gotests_jenkins.sh

# ---------------------------------------------------------------------------
# Defaults and variables
start=`pwd`
job_name="$JOB_NAME-$BUILD_NUMBER"

# ---------------------------------------------------------------------------
# Define the tests and passed functions
tests() {
  go test -coverpkg ./initialize/... && passed Initialize
  if [ $? -ne 0 ]; then return 1; fi
  go test -coverpkg ./util/... && passed Util
  if [ $? -ne 0 ]; then return 1; fi
  go test -coverpkg ./config/... && passed Config
  if [ $? -ne 0 ]; then return 1; fi
  go test -coverpkg ./loaders/... && passed Loaders
  if [ $? -ne 0 ]; then return 1; fi
  go test -coverpkg ./perform/... && passed Perform
  if [ $? -ne 0 ]; then return 1; fi
  go test -coverpkg ./data/... && passed Data
  if [ $? -ne 0 ]; then return 1; fi
  go test -coverpkg ./files/... && passed Files
  if [ $? -ne 0 ]; then return 1; fi
  go test -coverpkg ./services/... && passed Services
  if [ $? -ne 0 ]; then return 1; fi
  go test -coverpkg ./chains/... && passed Chains
  if [ $? -ne 0 ]; then return 1; fi
  go test -coverpkg ./keys/... && passed Keys
  if [ $? -ne 0 ]; then return 1; fi
  go test -coverpkg ./pkgs/... && passed Packages
  if [ $? -ne 0 ]; then return 1; fi
  go test -coverpkg ./actions/... && passed Actions
  if [ $? -ne 0 ]; then return 1; fi
  go test -coverpkg ./agent/... && passed Agent
  if [ $? -ne 0 ]; then return 1; fi

  # go test -coverpkg ./remotes/... && passed Remotes
  # if [ $? -ne 0 ]; then return 1; fi
  # go test -coverpkg ./apps/... && passed Apps
  # if [ $? -ne 0 ]; then return 1; fi
  # go test -coverpkg ./update/... && passed Update
  # if [ $? -ne 0 ]; then return 1; fi

  go test -coverpkg ./clean/... && passed Clean
  if [ $? -ne 0 ]; then return 1; fi
}

# ---------------------------------------------------------------------------
# Utility functions
checks() {
  if [ "$CLI_REPO" = "" ]
  then
    echo "Cannot run without CLI_REPO being set"
    exit 1
  fi
  if [ "$ERIS_CLI_TESTS_PORT" = "" ]
  then
    echo "Cannot run without ERIS_CLI_TESTS_PORT being set"
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
  eris clean --yes --all
  eris version
  eris init --yes --testing
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
  eris clean --all --yes
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