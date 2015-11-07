#!/usr/bin/env bash

# ----------------------------------------------------------------------------
# Set definitions and defaults

# Where are the Things
base=github.com/eris-ltd/eris-cli
if [ "$CIRCLE_BRANCH" ]
then
  repo=${GOPATH%%:*}/src/github.com/${CIRCLE_PROJECT_USERNAME}/${CIRCLE_PROJECT_REPONAME}
  circle=true
else
  repo=$GOPATH/src/$base
  circle=false
fi
branch=${CIRCLE_BRANCH:=master}
branch=${branch/-/_}
epm_test_dir=$repo/../eris-pm/tests
epm_branch="master"

start=`pwd`

cd $repo

if [ -d "$epm_test_dir" ]; then
  cd $epm_test_dir
else
  cd ..
  git clone https://github.com/eris-ltd/eris-pm.git 1>/dev/null
  cd eris-pm
  git checkout origin/$epm_branch &>/dev/null
  cd $epm_test_dir
fi

./test.sh
test_exit=$?

cd $start
exit $test_exit
