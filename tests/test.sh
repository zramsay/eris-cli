#!/usr/bin/env sh
set -e

passed() {
  echo ""
  echo ""
  echo "$1 Passed."
  echo ""
  echo ""
}

if [ -z "$CIRCLE_BUILD_NUM" ]; then
  echo "Testing NOT in Circle Environment."
  eris services start ipfs
  sleep 3
else
  echo "Testing in Circle Environment."
  export TEST_IN_CIRCLE=true
fi

cd perform && go test
passed Perform
cd ../util && go test
passed Util
cd ../data && go test
passed Data
if [ -z "$CIRCLE_BUILD_NUM" ]; then
  cd ../files && go test # circle hates these
  passed Files
fi
cd ../config && go test
passed Config
eris services stop ipfs -rx
cd ../services && go test
passed Services
cd ../chains && go test
passed Chains
cd ../actions && go test
passed Actions
cd ../contracts && go test
passed Contracts
# cd ../projects && go test
# passed Projects
# cd ../remotes && go test
# passed Remotes
cd ../commands && go test
passed commands

echo "Congratulations! All Tests Passed. We're Green"
echo ""
echo ""
