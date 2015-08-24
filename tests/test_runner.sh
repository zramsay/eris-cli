#!/usr/bin/env bash
# do not use; still a WIP
set -e


declare -a docker_version=( "1.6.0" "1.7.0" "1.7.1" )

for ver in "${docker_version[@]}"
do
  echo ""
  echo ""
  echo "Testing against docker version $ver"
  echo ""
  echo ""
  curl -sSL -o /usr/bin/docker https://get.docker.com/builds/Linux/x86_64/docker-$ver
  docker -d &
  docker run --rm -v /var/run/docker.sock:/var/run/docker.sock --entrypoint "/home/eris/test.sh" --user eris eris/eris
done
