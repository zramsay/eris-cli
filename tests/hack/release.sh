#!/usr/bin/env bash
# assumes golang build from source or golang >= 1.5
# assumes one is on Casey's machine.
# this is still a WIP.

# -----------------------------------------------------------------
# Defaults
start=`pwd`
repo=$GOPATH/src/github.com/eris-ltd/eris-cli
declare -a distros=( precise trusty utopic vivid wheezy jessie stretch )
package=eris
version=0.10.2
aptmachine=eris-build-ams3-apt

# -----------------------------------------------------------------
# Cross Compile
cd $repo
go get github.com/laher/goxc
mkdir builds
goxc -wd cmd/eris -n eris -d builds

# -----------------------------------------------------------------
# Send ... still WIP ... scp to builder....
docker-machine scp builds/$ver/eris_$ver_amd64.deb $aptmachine:~
docker-machine ssh $aptmachine
# for distro in "${distros[@]}"
# do
#   echo -e "\n\nAdding $package to $distro\n\n"
#   reprepro -Vb /var/repositories includedeb $distro ./$package*.deb
# done

# echo -e "\n\nAfter Adding we have the following.\n"
# reprepro -b /var/repositories ls $package
# rm ./$package*.deb

cd $start