#!/usr/bin/env bash

# -----------------------------------------------------------------
# Defaults

package=eris
declare -a distros=( precise trusty utopic vivid wheezy jessie stretch )

for distro in "${distros[@]}"
do
  echo -e "\n\nRemoving (old) $package to $distro\n\n"
  reprepro -Vb /var/repositories remove $distro $package

  echo -e "\n\nAdding $package to $distro\n\n"
  reprepro -Vb /var/repositories includedeb $distro ./$package*.deb
done

echo -e "\n\nAfter Adding we have the following.\n"
reprepro -b /var/repositories ls $package
