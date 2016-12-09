Circle CI is the **testing** and **mid-pathway** deployment tool for Eris. 

## Continuous Integration Overview

* Push to Github
* Circle CI clones the repo
* Circle CI steps through the `circle.yml` step by step according to their [docs](https://circleci.com/docs/getting-started). 
* At the end of most `circle.yml` we should either build (or if already built, push) a docker container to Docker Hub.
* From Docker Hub we can then set up webhooks to chain peer server clusters using our containers to do something (via Tutum).

## Shells

Each line in the `circle.yml` file is executed in its own `/bin/sh` instance. Environment variables not assigned in the CircleCI webui will not carry between shells. 

## Docker version

Note that as of 2015-02-26 the version of Docker on CircleCI is 1.4.1 which caused a build failure with a Dockerfile expecting a feature from version 1.5.

## docker exec

CircleCI doesn't support `docker exec`.  You can do something like the following but be aware you'll lose $HOME and $PATH:

```
# CircleCI doesn't support 'docker exec'.
if [ -n "$CI" ]; then
  docker_exec() {
    sudo lxc-attach -n $(docker-compose ps -q $1) -- bash -c "$2"
  }
else
  docker_exec() {
    docker exec $(docker-compose ps -q $1) $2
  }
fi
```

## Caching

If a build is funky and you can't figure it out. Try rebuilding and clearing the cache before you go too deep into a rabbit hole.  