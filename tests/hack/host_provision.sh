#!/usr/bin/env bash
#
# how to use......
# for reference only. it takes forever to provision these
#
# docker-machine create --driver digitalocean \
#   --digitalocean-access-token $DO_TOKEN \
#   --digitalocean-region ams3 \
#   --engine-env DOCKER_VERSION=1.8.1 \
#   eris-test-ams3-1.8.1
#
# docker-machine scp host_provision.sh eris-test-ams3-1.8.1:~
# docker-machine ssh eris-test-ams3-1.8.1
# ./host_provision.sh && exit

default_docker="1.8.1"

read -p "This script only works on Ubuntu (and does no checking). It may work on some debians. Do you wish to proceed? (y/n) " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]
then
  echo "OK. Not doing anything. Bye."
  exit 1
fi
echo "You confirmed you are on Ubuntu (or waived compatibility)."

if [[ "$USER" != "root" ]]
then
  read -p "You are not the root user. Did you run this as (sudo). (y/n) " -n 1 -r
  echo
  if [[ ! $REPLY =~ ^[Yy]$ ]]
  then
    echo "OK. Not doing anything. Bye."
    exit 1
  fi
fi
echo "Privileges confirmed."

if [ -z "$DOCKER_VERSION" ]
then
  echo "You do not have the \$DOCKER_VERSION set. Trying via hostname (an Eris paradigm)."
  export DOCKER_VERSION=$(hostname | cut -d'-' -f4)
  if [[ "$DOCKER_VERSION" == `hostname` ]]
  then
    read -p "I cannot find the Docker Version to Install. You can rerun me with \$DOCKER_VERSION set or use the defaults. Would you like the defaults? (y/n) " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]
    then
      export DOCKER_VERSION="$default_docker"
    fi
  fi
fi

echo "Will install Docker for Version: $DOCKER_VERSION"
echo ""
echo ""

# very dumb pre 1.7 install
if [[ "$DOCKER_VERSION" == "1.6.2" ]]
then
  echo "You're docker version is seriously old. Please consider upgrading."
  echo ""
  sudo apt-key adv --keyserver hkp://keyserver.ubuntu.com:80 --recv-keys 36A1D7869245C8950F966E92D8576A8BA88D21E9
  sudo sh -c "echo deb https://get.docker.io/ubuntu docker main > /etc/apt/sources.list.d/docker.list"
  sudo apt-get update -qq
  sudo apt-get install -qqy lxc-docker-1.6.2
  echo ""
  echo "Docker installed"
else
  echo ""
  wget -qO- https://get.docker.io/gpg | apt-key add -
  echo deb http://get.docker.io/ubuntu docker main > /etc/apt/sources.list.d/docker.list
  apt-get update -qq
  apt-get install -qqy docker-engine
  stop docker
  curl -sSL --ssl-req -o $(which docker) https://get.docker.com/builds/Linux/x86_64/docker-$DOCKER_VERSION
  echo ""
  echo "Docker installed"
fi

echo "Restarting Newly Installed Docker"
echo ""
start docker
echo ""
sleep 3 # boot time
echo ""
docker version
echo ""
echo ""
if [[ "$USER" != "root" ]]
then
  read -p "Would you like to add this user to the docker group? (y/n) " -n 1 -r
  echo
  if [[ $REPLY =~ ^[Yy]$ ]]
  then
    echo "OK. Adding to group"
    usermod -a -G docker $USER
  fi
fi
