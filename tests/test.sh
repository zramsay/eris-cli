#!/usr/bin/env bash

# ---------------------------------------------------------------------------
# PURPOSE

# This script will test the eris. If it is given the "local" argument, then
# it will run against the local docker backend. If it is not given the local
# argument then it will run against a series of docker backends defined in
# arrays at the top of the script.
#
# What this script will do is first it will define what should be run, then
# it will make sure that it has access to the eris' test machine definition
# image files necessary to connect into the backends. Then it will run the
# test_tool and test_stack scripts against a set of backends. That set will
# either be a random element of a given array for a major docker version, or
# it will run against the entire suite of backends.

# ---------------------------------------------------------------------------
# REQUIREMENTS

# Docker installed locally

# ---------------------------------------------------------------------------
# USAGE

# test.sh [local]

# ----------------------------------------------------------------------------
# Set definitions and defaults

# Where are the Things
base=github.com/eris-ltd/eris-cli
if [ "$CIRCLE_BRANCH" ]
then
  repo=${GOPATH%%:*}/src/github.com/${CIRCLE_PROJECT_USERNAME}/${CIRCLE_PROJECT_REPONAME}
  circle=true
else
  repo=$GOPATH/src/$base
  circle=false
fi
branch=${CIRCLE_BRANCH:=master}
branch=${branch/-/_}
if [[ "$branch" == "" ]]
then
  # get from git
  branch=`git rev-parse --abbrev-ref HEAD`
fi

# # Docker Backend Versions Eris Tests Against -- Final element in the final array is the
# #   definitive one. Circle passes or fails based on it. We define "authoritative" to mean
# #   "what docker installs by default on Linux"
# declare -a docker_versions18=( "1.8.0" "1.8.1" "1.8.2" "1.8.3" )
# declare -a docker_versions19=( "1.9.0" "1.9.1" )
# declare -a machine_results=()

# # Primary and secondary swarm of backend machines. Swarms here are really data centers.
# #   These boxes are on AWS.
# swarm_prim="dca1"
# swarm_back="fra1"
# swarm=$swarm_prim

# Define now the tool tests within the Docker container will be booted from docker run
entrypoint="/home/eris/test_tool.sh"
testimage=quay.io/eris/eris
testuser=eris
remotesocket=2376
localsocket=/var/run/docker.sock
dm_path=".docker/machine"
# machine_definitions=matDef
strt=`pwd`

# ----------------------------------------------------------------------------
# Check swarm and machine stuff

# # Sets the name of the machine using eris box conventions
# set_machine() {
#   echo "eris-test-$swarm-$ver"
# }

# Checks whether the primary swarm for the current version of docker is running
#   which indicates it is being used by a different test run. If the primary
#   swarm for the current machine is running, then it will switch to using the
# #   secondary swarm for the same version of the version. If that box is also
# #   being used by another test run then that version will not be used.
# check_swarm() {
#   machine=$(set_machine)

#   if [[ $(docker-machine status $machine) == "Running" ]]
#   then
#     echo "Machine Running. Switching Swarm."
#     if [[ "$swarm" == "$swarm_back" ]]
#     then
#       swarm=$swarm_prim
#     else
#       swarm=$swarm_back
#     fi

#     machine=$(set_machine)
#     if [[ $(docker-machine status $machine) == "Running" ]]
#     then
#       echo "Backup Swarm Machine Also Running."
#       return 1
#     fi
#   else
#     echo "Machine not Running. Keeping Swarm."
#     machine=$(set_machine)
#   fi
# }

# # Changes back to the primary swarm
# reset_swarm() {
#   swarm=$swarm_prim
# }

# Adds the results for a particular box to the machine_results array
#   which is displayed at the end of the tests.
log_results() {
  if [ "$test_exit" -eq 0 ]
  then
    machine_results+=("$machine is Green!")
  else
    machine_results+=("$machine is Red.  :(")
  fi
}

# ----------------------------------------------------------------------------
# Define how tests will run

runTests(){
  # if [[ $1 == "local" ]]
  # then
  #   # machine="eris-test-local"
  #   # swarm=solo

  #   # # Note NEVER do this in circle. It will explode.
  #   # echo -e "Starting Eris Docker container.\n"
  #   # if [ "$circle" = false ]
  #   # then
  #   #   if [[ $(uname -s) == "Linux" ]]
  #   #   then
  #   #     docker run --rm --entrypoint $entrypoint -e MACHINE_NAME=$machine -v $localsocket:$localsocket --user $testuser $testimage:$1
  #   #   else
  #   #     docker run --rm --entrypoint $entrypoint -e MACHINE_NAME=$machine -p $remotesocket --user $testuser $testimage:$1
  #   #   fi
  #   # else
  #   #   echo "Don't run local in Circle environment."
  #   # fi

  #   # # logging the exit code
  #   # test_exit=$(echo $?)
  #   # log_results

  #   # # reset the swarm
  #   # reset_swarm
  # else
    # check_swarm
    if [ $? -ne 0 ]; then return 1; fi

    echo "Starting Eris Docker container."
    if [ "$circle" = true ]
    then
      docker run --volume $HOME/$dm_path:/home/$testuser/$dm_path --entrypoint $entrypoint -e MACHINE_NAME=$machine -p $remotesocket --user $testuser $testimage:$1
    else
      docker run --rm --volume $HOME/$dm_path:/home/$testuser/$dm_path --entrypoint $entrypoint -e MACHINE_NAME=$machine -p $remotesocket --user $testuser $testimage:$1
    fi

    # logging the exit code
    test_exit=$(echo $?)
    log_results

    # reset the swarm
    # reset_swarm
  # fi
}

# ---------------------------------------------------------------------------
# Get the things build and dependencies turned on

echo "Hello! I'm the testing suite for eris."
echo ""
echo "Building eris in a docker container."
cd $repo
export testimage
export repo
tests/build_tool.sh 1>/dev/null
if [ $? -ne 0 ]
then
  echo "Could not build eris. Debug via by directly running [`pwd`/tests/build_tool.sh]"
  exit 1
fi

echo ""
echo "Getting machines sorted."
cd $repo/tests/machines
machines=( $(go run setup.go) )
if [ $? -ne 0 ]
then
  docker-machine rm $(docker-machine ls --filter -q)
  echo "Failure making machines. Exiting Tests."
  printf '%s\n' "${machines[@]}"
  exit 1
fi

# ---------------------------------------------------------------------------
# Go!

echo ""
for machine in "${machines[@]}"
do
  if [[ "$machine" == *1.8* ]]
  then
    runTests "docker18"
  else
    runTests $branch
  fi
done
# if [[ $1 == "local" ]]
# then
#   runTests "local"
# else
#   # we only run against all backends on a few branches
#   if [[ "$branch" == "master" ]]
#   then
#     for ver in "${docker_versions18[@]}"
#     do
#     done

#     for ver in "${docker_versions19[@]}"
#     do
# 	  done
#   else
# 	  # run the tests for only one of the docker versions at random
#     ver=${docker_versions18[RANDOM%${#docker_versions18[@]}]}
#     runTests "docker18"

# 	  ver=${docker_versions19[RANDOM%${#docker_versions19[@]}]}
#     runTests $branch
#   fi
# fi

# ---------------------------------------------------------------------------
# Cleaning up

for machine in "${machines[@]}"
do
  docker-machine rm --force $machine
done
echo ""
echo ""
echo "Your summary good human...."
printf '%s\n' "${machine_results[@]}"
echo ""
echo ""
echo "Done. Exiting with code: $test_exit"
cd $strt
exit $test_exit
