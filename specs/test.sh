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

# testing these here because go test wierdness
prev=$(eris services export ipfs)
eris services -v import 1234 ipfs:$prev
eris services known
eris services ls
eris services ps
prev=$(eris chains export testchain)
eris chains -v import test2 ipfs:$prev
eris chains known
eris chains ls
eris chains ps
prev=$(eris actions export do not use)
eris actions -v import "use this bad boy" ipfs:$prev
eris actions known
eris actions ls
eris actions ps
passed "Import / Export"