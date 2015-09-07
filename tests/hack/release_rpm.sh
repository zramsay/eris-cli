#!/usr/bin/env bash

# -----------------------------------------------------------------
# Prerequisites

# curl https://www.franzoni.eu/keys/D1270819.txt | sudo apt-key add -
# sudo sh -c "echo deb http://www.a9f.eu/apt/docker-rpm-builder-v1/ubuntu vivid main > /etc/apt/sources.list.d/docker_rpm_builder.list"
# sudo apt-get -qq update
# sudo apt-get -qq download docker-rpm-builder
# sudo dpkg --force-all -i *.deb
# sudo rm *.deb
# sudo rm /etc/apt/sources.list.d/docker_rpm_builder.list
# sudo apt-get -f -y install

# -----------------------------------------------------------------
# Defaults

# package=eris
# declare -a distros=( precise trusty utopic vivid wheezy jessie stretch )

# for distro in "${distros[@]}"
# do
#   echo -e "\n\nRemoving (old) $package to $distro\n\n"
#   reprepro -Vb /var/repositories remove $distro $package

#   echo -e "\n\nAdding $package to $distro\n\n"
#   reprepro -Vb /var/repositories includedeb $distro ./$package*.deb
# done

# echo -e "\n\nAfter Adding we have the following.\n"
# reprepro -b /var/repositories ls $package
