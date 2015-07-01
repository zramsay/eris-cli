#!/usr/bin/env bash
set -e

export TEST_IN_CIRCLE=true
cd services && go test -v -timeout 30m
