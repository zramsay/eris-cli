#!/usr/bin/env bash
set -e

# ----------------------------------------------------------------------------
# Set definitions and defaults

# Docker Backend Versions Eris Tests Against
declare -a docker_versions=( "1.8.0" "1.8.1" "1.7.1" )
# declare -a docker_versions=( "1.7.1" )

# Where are the Things
base=github.com/eris-ltd/eris-cli
if [ "$CIRCLE_BRANCH" ]
then
  repo=${GOPATH%%:*}/src/github.com/${CIRCLE_PROJECT_USERNAME}/${CIRCLE_PROJECT_REPONAME}
else
  repo=$GOPATH/src/$base
fi

# For quick dev use comment out the override (second line); circle should use the override
machine_definitions="$(strings /dev/urandom | grep -o '[[:alnum:]]' | head -n 10 | tr -d '\n' ; echo)"
machine_definitions=matDef

# secondary suite of backend machines (swarm, or data_center)
swarm="nyc2"
# primary swarm of backend machines -- comment out to use secondary machines
swarm="ams3"

# how will the tool test run?
entrypoint="/home/eris/test_tool.sh"
testimage=eris/eris
testuser=eris
remotesocket=2376
localsocket=/var/run/docker.sock

# ----------------------------------------------------------------------------
# Define how tests will run

runTests(){
  if [[ $1 == "local" ]]
  then
    machine="eris-test-local"
    swarmb4=$swarm # need to save this value when called with "all"
    swarm=solo
    ver=$(docker version | grep "Client version" | cut -d':' -f2 | sed -e 's/^[[:space:]]*//')
    echo ""
    echo ""
    echo "Testing (locally) against"
    echo -e "Docker version:\t$ver"
    echo -e "In Data Center:\t$swarm"
    echo -e "Machine name:\t$machine"
    # Note NEVER do this in circle. It will explode
    docker run --rm --volumes-from $machine_definitions --entrypoint $entrypoint -e MACHINE_NAME=$machine -v $localsocket:$localsocket --user $testuser $testimage
    swarm=$swarmb4
  else
    machine=eris-test-$swarm-$ver
    echo ""
    echo ""
    echo "Testing against"
    echo -e "Docker version:\t$ver"
    echo -e "In Data Center:\t$swarm"
    echo -e "Machine name:\t$machine"
    echo ""
    echo ""
    echo "Starting Machine."
    docker-machine start $machine
    sleep 15 # boot time for the machine. TODO: curl on a loop..?
    echo "Machine Started."
    if [ $run_count -ne $run_len ]
    then
      set +e
    fi
    if [ "$CIRCLE_BRANCH" ]
    then
      docker run --volumes-from $machine_definitions --entrypoint $entrypoint -e MACHINE_NAME=$machine -p $remotesocket:$remotesocket --user $testuser $testimage:$CIRCLE_BRANCH
    else
      docker run --rm --volumes-from $machine_definitions --entrypoint $entrypoint -e MACHINE_NAME=$machine -p $remotesocket:$remotesocket --user $testuser $testimage
    fi
    if [ $run_count -ne $run_len ]
    then
      set -e
    fi
    echo "Stopping Machine."
    docker-machine kill $machine
    echo "Machine Stopped."
  fi
}

# ---------------------------------------------------------------------------
# Go!
echo ""
echo ""
echo "Getting machine definition files sorted."
echo ""
echo ""
# output suppressed to optimize logs captured. to debug switch commented lines
sh -c "docker run --name $machine_definitions erisindustries/test_machines" &>/dev/null
# docker run --name $machine_definitions erisindustries/test_machines
if [ "$CIRCLE_BRANCH" ]
then
  # Circle won't have the machine definitions on the host. Need to export them
  docker cp $machine_definitions:/home/eris/.docker $HOME
fi

echo ""
echo ""
echo "Building eris in a docker container."
echo ""
echo ""
strt=`pwd`
cd $repo
export testimage
export repo
# suppressed by default as too chatty. to debug, switch commented lines
tests/build_tool.sh > /dev/null
# tests/build_tool.sh

echo "Testing tool."
if [[ $1 == "local" ]]
then
  # We don't wan't purely local tests to exit the script or it will preempt teardown
  set +e
  runTests "local"
  set -e
else
  if [[ $1 == "all" ]]
  then
    # comment the sets if you want "all" to fail if local fails
    # set +e
    runTests "local"
    # set -e
  fi

  # The last API in the array should be the *authoritative* one that will get reported to circle
  run_count=0
  run_len=${#docker_versions[@]}
  for ver in "${docker_versions[@]}"
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
echo ""
echo ""
if [ ! "$CIRCLE_BRANCH" ]
then
  docker rm $machine_definitions
fi

cd $strt
