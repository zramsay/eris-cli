#!/usr/bin/env bash

# -----------------------------------------------------------------
# Defaults

package=eris
host_location=/var/www/html
declare -a distros=( precise trusty utopic vivid wheezy jessie stretch )

echo "Prepping keys"
gpg --armor --export 'support@erisindustries.com' > $host_location/APT-GPG-KEY

for distro in "${distros[@]}"
do
  echo -e "\n\nRemoving (old) $package to $distro\n\n"
  reprepro -Vb $host_location remove $distro $package

  echo -e "\n\nAdding $package to $distro\n\n"
  reprepro -Vb $host_location includedeb $distro ./$package*.deb
done

echo -e "\n\nAfter Adding we have the following.\n"
reprepro -b $host_location ls $package

echo "Cleaning up"
rm $HOME/$package*.deb