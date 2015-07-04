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
cd perform && go test -v
passed Perform
cd ../util && go test -v
passed Util
cd ../services && go test -v
passed Services
cd ../chains && go test -v
passed Chains
cd ../actions && go test -v
passed Actions
cd ../projects && go test -v
passed Projects
cd ../remotes && go test -v
passed Remotes
cd ../data && go test -v
passed Data
cd ../files && go test -v
passed Files
cd ../init && go test -v
passed Init
cd ../config && go test -v
passed Config
