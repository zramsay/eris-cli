#!/usr/bin/env sh
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
# ./host_provision.sh && docker version && exit

export DOCKER_VERSION=$(hostname | cut -d'-' -f4)
wget -qO- https://get.docker.io/gpg | apt-key add -
echo deb http://get.docker.io/ubuntu docker main > /etc/apt/sources.list.d/docker.list
apt-get update -qq
apt-get install -qqy docker-engine
stop docker
curl -sSL --ssl-req -o $(which docker) https://get.docker.com/builds/Linux/x86_64/docker-$DOCKER_VERSION
start docker