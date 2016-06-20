#!/usr/bin/env bash

# Cleans up if tests are interrupted.
MACHINES=$(docker-machine ls -q)
if [ "$MACHINES" ]; then
    docker-machine rm --force $MACHINES
fi

