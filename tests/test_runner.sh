#!/usr/bin/env bash
# ----------------------------------------------------------

# tests user stories for [monax pkgs do] (soon to be [monax run])

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

# Use the current built target, if it exists
# Otherwise default to system wide executable
COMMIT_SHA=$(git rev-parse --short --verify HEAD)
cli_exec="$GOPATH/src/github.com/monax/monax/target/cli-${COMMIT_SHA}"
if ! [ -e $cli_exec ]
then
  cli_exec="monax"
fi

was_running=0
test_exit=0
chains_dir=$HOME/.monax/chains
name_base="monax-runner-tests"
chain_name=$name_base-$uuid
name_full="$chain_name"_full_000
name_part="$chain_name"_participant_000
chain_dir=$chains_dir/$chain_name
repo=`pwd`

export MONAX_PULL_APPROVE="true"
export MONAX_MIGRATE_APPROVE="true"

# ---------------------------------------------------------------------------
# Needed functionality

ensure_running(){
  if [[ "$($cli_exec ls -format {{.ShortName}} | grep $1)" == "$1" ]]
  then
    echo "$1 already started. Not starting."
    was_running=1
  else
    echo "Starting service: $1"
    $cli_exec services start $1 1>/dev/null
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
      $cli_exec services stop keys
    else
      $cli_exec services stop -r keys
    fi
  fi
  exit 1
}

test_setup(){
  echo "Getting Setup"
  ensure_running keys

  # make a chain
  $cli_exec clean -y
  $cli_exec chains make --account-types=Full:1,Participant:1 $chain_name --unsafe

  key1_addr=$(cat $chain_dir/addresses.csv | grep $name_full | cut -d ',' -f 1)
  key2_addr=$(cat $chain_dir/addresses.csv | grep $name_part | cut -d ',' -f 1)
  key2_pub=$(cat $chain_dir/accounts.csv | grep $name_part | cut -d ',' -f 1)
  echo -e "Default Key =>\t\t\t\t$key1_addr"
  echo -e "Backup Key =>\t\t\t\t$key2_addr"

  $cli_exec chains start $chain_name --init-dir $chain_dir/$name_full
  sleep 5 # boot time
  chain_ip=$($cli_exec chains ip $chain_name)
  keys_ip=$($cli_exec services ip keys)
  echo -e "Chain at =>\t\t\t\t$chain_ip"
  echo -e "Keys at =>\t\t\t\t$keys_ip"
  echo "Setup complete"
}

check_test(){
  # check chain is running
  chain=( $($cli_exec ls --running --format {{.ShortName}} | grep $chain_name) )
  if [ "$chain" != "$chain_name" ]
  then
    echo "chain does not appear to be running"
    echo
    ls -la $dir_to_use
    test_exit=1
    return 1
  fi

  # check that the expected files are there
  # might need to use arguments in here
}

perform_tests(){
  echo
  echo "simple run"
  cd $repo/tests/run_fixtures/simple
  $cli_exec pkgs do --chain $chain_name --address $key1_addr
  cd $repo
  if [ ! -f $repo/tests/run_fixtures/simple/epm.output.json ]
  then
    echo "epm.output.json not found"
    return 1
  fi
  rm $repo/tests/run_fixtures/simple/epm.output.json
  #check_test
  if [ $test_exit -eq 1 ]
  then
    return 1
  fi
  echo

  echo "simple run with --file flag"
  cd $repo/tests/run_fixtures/simple
  $cli_exec pkgs do --chain $chain_name --address $key1_addr --file notpm.yaml
  cd $repo
  if [ ! -f $repo/tests/run_fixtures/simple/notpm.output.json ]
  then
    echo "notpm.output.json not found"
    return 1
  fi
  rm $repo/tests/run_fixtures/simple/notpm.output.json
  #check_test
  if [ $test_exit -eq 1 ]
  then
    return 1
  fi
  echo

  echo "simple run with --output flag"
  cd $repo/tests/run_fixtures/simple
  $cli_exec pkgs do --chain $chain_name --address $key1_addr --output newpm.output.json
  cd $repo
  if [ ! -f $repo/tests/run_fixtures/simple/newpm.output.json ]
  then
    echo "newpm.output.json not found"
    return 1
  fi
  rm $repo/tests/run_fixtures/simple/newpm.output.json
  #check_test
  if [ $test_exit -eq 1 ]
  then
    return 1
  fi
  echo

  echo "simple run with --file and --output flag"
  cd $repo/tests/run_fixtures/simple
  $cli_exec pkgs do --chain $chain_name --address $key1_addr --file notpm.yaml --output newpm.output.json
  cd $repo
  if [ ! -f $repo/tests/run_fixtures/simple/newpm.output.json ]
  then
    echo "newpm.output.json not found"
    return 1
  fi
  rm $repo/tests/run_fixtures/simple/newpm.output.json
  #check_test
  if [ $test_exit -eq 1 ]
  then
    return 1
  fi
  echo

  echo "simple run with --dir flag"
  $cli_exec pkgs do --chain $chain_name --address $key1_addr --dir $repo/tests/run_fixtures/simple
  if [ ! -f $repo/tests/run_fixtures/simple/epm.output.json ]
  then
    echo "epm.output.json not found"
    return 1
  fi
  rm $repo/tests/run_fixtures/simple/epm.output.json
  #check_test
  if [ $test_exit -eq 1 ]
  then
    return 1
  fi
  echo

  echo "simple run with --dir and --file flag"
  $cli_exec pkgs do --chain $chain_name --address $key1_addr --dir $repo/tests/run_fixtures/simple --file MyEPM.yaml
  if [ ! -f $repo/tests/run_fixtures/simple/MyEPM.output.json ]
  then
    echo "MyEPM.output.json not found"
    return 1
  fi
  rm $repo/tests/run_fixtures/simple/MyEPM.output.json
  #check_test
  if [ $test_exit -eq 1 ]
  then
    return 1
  fi
  echo

  echo "ensure blank .json file isn't consumed (#1220)"
  $cli_exec pkgs do --chain $chain_name --address $key1_addr --file $repo/tests/run_fixtures/simple/MyEPM.yaml
  if [ ! -f $repo/tests/run_fixtures/simple/MyEPM.output.json ]
  then
    echo "MyEPM.output.json not found"
    return 1
  fi
  rm $repo/tests/run_fixtures/simple/MyEPM.output.json
  #check_test
  if [ $test_exit -eq 1 ]
  then
    return 1
  fi
  echo
}

test_teardown(){
  if [ -z "$ci" ]
  then
    echo
    if [ "$was_running" -eq 0 ]
    then
      $cli_exec services stop -rx keys
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

echo "Hello! I'm the marmot that tests the [monax pkgs do] command"
echo
echo "testing with target $cli_exec"
echo
start=`pwd`
cd $repo
test_setup

# ---------------------------------------------------------------------------
# Go!

echo "Running Tests..."
perform_tests

# ---------------------------------------------------------------------------
# Cleaning up

test_teardown
