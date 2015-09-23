#!/usr/bin/env bash

# -----------------------------------------------------------
# PURPOSE

# This script will check a swarm of boxes to make sure they are
# all running and do not have docker problems. Sometimes they
# need to be cycled for testing purposes.

# -----------------------------------------------------------
# REQUIREMENTS

# Docker installed
# Docker-Machine installed
# Eris testing machine definition files

# -----------------------------------------------------------
# USAGE

# swarm_restart.sh

# -----------------------------------------------------------
# LICENSE

# GPL3 -- see eris-cli's LICENSE.md

# -----------------------------------------------------------
# Set defaults

start=`pwd`
declare -a swarms
declare -a vers
declare -a results
declare -a machines

swarms=( "ams3" "nyc2" )
# not cycling 1.6.2 and 1.7.1 cause different api versions.
# vers=( "1.8.0" "1.8.1" "1.8.2" )
vers=( "1.6.2" "1.7.1" "1.8.0" "1.8.1" "1.8.2" )

echo "Hello. I'm the friendly eris testing swarm cycling marmot. I take a while."

# ----------------------------------------------------------
# Async process functions

set_procs() {
  checks[$1]=$!
  machines[$1]=$machine
}

wait_procs() {
  for chk in "${!checks[@]}"
  do
    wait ${checks[$chk]}
    results[$chk]=$?
  done
}

check_procs() {
  for res in "${!results[@]}"
  do
    if [ ${results[$res]} -ne 0 ]
    then
      machine_results+=("${machines[$res]} is Red. :(")
    else
      machine_results+=("${machines[$res]} is Green!")
    fi
  done
}

# -----------------------------------------------------------
# Set functions

set_machine() {
  machine=eris-test-$swarm-$ver
}

check_machine() {
  ping_times=0
  until [[ $(docker-machine status $machine) = "Running" ]] || [ $ping_times -eq 10 ]
  do
    ping_times=$[$ping_times +1]
    sleep 3
  done

  if [[ $(docker-machine status $machine) != "Running" ]]
  then
    echo "Could not start the machine."
    exit 1
  fi
  echo "Machine Started."
  echo -e "Connecting to machine =>\t$machine"
  set -e
  eval "$(docker-machine env $machine)" 1>/dev/null
  set +e
  echo "Connected to Machine."
}

display_info() {
  sleep 5
  docker version
  if [ $? -ne 0 ]; then return 1; fi
  echo ""
  echo ""
  docker ps -a
  if [ $? -ne 0 ]; then return 1; fi
  echo ""
  echo ""
  docker images
  if [ $? -ne 0 ]; then return 1; fi
}

run_sequence() {
  echo -e "Starting Machine Sequence =>\t$machine"
  docker-machine kill $machine &>/dev/null
  if [ $? -ne 0 ]; then return 1; fi
  docker-machine start $machine 1>/dev/null
  if [ $? -ne 0 ]; then return 1; fi
  if [[ "$ver" != 1.8* ]]
  then
    echo -e "Skipping display and check machine for this machine (incompatible API versions)."
  else
    check_machine
    if [ $? -ne 0 ]; then return 1; fi
    echo -e "Info for machine =>\t\t$machine"
    display_info
    if [ $? -ne 0 ]; then return 1; fi
  fi
  echo -e "Stopping machine =>\t\t$machine"
  docker-machine kill $machine 1>/dev/null
  if [ $? -ne 0 ]; then return 1; fi
  echo -e "Machine Sequence Complete =>\t$machine"
}

# ----------------------------------------------------------
# :dowit:

i=0
for swarm in "${swarms[@]}"
do
  for ver in "${vers[@]}"
  do
    set_machine
    run_sequence &
    set_procs $i
    let i=i+1
  done
done

wait_procs
check_procs

echo ""
echo ""
docker-machine ls

# ----------------------------------------------------------
# Cleanup
echo ""
echo ""
echo "Your summary good human...."
printf '%s\n' "${machine_results[@]}"
cd $start