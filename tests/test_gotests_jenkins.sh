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
  go test ./initialize/... && passed Initialize
  if [ $? -ne 0 ]; then return 1; fi
  go test ./util/... && passed Util
  if [ $? -ne 0 ]; then return 1; fi
  go test ./config/... && passed Config
  if [ $? -ne 0 ]; then return 1; fi
  go test ./loaders/... && passed Loaders
  if [ $? -ne 0 ]; then return 1; fi
  go test ./perform/... && passed Perform
  if [ $? -ne 0 ]; then return 1; fi
  go test ./data/... && passed Data
  if [ $? -ne 0 ]; then return 1; fi
  go test ./files/... && passed Files
  if [ $? -ne 0 ]; then return 1; fi
  go test ./services/... && passed Services
  if [ $? -ne 0 ]; then return 1; fi
  go test ./chains/... && passed Chains
  if [ $? -ne 0 ]; then return 1; fi
  go test ./keys/... && passed Keys
  if [ $? -ne 0 ]; then return 1; fi
  go test ./pkgs/... && passed Packages
  if [ $? -ne 0 ]; then return 1; fi
  go test ./actions/... && passed Actions
  if [ $? -ne 0 ]; then return 1; fi
  go test ./agent/... && passed Agent
  if [ $? -ne 0 ]; then return 1; fi

  # go test ./remotes/... && passed Remotes
  # if [ $? -ne 0 ]; then return 1; fi
  # go test ./apps/... && passed Apps
  # if [ $? -ne 0 ]; then return 1; fi
  # go test ./update/... && passed Update
  # if [ $? -ne 0 ]; then return 1; fi

  go test ./clean/... && passed Clean
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