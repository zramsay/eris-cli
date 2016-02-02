#!/usr/bin/env bash

# -----------------------------------------------------------------
# Prerequisites

# gpg2 --import secret.asc
# shred -u secret.asc
# apt-get update && apt-get upgrade && apt-get install gcc rpm createrepo
# echo -e "%_signature gpg
# %_gpg_path /root/.gnupg
# %_gpg_name <support@erisindustries.com>
# %_gpgbin %{_bindir}/gpg2" > ~/.rpmmacros
# mkdir /var/www/html/
# gpg2 --armor --export support@erisindustries.com > tmp
# rpm --import tmp && rm tmp

# -----------------------------------------------------------------
# Defaults

start=`pwd`
arch_type="x86_64"
this_user="eris-ltd"
this_repo="eris-cli"
host_location=/var/www/html

# -----------------------------------------------------------------
# Get it

echo "Checking version..."
version=$(cat version)

# -----------------------------------------------------------------
# Build it

rm -rf $HOME/rpmbuild/*
$HOME/eris init --yes --pull-images=false
$HOME/eris man --dump > $HOME/eris.1
gpg2 --armor --export 3C7AFAEB > $host_location/RPM-GPG-KEY
rpm --import $host_location/RPM-GPG-KEY
export ERIS_VERSION=$version
export ERIS_RELEASE=$arch_type
echo -e "Releasing =>\t\t\t$ERIS_VERSION:$ERIS_RELEASE"
rpmbuild -ba --sign $HOME/"$this_repo".spec

# -----------------------------------------------------------------
# Move into position

echo "Moving files into position"
cp -v $HOME/rpmbuild/RPMS/$ERIS_RELEASE/"$this_repo"-"$ERIS_VERSION"-"$ERIS_RELEASE"."$ERIS_RELEASE".rpm $host_location/$ERIS_RELEASE/
cp -v $HOME/rpmbuild/SRPMS/"$this_repo"-"$ERIS_VERSION"-"$ERIS_RELEASE".src.rpm  $host_location/source/
cp -v $HOME/eris.repo $host_location/eris.repo

echo "Preparing repo files"
createrepo $host_location/$ERIS_RELEASE
createrepo $host_location/source

# -----------------------------------------------------------------
# Cleanup

echo "All done!"
cd $start
