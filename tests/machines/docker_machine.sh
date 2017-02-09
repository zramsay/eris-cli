#!/usr/bin/env bash
# ---------------------------------------------------------------------------
# PURPOSE

# Simply a library of docker machine/docker image setup/utility commands 
# segregated for better modularity of the testing suite.

# ----------------------------------------------------------------------------
# Utility functions

# this function is to provide simple output to stdOut during the sometimes
# long-ish machine building process. If the timing of this function is
# changed it should be harmonized with the timing of the `timeOutTicker`
# of setup.go
function sleeper() {
  sleep 15
  sp="...."
  sc=0
  ticks=0
  # 15 minutes + standard CI 10 minutes will suffice
  until [ $ticks -eq 15 ]
  do
    printf "\b${sp:sc++:1}"
    ((sc==${#sp})) && sc=0
    sleep 60
    ticks=$((ticks + 1))
  done
}

# ----------------------------------------------------------------------------
# Build eris in a docker images

# build docker images; runs as background process
function build_eris() {
  echo "Building eris in a docker container."
  cd $repo
  export BRANCH=$1
  tests/build_tool.sh &>/dev/null
  if [ $? -ne 0 ]
  then
    exit 1
  fi
  exit 0
}

# ensure eris image builds; machines built; and machines can be connected to
function check_build() {
  echo
  if [ "$build_result" -ne 0 ] && [ -z $1 ]
  then
    echo "Could not build eris image. Rebuilding eris image."
    sleep 5
    build_eris $BRANCH &
    wait $!
    build_result=$?
    check_build "rebuild"
  elif [ "$build_result" -ne 0 ] && [ ! -z $1 ]
  then
    echo "Failure building eris image. Debug via by directly running [`pwd`/tests/build_tool.sh]. Exiting tests."
    remove_machine
    exit 1
  fi

  if [ -z $1 ]
  then 
    echo "Building machine."
    create_machine
  elif [ "$machi_result" -ne 0 ]
  then
    echo "Failure making machines. Exiting tests."
    remove_machine
    exit 1
  fi

  if [ $? -ne 0 ]
  then
    echo "Could not connect to machine(s). Rebuilding machines."
    clear_machine
    remove_machine
    sleep 5
    create_machine
    machi_result=$?
    check_build "rebuild"
  elif [ $? -ne 0 ] && [ ! -z $1 ]
  then
    echo "Failure connecting to machines. Exiting tests."
    remove_machine
    exit 1
  fi

  echo "Setup and checks complete."
}

# ----------------------------------------------------------------------------
# Machine management functions

# create the machine
function create_machine() {
  echo
  echo "creating machine $MACHINE_NAME"
  sleeper &
  ticker=$!
  if [ "$ci" = true ]
  then
    AWS_SSH_USER=""
    if [ "$win" = true ] || [ "$osx" = true ]
    then 
      docker-machine create                                 \
        --driver amazonec2                                  \
        --amazonec2-access-key      "$AWS_ACCESS_KEY_ID"    \
        --amazonec2-secret-key      "$AWS_SECRET_ACCESS_KEY"\
        --amazonec2-region          "eu-west-1"             \
        --amazonec2-vpc-id          "$AWS_VPC_ID"           \
        --amazonec2-security-group  "$AWS_SECURITY_GROUP"   \
        --amazonec2-zone            "b"                     \
        "$MACHINE_NAME"
    else
      docker-machine create                                 \
        --driver amazonec2                                  \
        --amazonec2-access-key      "$AWS_ACCESS_KEY_ID"    \
        --amazonec2-secret-key      "$AWS_SECRET_ACCESS_KEY"\
        --amazonec2-region          "eu-west-1"             \
        --amazonec2-vpc-id          "$AWS_VPC_ID"           \
        --amazonec2-security-group  "$AWS_SECURITY_GROUP"   \
        --amazonec2-zone            "a"                     \
        "$MACHINE_NAME"
    fi
    setup_result=$?
  else
    if [ "$win" = true ]
    then
      docker-machine create --driver hyperv "$MACHINE_NAME"
    elif [ "$osx" = true ]
    then
      docker-machine create --virtualbox virtualbox "$MACHINE_NAME"
    else
      #assume linux...not sure what to put here for the driver
      return 1
    fi
  fi
  kill $ticker
  wait $ticker 2>/dev/null
  if [ "$setup_result" -ne 0 ]
  then
    return 1
  fi
  echo
  echo "Machine created, attempting to start..."
  start_machine
  start_result=$?
  if [ "$start_result" -ne 0 ]
  then
    return 1
  fi
  return 0
}

# start the machine
function start_machine() {
  echo
  if [ -z "$MACHINE_NAME" ]
  then
    echo "Could not find machine to start"
    return 1
  fi
  echo "Attempting to start" "$MACHINE_NAME"
  sleeper &
  ticker=$!
  cd $repo/tests/machines
  docker-machine start "$MACHINE_NAME"
  #pass in local env vars
  eval "$(docker-machine env $MACHINE_NAME)"
  
  if [ "$osx" != true ] && [ "$win" != true ]
  then
    #copy in script to machine
    echo "Got inside start machine"
    docker-machine scp "$script" "$MACHINE_NAME":
    if [ $? -ne 0 ]
    then
      return 1
    fi
    #run command
    docker-machine ssh "$MACHINE_NAME" sudo ./"$script"
    if [ $? -ne 0 ]
    then
      return 1
    fi
  fi

  kill $ticker
  wait $ticker 2>/dev/null
  echo "Machine successfully started"
  return 0
}

# remove env vars
function clear_machine() {
  unset DOCKER_TLS_VERIFY
  unset DOCKER_HOST
  unset DOCKER_CERT_PATH
  unset DOCKER_MACHINE_NAME
}

# Adds the results for a particular box to the MACH_RESULTS array
#   which is displayed at the end of the tests.
function log_machine() {
  if [ "$1" -eq 0 ]
  then
    MACH_RESULTS+=( "$machine is Green!" )
  else
    MACH_RESULTS+=( "$machine is Red.  :(" )
  fi
}

# remove the machines
function remove_machine() {
  if [ "$ci" = true ]
  then
    docker-machine rm --force $(docker-machine ls -q)
  else
    docker-machine rm --force $machine
  fi
}
