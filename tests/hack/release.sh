#!/usr/bin/env bash
# assumes golang build from source or golang >= 1.5
# assumes one is on Casey's machine.
# this is still a WIP.

# Prerequisites -- Uncomment on new box
# go get github.com/aktau/github-release
# go get github.com/laher/goxc
# export GITHUB_TOKEN="token"

# -----------------------------------------------------------------
# Setting Defaults

aptmachine="eris-build-ams3-apt"
yummachine="eris-build-ams3-yum"
this_user="eris-ltd"
this_repo="eris-cli"
build_dir="builds"
cmd_path="cmd/eris"
pkg_name="eris"
repo=$GOPATH/src/github.com/$this_user/$this_repo
version=$(cat $repo/version/version.go | tail -n 1 | cut -d \  -f 4 | tr -d '"')

start=`pwd`

# -----------------------------------------------------------------
# Prerequisites

read -p "Have you done the [git tag -a v$version] and filled out the changelog yet? (Y/n) " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]
then
  echo "OK. Not doing anything. Rerun me after you've done that."
  exit 1
fi
echo "OK. Moving on then."
echo ""
echo ""

# -----------------------------------------------------------------
# Getting tags sorted && create a github release

echo "Pushing Tags to Github and Creating a Github Release"
latest_tag=$(git tag | tail -n 1 | cut -c 2-)
if [[ "$latest_tag" != "$version" ]]
then
  echo "Something isn't right. The last tagged version, does not match the version to be released. Exiting."
  exit 1
fi

echo "Tag check looks good. Sending the info to Github."
git push origin --tags
desc=$(git show v$version)
github-release release \
  --user $this_user \
  --repo $this_repo \
  --tag v$version \
  --name "Release of Version: $version" \
  --description "$desc"
echo "Finished sending tags and release info to Github."
echo ""
echo ""

# -----------------------------------------------------------------
# Cross Compile

echo "Starting Cross Compile"
cd $repo
mkdir $build_dir
goxc -wd $cmd_path -n $pkg_name -d $build_dir -pv $version
rm -rf $build_dir/$version/.goxc-temp
for dir in $build_dir/$version/*/
do
  rm -rf "$dir"
done
echo "Cross Compile Completed."
echo ""
echo ""

# -----------------------------------------------------------------
# Uploading Compiled Binaries to Github

echo "Uploading Binaries to Github."
for file in $build_dir/$version/*
do
  echo "Uploading: ${file##*/}"
  github-release upload \
    --user $this_user \
    --repo $this_repo \
    --tag v$version \
    --name "${file##*/}" \
    --file $file
done
echo "Uploading Complete."
echo ""
echo ""

# -----------------------------------------------------------------
# Send deb packages to APT repository

echo "Moving on to APT relase. Uploading files to APT server."
docker-machine scp $build_dir/$version/eris_"$version"_amd64.deb $aptmachine:~
docker-machine ssh $aptmachine
echo "Finished with APT release."

# -----------------------------------------------------------------
# Send rpm packages to YUM repository

echo "Moving on to YUM relase. Uploading files to YUM server."
echo "TODO........."
# docker-machine scp $build_dir/$version/eris_"$version"_amd64.rpm $yummachine:~
# docker-machine ssh $yummachine
echo "Finished with YUM release."

echo "Cleaning up and exiting... Billings Shipit!"
rm -rf $build_dir
cd $start