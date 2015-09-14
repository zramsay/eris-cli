#!/usr/bin/env bash

# -----------------------------------------------------------
# PURPOSE

# This script is made to provision an eris host. It is opinionated
# and should not be used generally.

# -----------------------------------------------------------
# REQUIREMENTS

# Docker installed
# Docker-Machine installed
# DO_TOKEN env var set with DIGITAL_OCEAN_API_KEY

# -----------------------------------------------------------
# USAGE

# host_make.sh DIGITAL_OCEAN_DATA_CENTER DOCKER_VERSION [HOST_TYPE]

# -----------------------------------------------------------
# LICENSE

# GPL3 -- see eris-cli's LICENSE.md

# -----------------------------------------------------------
# Set defaults

start=`pwd`
data_center=$1
docker_version=$2
if [ -z $3 ]
then
  host_type="test"
else
  host_type=$3
fi
machine_base=eris-$host_type-$data_center-$docker_version
scripts_dir=$( dirname "${BASH_SOURCE[0]}" )
ping_times=0

# -----------------------------------------------------------
# Idiot check

if [ -z $data_center ]
then
  echo "Please rerun with a data center as the first arg."
  exit 1
fi

if [ -z $docker_version ]
then
  echo "Please rerun with a docker version as the second arg."
  exit 1
fi

# -----------------------------------------------------------
# Set the functions

create_machine () {
  echo -e "Creating machine =>\t\t$machine_base"

  docker-machine create --driver digitalocean \
    --digitalocean-access-token $DO_TOKEN \
    --digitalocean-region $data_center \
    --engine-env DOCKER_VERSION=$docker_version \
    $machine_base

  if [ $? -ne 0 ]
  then
    return 1
  fi

  echo -e "Machine created."
}

ping_machine() {
  until [[ $(docker-machine status $1) == "Running" ]] || [ $ping_times -eq 10 ]
  do
     ping_times=$[$ping_times +1]
     sleep 3
  done

  if [[ $(docker-machine status $1) != "Running" ]]
  then
    echo "Could not start the machine."
    return 1
  fi

  return 0
}

check_machine() {
  echo -e "Checking Machine's Status =>\t\t$machine_base"
  status=$(docker-machine status $machine_base 2>/dev/null)
  if [[ "$status" == "Error" ]]
  then
    echo -e "Error creating machine =>\t\t$machine_base.\nOld DC =>\t\t\t\t${data_center_array[$1]}"
    return 1
  elif [[ "$status" == "Starting" ]]
  then
    echo -e "Machine is starting. Waiting. =>\t$machine_base"
    ping_machine $machine_base
    return $?
  elif [[ "$status" == "Stopped" ]]
  then
    echo -e "Machine is stopped. Starting. =>\t$machine_base"
    docker-machine start $machine_base 1>/dev/null
    ping_machine $machine_base
    return $?
  elif [[ "$status" == "Running" ]]
  then
    echo -e "Machine is running. Proceding. =>\t$machine_base"
    return 0
  fi
}

# -----------------------------------------------------------
# :dowit:

set -e
create_machine
check_machine

echo -e "Will copy provisioning script into machine.\nPrepare to be ssh'ed in."
docker-machine scp $scripts_dir/host_provision.sh $machine_base:~
docker-machine ssh $machine_base
echo -e "Assuming host is provisioned. Proceeding to package and push the machine files."

# ./host_provision.sh && exit

# -----------------------------------------------------------
# Clean up
cd $start