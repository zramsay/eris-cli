#!/usr/bin/env bash

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

# Docker Backend Versions Eris Tests Against -- Final element in this array is the definitive one.
#   Circle passes or fails based on it. To speed testing uncomment out the second line to override
#   the array and just test against the authoritative one. If testing against a specific backend
#   then change the authoritative one to use that. We define "authoritative" to mean "what docker
#   installs by default on Linux"
declare -a docker_versions17=( "1.7.1" )
declare -a docker_versions18=( "1.8.0" "1.8.1" )
# declare -a docker_versions18=( "1.8.1" )

# Primary swarm of backend machines -- uncomment out second line to use the secondary swarm
#   if/when the primary swarm is either too slow or non-responsive. Swarms here are really
#   data centers. These boxes are on Digital Ocean.
swarm="ams3"
# swarm="nyc2"

# Define now the tool tests within the Docker container will be booted from docker run
entrypoint="/home/eris/test_tool.sh"
testimage=eris/eris
testuser=eris
remotesocket=2376
localsocket=/var/run/docker.sock
machine_definitions=matDef

# ----------------------------------------------------------------------------
# Define how tests will run

runTests(){
  if [[ $1 == "local" ]]
  then
    machine="eris-test-local"
    # need to save this value when called with "all"
    swarmb4=$swarm
    swarm=solo
    ver=$(docker version | grep "Client version" | cut -d':' -f2 | sed -e 's/^[[:space:]]*//')
    # Note NEVER do this in circle. It will explode.
    echo -e "Starting Eris Docker container.\n"
    if [ "$circle" = false ]
    then
      if [[ $(uname -s) == "Linux" ]]
      then
        docker run --rm --volumes-from $machine_definitions --entrypoint $entrypoint -e MACHINE_NAME=$machine -e SWARM=$swarm -e APIVERSION=$ver -v $localsocket:$localsocket --user $testuser $testimage
      else
        docker run --rm --volumes-from $machine_definitions --entrypoint $entrypoint -e MACHINE_NAME=$machine -e SWARM=$swarm -e APIVERSION=$ver -p $remotesocket:$remotesocket --user $testuser $testimage
      fi
    else
      echo "Don't run local in Circle environment."
    fi
    # when doing "tests/test.sh all" we want to use the swarm despite changing above
    test_exit=$(echo $?)
    swarm=$swarmb4
  else
    machine=eris-test-$swarm-$ver
    if [[ "$branch" == "master" ]]
    then
      branch="latest"
    fi
    # only the last element in the backend array should cause this script to exit with
    #   a non-zero exit code
    echo "Starting Eris Docker container."
    if [[ "$1" == "1.7" ]]
    then
      if [ "$circle" = true ]
      then
        docker run --volumes-from $machine_definitions --entrypoint $entrypoint -e MACHINE_NAME=$machine -e SWARM=$swarm -e APIVERSION=$ver -p $remotesocket:$remotesocket --user $testuser $testimage:docker17
      else
        docker run --rm --volumes-from $machine_definitions --entrypoint $entrypoint -e MACHINE_NAME=$machine -e SWARM=$swarm -e APIVERSION=$ver -p $remotesocket:$remotesocket --user $testuser $testimage:docker17
      fi
    else
      if [ "$circle" = true ]
      then
        docker run --volumes-from $machine_definitions --entrypoint $entrypoint -e MACHINE_NAME=$machine -e SWARM=$swarm -e APIVERSION=$ver -p $remotesocket:$remotesocket --user $testuser $testimage:$branch
      else
        docker run --rm --volumes-from $machine_definitions --entrypoint $entrypoint -e MACHINE_NAME=$machine -e SWARM=$swarm -e APIVERSION=$ver -p $remotesocket:$remotesocket --user $testuser $testimage
      fi
    fi
    test_exit=$(echo $?)
  fi
}

# ---------------------------------------------------------------------------
# Get the things build and dependencies turned on

echo "Hello! I'm the testing suite for eris."
echo ""
echo "Getting machine definition files sorted."
# suppressed by default as too chatty
sh -c "docker run --name $machine_definitions erisindustries/test_machines" &>/dev/null

echo ""
echo "Building eris in a docker container."
strt=`pwd`
cd $repo
export testimage
export repo
# suppressed by default as too chatty
tests/build_tool.sh > /dev/null
# tests/build_tool.sh
if [ $? -ne 0 ]
then
  echo "Could not build eris. Debug via by directly running [`pwd`/tests/build_tool.sh]"
  exit 1
fi

# ---------------------------------------------------------------------------
# Go!

echo ""
if [[ $1 == "local" ]]
then
  # We don't wan't purely local tests to exit the script or it will preempt teardown
  runTests 'local'
else
  if [[ $1 == "all" ]]
  then
    runTests "local"
  fi

  # The last API in the array should be the *authoritative* one that will get reported to circle
  run_count=0
  run_len=${#docker_versions18[@]}
  for ver in "${docker_versions17[@]}"
  do
    runTests "1.7"
  done
  for ver in "${docker_versions18[@]}"
  do
    run_count=$[$run_count +1]
    runTests
  done
fi

# ---------------------------------------------------------------------------
# Cleaning up

echo ""
echo ""
echo "Cleaning up"
if [ "$circle" = false ]
then
  sh -c "docker rm $machine_definitions" &>/dev/null
fi

cd $strt
exit $test_exit
