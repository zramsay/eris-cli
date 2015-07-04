#!/usr/bin/env sh
set -e

passed() {
  echo ""
  echo ""
  echo "$1 Passed."
  echo ""
  echo ""
}

export TEST_IN_CIRCLE=true
cd actions && go test -v
passed Actions
cd ../chains && go test -v
passed Chains
cd ../services && go test -v
passed Services

cd ../config && go test -v
passed Config
cd ../util && go test -v
passed Util
