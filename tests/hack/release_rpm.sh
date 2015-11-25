#!/usr/bin/env bash

# -----------------------------------------------------------------
# Prerequisites

# gpg --import secret.asc
# shred -u secret.asc
# apt-get update && apt-get upgrade && apt-get install gcc rpm createrepo
# echo -e "%_signature gpg
# %_gpg_name <support@erisindustries.com>" > ~/.rpmmacros
# wget https://storage.googleapis.com/golang/go1.5.1.linux-amd64.tar.gz | tar -C /usr/local -xzf -
# export PATH=$PATH:/usr/local/go/bin
# mkdir /var/www/html/eris
# gpg --armor --export 'support@erisindustries.com' > /var/www/html/eris/RPM-GPG-KEY
# #### START SERVER
# dm scp tests/hack/eris-cli.spec eris-build-ams3-yum:
# dm scp tests/hack/release_rpm.sh eris-build-ams3-yum:

# -----------------------------------------------------------------
# Defaults

start=`pwd`
arch_type="x86_64"
this_user="eris-ltd"
this_repo="eris-cli"
host_location=/var/www/html
erisbuilddir=/var/tmp/eris-rpmbuild.tmp
repodir=$erisbuilddir/src/github.com/$this_user/$this_repo

# -----------------------------------------------------------------
# Get it

echo "Checking version..."
version=$(cat version)

# -----------------------------------------------------------------
# Build it

export ERIS_VERSION=$version
export ERIS_RELEASE=$arch_type
echo -e "Releasing =>\t\t\t$ERIS_VERSION:$ERIS_RELEASE"
rpmbuild -ba --sign $HOME/"$this_repo".spec

# -----------------------------------------------------------------
# Move into position

echo "Moving files into position"
cp -v $HOME/rpmbuild/RPMS/$ERIS_RELEASE/"$this_repo"-"$ERIS_VERSION"-"$ERIS_RELEASE"."$ERIS_RELEASE".rpm $host_location/$ERIS_RELEASE/
cp -v $HOME/rpmbuild/SRPMS/"$this_repo"-"$ERIS_VERSION"-"$ERIS_RELEASE".src.rpm  $host_location/sources/
gpg --armor --export 'support@erisindustries.com' > $host_location/RPM-GPG-KEY
cp -v $HOME/eris.repo $host_location/eris.repo

echo "Preparing repo files"
createrepo $host_location/$ERIS_RELEASE
createrepo $host_location/sources

# -----------------------------------------------------------------
# Cleanup

echo "All done!"
cd $start
