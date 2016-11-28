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


# Other variables
if [[ "$(uname -s)" == "Linux" ]]
then
  uuid=$(cat /proc/sys/kernel/random/uuid | tr -dc 'a-zA-Z0-9' | fold -w 12 | head -n 1)
elif [[ "$(uname -s)" == "Darwin" ]]
then
  uuid=$(uuidgen | tr -dc 'a-zA-Z0-9' | fold -w 12 | head -n 1  | tr '[:upper:]' '[:lower:]')
else
  uuid="62d1486f0fe5"
fi

was_running=0
test_exit=0
chains_dir=$HOME/.eris/chains
name_base=eris-pm-tests
chain_name=$name_base-$uuid
name_full="$chain_name"_full_000
name_part="$chain_name"_participant_000
chain_dir=$chains_dir/$chain_name


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

  echo "There was an error during setup; keys were not properly imported. Exiting."
  if [ "$was_running" -eq 0 ]
  then
    if [ "$ci" = true ]
    then
      eris services stop keys
    else
      eris services stop -rx keys
    fi
  fi
  exit 1
}

test_setup(){
  echo "Getting Setup"
  ensure_running keys

  # make a chain
  eris clean -y
  eris chains make --account-types=Full:1,Participant:1 $chain_name #1>/dev/null
  key1_addr=$(cat $chain_dir/addresses.csv | grep $name_full | cut -d ',' -f 1)
  key2_addr=$(cat $chain_dir/addresses.csv | grep $name_part | cut -d ',' -f 1)
  key2_pub=$(cat $chain_dir/accounts.csv | grep $name_part | cut -d ',' -f 1)
  echo -e "Default Key =>\t\t\t\t$key1_addr"
  echo -e "Backup Key =>\t\t\t\t$key2_addr"
  eris chains start $chain_name --init-dir $chain_dir/$name_full 1>/dev/null
  sleep 5 # boot time
  chain_ip=$(eris chains inspect $chain_name NetworkSettings.IPAddress)
  keys_ip=$(eris services inspect keys NetworkSettings.IPAddress)
  echo -e "Chain at =>\t\t\t\t$chain_ip"
  echo -e "Keys at =>\t\t\t\t$keys_ip"
  echo "Setup complete"
}

goto_base(){
  cd tests/fixtures
}

run_test(){
  # Run the epm deploy
  echo ""
  echo -e "Testing eris-pm using fixture =>\t$1"
  goto_base
  cd $1
  if [ "$ci" = false ]
  then
    echo
    cat readme.md
    echo
    eris pkgs do --chain "$chain_name" --address "$key1_addr" --set "addr1=$key1_addr" --set "addr2=$key2_addr" --set "addr2_pub=$key2_pub" --local-compiler #--debug
  else
    echo
    cat readme.md
    echo
    eris pkgs do --chain "$chain_name" --address "$key1_addr" --set "addr1=$key1_addr" --set "addr2=$key2_addr" --set "addr2_pub=$key2_pub" --rm
  fi
  test_exit=$?

  rm -rf ./abi &>/dev/null
  rm *.bin &>/dev/null
  rm ./jobs_output.json &>/dev/null
  rm ./epm.csv &>/dev/null

  # Reset for next run
  goto_base
  return $test_exit
}

perform_tests(){
  echo ""
  goto_base
  apps=(app*/)
  for app in "${apps[@]}"
  do
    run_test $app

    # Set exit code properly
    test_exit=$?
    if [ $test_exit -ne 0 ]
    then
      failing_dir=`pwd`
      break
    fi
  done
}

test_teardown(){
  if [ "$ci" = false ]
  then
    echo ""
    if [ "$was_running" -eq 0 ]
    then
      eris services stop -rx keys
    fi
    eris chains stop --force $chain_name 1>/dev/null
    # eris chains logs $chain_name -t 200 # uncomment me to dump recent VM/Chain logs
    # eris chains logs $chain_name -t all # uncomment me to dump all VM/Chain logs
    # eris chains logs $chain_name -t all | grep 'CALLDATALOAD\|Calling' # uncomment me to dump all VM/Chain logs and parse for Calls/Calldataload
    # eris chains logs $chain_name -t all | grep 'CALLDATALOAD\|Calling' > error.log # uncomment me to dump all VM/Chain logs and parse for Calls/Calldataload dump to a file
    eris chains rm --data $chain_name 1>/dev/null
    rm -rf $HOME/.eris/scratch/data/$name_base-*
    rm -rf $chain_dir
  else
    eris chains stop -f $chain_name 1>/dev/null
  fi
  echo ""
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
# Setup


echo "Hello! I'm the marmot that tests the eris-pm tooling."
echo
start=`pwd`
cd $repo
test_setup

# ---------------------------------------------------------------------------
# Get the things build and dependencies turned on
if [ "$SKIP_BUILD" != "true" ]
then
  echo
  echo "Building eris-pm in a docker container."
  set -e
  tests/build_tool.sh 1>/dev/null
  if [ $? -ne 0 ]
  then
    echo "Could not build eris-pm. Debug via by directly running [`pwd`/tests/build_outside_tool.sh]"
    exit 1
  fi
  set +e
  echo "Build complete."
  echo ""
fi

# ---------------------------------------------------------------------------
# Go!

if [[ "$1" != "setup" ]]
then
  if ! [ -z "$1" ]
  then
    echo "Running One Test..."
    run_test "$1*/"
  else
    echo "Running All Tests..."
    perform_tests
  fi
fi

# ---------------------------------------------------------------------------
# Cleaning up

if [[ "$1" != "setup" ]]
then
  test_teardown
fi
