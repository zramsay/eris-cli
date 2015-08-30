#!/usr/bin/env sh
set -e

# for reference only. it takes forever to provision these
# docker-machine create --driver digitalocean \
#   --digitalocean-access-token $DO_TOKEN \
#   --digitalocean-region ams3 \
#   --engine-env DOCKER_VERSION=1.8.1 \
#   eris-build-ams3-apt

export DOCKER_VERSION=$(hostname | cut -d'-' -f4)
wget -qO- https://get.docker.io/gpg | apt-key add -
echo deb http://get.docker.io/ubuntu docker main > /etc/apt/sources.list.d/docker.list
apt-get update -qq
apt-get install -qqy lxc-docker
stop docker
curl -sSL --ssl-req -o $(which docker) https://get.docker.com/builds/Linux/x86_64/docker-$DOCKER_VERSION
start docker