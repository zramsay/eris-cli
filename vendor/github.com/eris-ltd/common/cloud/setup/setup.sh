#!/usr/bin/env bash
# -----------------------------------------------------------------------------
# PURPOSE

# This script will setup eris and all of its dependencies. It is primarily meant
# for running on cloud providers as a setup script.

# Specifically the script will install:
#   * nodejs (useful for middleware)
#   * go+git (useful for quick updates of the eris tool)
#   * eris

# The script assumes that it will be ran by a root user or a user with sudo
# privileges on the node. Note that it does not currently check that it has
# elevate privileges on the node.

# Note that the script, by default, will **not** install Docker which is a
# **required** dependency for Eris. If, however, the environment variable
# $INSTALL_DOCKER is not blank, then the script will install docker via the
# easy docker installation. If this makes you paranoid then you should
# manually install docker **before** running this script.

# Note that the script also assumes that the user will be a bash user.

# -----------------------------------------------------------------------------
# LICENSE

# The MIT License (MIT)
# Copyright (c) 2016-Present Eris Industries, Ltd.

# Permission is hereby granted, free of charge, to any person obtaining a copy
# of this software and associated documentation files (the "Software"), to deal
# in the Software without restriction, including without limitation the rights
# to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
# copies of the Software, and to permit persons to whom the Software is
# furnished to do so, subject to the following conditions:

# The above copyright notice and this permission notice shall be included in
# all copies or substantial portions of the Software.

# THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
# IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
# FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
# AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
# LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
# FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS
# IN THE SOFTWARE.

# -----------------------------------------------------------------------------
# REQUIREMENTS

# Ubuntu
# Docker (**unless** INSTALL_DOCKER is not blank)

# -----------------------------------------------------------------------------
# USAGE

# setup.sh [USER] [SERVICESTOSTART] [CHAINSTOSTART]

# -----------------------------------------------------------------------------
# Set defaults

erisUser=$1
if [[ "$erisUser" == "" ]]
then
  erisUser=$USER
fi
if [[ "$erisUser" == "root" ]]
then
  userHome=/root
else
  userHome=/home/$erisUser
fi
services=( $(echo $2 | tr "," "\n") )
chains=( $(echo $3 | tr "," "\n") )
toStart=( "${services[@]}" "${chains[@]}" )

# -----------------------------------------------------------------------------
# Defaults

GOVERSION="1.6.3"
NODEVERSION="4"

# -----------------------------------------------------------------------------
# Install dependencies

echo "Hello there! I'm the marmot that installs Eris"
echo
echo
echo "Grabbing necessary dependencies"
export DEBIAN_FRONTEND=noninteractive
curl -sSL https://deb.nodesource.com/setup_"$NODEVERSION".x | sudo -E bash - &>/dev/null
sudo apt-get install -y jq gcc git build-essential nodejs &>/dev/null
rm -fr /usr/local/go
curl -sSL https://storage.googleapis.com/golang/go"$GOVERSION".linux-amd64.tar.gz | sudo tar -C /usr/local -xzf - &>/dev/null 
if [ -n "$INSTALL_DOCKER" ]
then
  curl -sSL https://get.docker.com/ | sudo -E bash - &>/dev/null
fi
sudo usermod -a -G docker $erisUser &>/dev/null
echo "Dependencies installed"
echo
echo

# -----------------------------------------------------------------------------
# Getting chains

echo "Getting chain managers"
curl -sSL -o $userHome/simplechain.sh https://raw.githubusercontent.com/eris-ltd/common/master/cloud/chains/simplechain.sh
chmod +x $userHome/*.sh
chown $erisUser:$erisUser $userHome/*.sh
echo "Chain managers acquired"
echo
echo

# -----------------------------------------------------------------------------
# Install Eris

sudo -u "$erisUser" -i env START="`printf ",%s" "${toStart[@]}"`" bash <<'EOF'
start=( $(echo $START | tr "," "\n") )
echo "Setting up Go for the user"
mkdir --parents $HOME/go
export GOPATH=$HOME/go
export PATH=/usr/local/go/bin:$HOME/go/bin:$PATH
echo "export GOROOT=/usr/local/go" >> $HOME/.bashrc
echo "export GOPATH=$HOME/go" >> $HOME/.bashrc
echo "export PATH=$HOME/go/bin:/usr/local/go/bin:$PATH" >> $HOME/.bashrc
echo "Finished setting up Go"
echo
echo
echo "Version information"
echo
go version
if [ $? -ne 0 ]
then
  echo
  echo Go is not installed, aborting
  exit 1
fi
echo
docker version
if [ $? -ne 0 ]
then
  echo
  echo Docker daemon is not running, aborting
  exit 1
fi
echo
echo
echo "Building Eris"
go get github.com/eris-ltd/eris-cli/cmd/eris
if [ $? -ne 0 ]
then
  echo 
  echo Failed building Eris repository, aborting    
  exit 1
fi
echo
echo
echo "Initializing Eris"
export ERIS_PULL_APPROVE="true"
export ERIS_MIGRATE_APPROVE="true"
echo "export ERIS_PULL_APPROVE=\"true\"" >> $HOME/.bashrc
echo "export ERIS_MIGRATE_APPROVE=\"true\"" >> $HOME/.bashrc
eris init --yes 2>/dev/null
if [ $? -ne 0 ]
then
  echo
  echo Failed pulling Eris images, aborting
  exit 1
fi
echo
echo
echo "Starting services and chains  ${start[@]}"
echo
if [ ${#start[@]} -eq 0 ]
then
  echo "No services or chains selected"
else
  for x in "${start[@]}"
  do
    if [ -f "$HOME/$x".sh ]
    then
      echo "Turning on chain: $x"
      $HOME/$x.sh
    else
      echo "Turning on service: $x"
      eris services start $x
    fi
  done
fi
EOF
if [ $? -ne 0 ]
then
  exit 1
fi

echo
echo "Finished starting services and chains"

echo
echo
echo "Eris installed!"
