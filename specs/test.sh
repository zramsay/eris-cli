#!/usr/bin/env bash
set -e

export TEST_IN_CIRCLE=true
cd actions && go test -v
cd ../chains && go test -v
cd ../services && go test -v

cd ../util && go test -v