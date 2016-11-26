#!/usr/bin/env bash
# ----------------------------------------------------------

# Other variables
repo=`pwd`
was_running=0
test_exit=0
chains_dir=$HOME/.eris/chains

export ERIS_PULL_APPROVE="true"
export ERIS_MIGRATE_APPROVE="true"

# ---------------------------------------------------------------------------
# Needed functionality

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
    if [ "$ci" = true ]
    then
      eris services stop keys
    else
      eris services stop -r keys
    fi
  fi
  exit 1
}

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

test_setup(){
  ensure_running keys
}

check_test(){
  # check chain is running
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
    diff  <(echo "$genOut" ) <(echo "$genIn") | colordiff
    return 1
  fi

  # check priv_validator
  privOut=$(cat $dir_to_use/priv_validator.json | tr '\n' ' ' | sed 's/[[:space:]]//g' | set 's/(,\"last_height\":[^0-9]+,\"last_round\":[^0-9]+,\"last_step\":[^0-9]+//g' )
  privIn=$(eris data exec $uuid "cat /home/eris/.eris/chains/$uuid/priv_validator.json" | tr '\n' ' ' | sed 's/[[:space:]]//g' | set 's/(,\"last_height\":[^0-9]+,\"last_round\":[^0-9]+,\"last_step\":[^0-9]+//g' )
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
    diff  <(echo "$privOut" ) <(echo "$privIn") | colordiff
    return 1
  fi
}

run_test(){
  if [ $? -ne 0 ]
  then
    test_exit=1
    return 1
  fi
  dir_to_use=$chains_dir/$uuid/$direct
  eris chains start $uuid --init-dir $uuid/$direct
  if [ $? -ne 0 ]
  then
    test_exit=1
    return 1
  fi
  sleep 3 # let 'er boot
  check_test
  if [ $? -ne 0 ]
  then
    test_exit=1
  fi
  eris chains stop --force $uuid
  if [ ! "$ci" = true ]
  then
    eris chains rm --data $uuid
  fi
  rm -rf $HOME/.eris/scratch/data/$uuid
  rm -rf $chains_dir/$uuid
}

perform_tests(){
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
  cp $repo/tests/cm_test_fixtures/tester.toml $chains_dir/account-types/.
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
  cp $repo/tests/cm_test_fixtures/testchain.toml $chains_dir/chain-types/.
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

#  # export/inspect zips
#  # todo

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
  gen=$(eris chains make $uuid --known --accounts $chains_dir/$uuid/accounts.csv --validators $chains_dir/$uuid/validators.csv)
  echo "$gen" > $chains_dir/$uuid/genesis.json
  run_test
  cd $prev_dir
  if [ $test_exit -eq 1 ]
  then
    return 1
  fi
  echo
}

test_teardown(){
  if [ "$ci" = false ]
  then
    echo
    if [ "$was_running" -eq 0 ]
    then
      eris services stop -rx keys
    fi
    echo
  fi
  if [ "$test_exit" -eq 0 ]
  then
    echo "Tests complete! Tests are Green. :)"
  else
    echo "Tests complete. Tests are Red. :("
  fi
  cd $start
  exit $test_exit
}

# ---------------------------------------------------------------------------
# Get the things build and dependencies turned on

echo "Hello! I'm the marmot that tests the [eris chains make] command"
start=`pwd`
cd $repo
test_setup
echo

# ---------------------------------------------------------------------------
# Go!

echo "Running Tests..."
perform_tests

# ---------------------------------------------------------------------------
# Cleaning up

test_teardown
