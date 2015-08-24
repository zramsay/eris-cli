#!/usr/bin/env bash
# do not use; still a WIP

docker build -t eris/eris -f tests/Dockerfile .

docker run --rm --entrypoint "/home/eris/test_runner.sh" eris/eris