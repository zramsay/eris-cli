#!/usr/bin/env bash

# -------------------------------------------------------------------
# Set vars (change if used in another repo)

base_name=eris-cli
user_name=eris-ltd
docs_site=docs.erisindustries.com

# -------------------------------------------------------------------
# Set vars (usually shouldn't be changed)

if [ "$CIRCLE_BRANCH" ]
then
  repo=`pwd`
else
  repo=$GOPATH/src/github.com/$user_name/$base_name
fi
release_min=$(cat $repo/version/version.go | tail -n 1 | cut -d \  -f 4 | tr -d '"')
start=`pwd`

# -------------------------------------------------------------------
# Build

mkdir -p docs/$base_name
go run docs/generator.go

if [[ "$1" == "master" ]]
then
  mkdir -p docs/$base_name/latest
  rsync -av docs/$base_name/$release_min/ docs/$base_name/latest/
  find docs/$base_name/latest -type f -name "*.md" -exec sed -i "s/$release_min/latest/g" {} +
fi

cd $HOME
git clone git@github.com:$user_name/$docs_site.git
cd $repo

rsync -av docs/$base_name $HOME/$docs_site/documentation/

# ------------------------------------------------------------------
# Commit and push if there's changes

cd $HOME/$docs_site
if [ -z "$(git status --porcelain)" ]; then
  echo "All Good!"
else
  git add -A :/ &&
  git commit -m "$base_name build number $CIRCLE_BUILD_NUM doc generation" &&
  git push origin master
fi

# ------------------------------------------------------------------
# Cleanup

cd $start