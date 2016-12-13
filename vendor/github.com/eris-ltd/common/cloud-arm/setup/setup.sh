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

GOVERSION="1.6"
NODEVERSION="4"
DOCKER_HYPRIOT_VERSION="1.10.3-1"

# -----------------------------------------------------------------------------
# Install dependencies

echo "Hello $erisUser! I'm the marmot that installs Eris."
echo
echo
echo "Grabbing necessary dependencies"
export DEBIAN_FRONTEND=noninteractive
curl -sSL https://deb.nodesource.com/setup_"$NODEVERSION".x | sudo -E bash - &>/dev/null
sudo apt-get install -y bc jq gcc git build-essential nodejs &>/dev/null

# -$- Install arm go -$-
if [ ! -d "/usr/local/go$GOVERSION" ]; then
  curl -sSL https://www.dropbox.com/s/1v8uxdn6oo48t2g/go1.6.tar.gz?dl=0 | sudo tar -C /usr/local -xzf - >/dev/null
  echo "Installed go$GOVERSION to /usr/local/"
fi

# -$- Install hypriot docker [http://blog.hypriot.com/downloads/] -$-
if [ -n "$INSTALL_DOCKER" ]
then
  wget https://downloads.hypriot.com/docker-hypriot_"$DOCKER_HYPRIOT_VERSION"_armhf.deb &&
  dpkg -i docker-hypriot_"$DOCKER_HYPRIOT_VERSION"_armhf.deb &&
  rm docker-hypriot_"$DOCKER_HYPRIOT_VERSION"_armhf.deb
  echo "Installed docker-hypriot_${DOCKER_HYPRIOT_VERSION}_armhf"
fi

sudo usermod -a -G docker $erisUser &>/dev/null
echo "Dependencies Installed."
echo
echo

# -----------------------------------------------------------------------------
# Getting chains

echo "Getting Chain managers"
curl -sSL -o $userHome/simplechain.sh https://raw.githubusercontent.com/eris-ltd/common/master/cloud/chains/simplechain.sh
chmod +x $userHome/*.sh
chown $erisUser:$erisUser $userHome/*.sh
echo "Chain managers acquired."
echo
echo

# -----------------------------------------------------------------------------
# Install eris

sudo -u "$erisUser" -i env START=$(printf ",%s" "${toStart[@]}") bash <<'EOF'
GOVERSION="1.6"
start=( $(echo $START | tr "," "\n") )
echo "Setting up Go for the user"
mkdir --parents $HOME/go
if [ -z "$GOPATH" ]
then
    export GOPATH=$HOME/go
    export GOROOT=/usr/local/go$GOVERSION
    export PATH=$HOME/go/bin:/usr/local/go"$GOVERSION"/bin:$PATH
    echo "export GOROOT=/usr/local/go$GOVERSION" >> $HOME/.bashrc
    echo "export GOPATH=$HOME/go" >> $HOME/.bashrc
    echo "export PATH=$HOME/go/bin:/usr/local/go$GOVERSION/bin:\$PATH" >> $HOME/.bashrc
    echo "Finished Setting up Go."
fi
echo
echo
echo "Version Information"
echo
go version
echo
docker version
echo
echo
echo "Building eris."
pre_dir=$pwd
go get -d github.com/eris-ltd/eris-cli/cmd/eris
cd $GOPATH/src/github.com/eris-ltd/eris-cli/cmd/eris
git checkout armhf
go build
go install
cd $pre_dir

echo "Eris-cli installed!"
echo
if [ -z "$ERIS_PULL_APPROVE" ]
then
    echo "Initializing eris."
    export ERIS_PULL_APPROVE="true"
    export ERIS_MIGRATE_APPROVE="true"
    echo "export ERIS_PULL_APPROVE=\"true\"" >> $HOME/.bashrc
    echo "export ERIS_MIGRATE_APPROVE=\"true\"" >> $HOME/.bashrc
fi
eris init --yes 2>/dev/null
echo
echo
echo "Starting Services and Chains: ${start[@]}"
echo
if [ ${#start[@]} -eq 0 ]
then
  echo "No services or chains selected"
else
  for x in "${start[@]}"
  do
    if [ -f "$HOME/$x".sh ]
    then
      echo "Turning on Chain: $x"
      $HOME/$x.sh
    else
      echo "Turning on Service: $x"
      eris services start $x
    fi
  done
fi
EOF

echo
echo "Finished starting services and chains."

# -------------------------------------------------------------------------------
# Cleanup

rm $userHome/*.sh
echo
echo
echo "Eris Installed!"
