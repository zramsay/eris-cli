#!/usr/bin/env bash
# assumes golang >= 1.5
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

if [[ "$1" == "pre" ]]
then
  if [[ "$2" == "" ]]
  then
    echo "you must tell me which release candidate this is... 1, 2, 3, etc. exiting."
    exit 1
  fi
  version=$version-rc$2
fi

# -----------------------------------------------------------------
# Prerequisites

pre_check() {
  read -p "Have you done the [git tag -a v$version] and filled out the changelog yet? (y/n) " -n 1 -r
  echo
  if [[ ! $REPLY =~ ^[Yy]$ ]]
  then
    echo "OK. Not doing anything. Rerun me after you've done that."
    exit 1
  fi
  echo "OK. Moving on then."
  echo ""
  echo ""
  latest_tag=$(git tag | tail -n 1 | cut -c 2-)
  if [[ "$latest_tag" != "$version" ]]
  then
    echo "Something isn't right. The last tagged version, does not match the version to be released. Exiting."
    exit 1
  fi
}

# -----------------------------------------------------------------
# Cross Compile

cross_compile() {
  echo "Starting Cross Compile"
  cd $repo
  mkdir $build_dir
  goxc -wd $cmd_path -n $pkg_name -d $build_dir -pv $version
  rm -rf $build_dir/$version/.goxc-temp
  for dir in $build_dir/$version/*/
  do
    rm -rf "$dir"
  done
  cd $start
  echo "Cross Compile Completed."
  echo ""
  echo ""
}

# -----------------------------------------------------------------
# Uploading Compiled Binaries to Github

prepare_gh() {
  echo "Pushing Tags to Github and Creating a Github Release"
  git push origin --tags
  desc=$(git show v$version)
  if [[ "$1" == "pre" ]]
  then
    github-release release \
      --user $this_user \
      --repo $this_repo \
      --tag v$version \
      --name "Release of Version: $version" \
      --description "$desc" \
      --pre-release
  else
    github-release release \
      --user $this_user \
      --repo $this_repo \
      --tag v$version \
      --name "Release of Version: $version" \
      --description "$desc"
  fi
  echo "Finished sending tags and release info to Github."
  echo ""
  echo ""
}

release_gh() {
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
}

# -----------------------------------------------------------------
# Send deb packages to APT repository

release_apt() {
  echo "Moving on to APT relase. Uploading files to APT server."
  docker-machine scp $repo/tests/hack/release_deb.sh $aptmachine:~
  docker-machine scp $build_dir/$version/eris_"$version"_amd64.deb $aptmachine:~
  docker-machine ssh $aptmachine
  echo "Finished with APT release."
}

# -----------------------------------------------------------------
# Send rpm packages to YUM repository

release_yum() {
  echo "Moving on to YUM relase. Uploading files to YUM server."
  docker-machine scp $repo/tests/hack/eris-cli.spec $yummachine:~
  docker-machine scp $repo/tests/hack/eris.repo $yummachine:~
  docker-machine scp $repo/tests/hack/release_rpm.sh $yummachine:~
  docker-machine scp $build_dir/$version/linux_amd64/eris $yummachine:~
  docker-machine ssh $yummachine "echo \"$version\" > version"
  docker-machine ssh $yummachine
  echo "Finished with YUM release."
}

# ---------------------------------------------------------------
# Cleanup

clean_up() {
  echo "Cleaning up and exiting... Billings Shipit!"
  rm -rf $build_dir
  cd $start
  exit $1
}

preRun() {
  pre_check "$@"
  cross_compile "$@"
  prepare_gh "$@"
  release_gh "$@"
  release_apt "$@"
  release_yum "$@"
  clean_up $?
}

fullRun() {
  pre_check "$@"
  cross_compile "$@"
  prepare_gh "$@"
  release_gh "$@"
  release_apt "$@"
  release_yum "$@"
  clean_up $?
}

main() {
  if [[ "$1" == "build" ]]
  then
    cross_compile "$@"
    return $?
  fi
  if [[ "$1" == "pre" ]]
  then
    preRun "$@"
    return $?
  fi
  fullRun "$@"
}

main "$@"