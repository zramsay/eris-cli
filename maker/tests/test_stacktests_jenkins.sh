#!/usr/bin/env bash
# ----------------------------------------------------------
# PURPOSE

# This is the test manager for ecm. It will run the testing
# sequence for ecm using docker.

# ----------------------------------------------------------
# REQUIREMENTS

# eris installed locally

# ----------------------------------------------------------
# USAGE

# test.sh

# ----------------------------------------------------------
# Defaults and variables
start=`pwd`
job_name="$JOB_NAME-$BUILD_NUMBER"

was_running=0
test_exit=0
chains_dir=$HOME/.eris/chains

# ---------------------------------------------------------------------------
# Define the tests and passed functions
tests(){
  echo "Running Tests..."
  echo
  echo "simplest test"
  uuid=$(get_uuid)
  direct=""
  eris chains make $uuid --account-types=Full:1
  run_test
  if [ $test_exit -eq 1 ]
  then
    return 1
  fi
  echo

  echo "more complex flags test"
  uuid=$(get_uuid)
  direct="$uuid"_validator_000
  eris chains make $uuid --account-types=Root:2,Developer:2,Participant:2,Validator:1
  run_test
  if [ $test_exit -eq 1 ]
  then
    return 1
  fi
  echo

  echo "chain-type test"
  uuid=$(get_uuid)
  direct=""
  eris chains make $uuid --chain-type=simplechain
  run_test
  if [ $test_exit -eq 1 ]
  then
    return 1
  fi
  echo

  echo "add a new account type test"
  uuid=$(get_uuid)
  direct=""
  cp $CM_REPO/tests/fixtures/tester.toml $chains_dir/account-types/.
  eris chains make $uuid --account-types=Test:1
  run_test
  if [ $test_exit -eq 1 ]
  then
    return 1
  fi
  rm $chains_dir/account-types/tester.toml
  echo

  echo "add a new chain type test"
  uuid=$(get_uuid)
  direct="$uuid"_full_000
  cp $CM_REPO/tests/fixtures/testchain.toml $chains_dir/chain-types/.
  eris chains make $uuid --chain-type=testchain
  run_test
  if [ $test_exit -eq 1 ]
  then
    return 1
  fi
  rm $chains_dir/chain-types/testchain.toml
  echo

  echo "export/inspect tars"
  uuid=$(get_uuid)
  direct=""
  eris chains make $uuid --account-types=Full:2 --tar
  if [ $? -ne 0 ]
  then
    test_exit=1
    return 1
  fi
  tar -xzf $chains_dir/$uuid/"$uuid"_full_000.tar.gz -C $chains_dir/$uuid/.
  run_test
  if [ $test_exit -eq 1 ]
  then
    return 1
  fi
  echo

  # export/inspect zips
  # todo

  echo "make a chain using csv test"
  uuid=$(get_uuid)
  direct=""
  eris chains make $uuid --account-types=Full:1
  if [ $? -ne 0 ]
  then
    test_exit=1
    return 1
  fi
  rm $chains_dir/$uuid/genesis.json
  prev_dir=`pwd`
  gen=$(eris chains make $uuid --known --accounts accounts.csv --validators validators.csv)
  echo "$gen" > $chains_dir/$uuid/genesis.json
  run_test
  cd $prev_dir
  if [ $test_exit -eq 1 ]
  then
    return 1
  fi
  echo
}

# ---------------------------------------------------------------------------
# Local Test Utility Functions
run_test(){
  if [ $? -ne 0 ]
  then
    test_exit=1
    return 1
  fi
  echo "Running test ..."
  dir_to_use=$chains_dir/$uuid/$direct
  echo "New-ing chain from $dir_to_use"
  eris chains start $uuid --init-dir $uuid/$direct
  if [ $? -ne 0 ]
  then
    test_exit=1
    return 1
  fi

  echo "Checking test ..."
  sleep 3 # let 'er boot
  check_test
  if [ $? -ne 0 ]
  then
    test_exit=1
  fi

  echo "Cleaning test ..."
  eris chains stop --force $uuid
  eris chains rm --file --data $uuid &>/dev/null
  rm -rf $HOME/.eris/scratch/data/$uuid
  rm -rf $chains_dir/$uuid
}

check_test(){
  # check chain is running
  echo "Checking chain running..."
  chain=( $(eris chains ls --quiet --running | grep $uuid) )
  if [ ${#chain[@]} -ne 1 ]
  then
    echo "chain does not appear to be running"
    echo
    ls -la $dir_to_use
    test_exit=1
    return 1
  fi

  # check results file exists
  echo "Checking accounts.csv ..."
  if [ ! -e "$chains_dir/$uuid/accounts.csv" ]
  then
    echo "accounts.csv not present"
    ls -la $chains_dir/$uuid
    pwd
    ls -la $chains_dir
    test_exit=1
    return 1
  fi

  # check genesis.json
  echo "Checking genesis.json ..."
  genOut=$(cat $dir_to_use/genesis.json | sed 's/[[:space:]]//g')
  genIn=$(eris chains plop $uuid genesis | sed 's/[[:space:]]//g')
  if [[ "$genOut" != "$genIn" ]]
  then
    test_exit=1
    echo "genesis.json's do not match"
    echo
    echo "expected"
    echo
    echo -e "$genOut"
    echo
    echo "received"
    echo
    echo -e "$genIn"
    echo
    echo "difference"
    echo
    diff  <(echo "$genOut" ) <(echo "$genIn")
    return 1
  fi

  # check priv_validator
  echo "Checking priv_validator.json ..."
  privOut=$(cat $dir_to_use/priv_validator.json | tr '\n' ' ' | sed 's/[[:space:]]//g' | sed 's/,\"last_height\":[[:digit:]]\+,\"last_round\":[[:digit:]]\+,\"last_step\":[[:digit:]]\+//g' )
  privIn=$(eris data exec $uuid "cat /home/eris/.eris/chains/$uuid/priv_validator.json" | tr '\n' ' ' | sed 's/[[:space:]]//g' | sed 's/,\"last_height\":[[:digit:]]\+,\"last_round\":[[:digit:]]\+,\"last_step\":[[:digit:]]\+//g' )
  if [[ "$privOut" != "$privIn" ]]
  then
    test_exit=1
    echo "priv_validator.json's do not match"
    echo
    echo "expected"
    echo
    echo -e "$privOut"
    echo
    echo "received"
    echo
    echo -e "$privIn"
    echo
    echo "difference"
    echo
    diff  <(echo "$privOut" ) <(echo "$privIn")
    return 1
  fi
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

  echo "There was an error duing setup; keys were not properly imported. Exiting."
  if [ "$was_running" -eq 0 ]
  then
    eris services stop -r keys
  fi
  exit 1
}

# ----------------------------------------------------------------------------
# Global functions
checks() {
  echo
  echo "Hello! I'm the marmot that tests the eris-cm tooling"
  echo
}

enviro() {
  echo "Testing against"
  echo -e "\tSlave node:\t$NODE_NAME"
  echo -e "\tJob name:\t$JOB_BASE_NAME"
  echo -e "\tJob number:\t$BUILD_ID"
  echo -e "\tCLI branch:\t$CLI_BRANCH"
  echo -e "\tCM branch:\t$CM_BRANCH"
  echo
  go version
  echo
  docker version
  echo
  eris clean --yes --containers --images --scratch
  eris version
  eris init --yes --testing
}

setup(){
  echo "Getting Setup"
  ensure_running keys
  echo "Setup complete"
}

build() {
  echo "Getting Built"
  release_tag=$(cat ~/.eris/eris.toml | grep CM | cut -d ':' -f 2 | sed -e 's/"//')
  echo "Overwriting init'ed image with build job artifacts for $CM_IMAGE:$release_tag"
  docker rmi $CM_IMAGE:$release_tag 1>/dev/null
  lzip -cd $CM_REPO/$CM_NAME.tar.lz | docker import - $CM_IMAGE:$release_tag
  if [ $? -ne 0 ]; then eris clean --yes --containers --images --scratch; exit 1; fi
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
    echo "Congratulations! All CM Tests Passed."
    echo "Job: $job_name is green."
    echo
  else
    echo
    echo "Boo :( A CM Test has failed."
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
  cd $CM_REPO
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
