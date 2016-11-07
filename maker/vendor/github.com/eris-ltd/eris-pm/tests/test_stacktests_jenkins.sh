#!/usr/bin/env bash
# ----------------------------------------------------------
# PURPOSE

# This is the test manager for eris-pm. It will run the testing
# sequence for eris-pm using docker.

# ----------------------------------------------------------
# REQUIREMENTS

# eris installed locally

# ----------------------------------------------------------
# USAGE

# test.sh [setup]

# ----------------------------------------------------------
# Defaults and variables
start=`pwd`
job_name="$JOB_NAME-$BUILD_NUMBER"

was_running=0
test_exit=0
chains_dir=""
name_base=""
uuid=""
chain_name=""
name_full=""
name_part=""
chain_dir=""

# ---------------------------------------------------------------------------
# Define the tests and passed functions
tests() {
  tests_setup
  if [ $? -ne 0 ]; then tests_teardown; return 1; fi
  goto_base
  apps=(app*/)
  for app in "${apps[@]}"
  do
    run_test $app
    if [ $test_exit -ne 0 ]
    then
      echo "Failure during testing $app."
      echo "Skipping remainder of tests."
      break
    fi
  done
  tests_teardown
}

# ---------------------------------------------------------------------------
# Local Test Utility Functions
run_test() {
  echo
  echo -e "Testing eris-pm using fixture =>\t$1"
  goto_base
  cd $1
  echo
  cat readme.md
  echo
  eris pkgs do --local-compiler --chain "$chain_name" --address "$key1_addr" --set "addr1=$key1_addr" --set "addr2=$key2_addr" --set "addr2_pub=$key2_pub" #--debug
  test_exit=$?
}

tests_setup() {
  # set variables
  chains_dir=$WORKSPACE/.eris/chains
  uuid=$(get_uuid)
  chain_name=$PM_NAME-$GIT_SHORT-$uuid
  name_full="$chain_name"_full_000
  name_part="$chain_name"_participant_000
  chain_dir=$chains_dir/$chain_name

  # make a chain
  eris chains make --account-types=Full:1,Participant:1 $chain_name 1>/dev/null
  if [ $? -ne 0 ]; then return 1; fi
  key1_addr=$(cat $chain_dir/addresses.csv | grep $name_full | cut -d ',' -f 1)
  key2_addr=$(cat $chain_dir/addresses.csv | grep $name_part | cut -d ',' -f 1)
  key2_pub=$(cat $chain_dir/accounts.csv | grep $name_part | cut -d ',' -f 1)
  echo -e "Default Key =>\t\t\t\t$key1_addr"
  echo -e "Backup Key =>\t\t\t\t$key2_addr"

  # boot the chain
  eris chains start $chain_name --init-dir $chain_dir/$name_full 1>/dev/null
  if [ $? -ne 0 ]; then return 1; fi
  sleep 5 # boot time
  echo "Tests Setup complete"
}

tests_teardown() {
  eris chains stop --force $chain_name 1>/dev/null
  eris chains rm --data $chain_name 1>/dev/null
  rm -rf $chain_dir
  echo
  echo "Tests Teardown complete"
}

goto_base() {
  cd $PM_REPO/tests/fixtures
}

# ---------------------------------------------------------------------------
# Local Utility Functions
get_uuid() {
  if [[ "$(uname -s)" == "Linux" ]]
  then
    uuid=$(cat /proc/sys/kernel/random/uuid | tr -dc 'a-zA-Z0-9' | fold -w 12 | head -n 1)
  elif [[ "$(uname -s)" == "Darwin" ]]
  then
    uuid=$(uuidgen | tr -dc 'a-zA-Z0-9' | fold -w 12 | head -n 1)
  else
    uuid="62d1486f0fe5"
  fi
  echo $uuid
}

ensure_running(){
  if [[ "$(eris services ls -qr | grep $1)" == "$1" ]]
  then
    echo "$1 already started. Not starting."
    was_running=1
  else
    echo "Starting service: $1"
    eris services start $1 1>/dev/null
    early_exit
    sleep 3 # boot time
  fi
}

early_exit(){
  if [ $? -eq 0 ]
  then
    return 0
  fi

  echo "There was an error during setup. Exiting."
  if [ "$was_running" -eq 0 ]
  then
    eris services stop -r keys
  fi
  eris clean --yes --containers --images --scratch
  exit 1
}

# ----------------------------------------------------------------------------
# Global functions
checks() {
  echo
  echo "Hello! I'm the marmot that tests the eris-pm tooling"
  echo
}

enviro() {
  echo "Testing against"
  echo -e "\tSlave node:\t$NODE_NAME"
  echo -e "\tJob name:\t$JOB_BASE_NAME"
  echo -e "\tJob number:\t$BUILD_ID"
  echo -e "\tCLI branch:\t$CLI_BRANCH"
  echo -e "\tPM branch:\t$PM_BRANCH"
  echo
  go version
  echo
  docker version
  echo
  eris clean --yes --containers --images --scratch
  eris version
  eris init --yes --testing
  early_exit
}

setup(){
  echo "Getting Setup"
  ensure_running keys
  echo "Setup complete"
}

build() {
  echo "Getting Built"
  release_tag=$(cat ~/.eris/eris.toml | grep PM | cut -d ':' -f 2 | sed -e 's/"//')
  echo "Overwriting init'ed image with build job artifacts for $PM_IMAGE:$release_tag"
  docker rmi $PM_IMAGE:$release_tag 1>/dev/null
  lzip -cd $PM_REPO/"$PM_NAME"-"$GIT_SHORT".tar.lz | docker import - $PM_IMAGE:$release_tag
  early_exit
  docker pull quay.io/eris/compilers:$release_tag &>/dev/null
  early_exit
  echo "Build complete"
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
    echo "Congratulations! All PM Tests Passed."
    echo "Job: $job_name is green."
    echo
  else
    echo
    echo "Boo :( A PM Test has failed."
    echo "Job: $job_name is red."
    echo
  fi
}

cleanup() {
  echo
  if [ "$was_running" -eq 0 ]
  then
    eris services stop -rx keys
  fi
  eris clean --yes --containers --images --scratch
}

# -------------------------------------------------------------------------
# Go!
main() {
  # run
  cd $PM_REPO
  checks
  enviro
  passed Env
  setup
  passed Setup
  build
  passed Build
  tests
  test_exit=$?

  # Clean up and report
  cleanup
  report
  cd $start
  exit $test_exit
}

main
