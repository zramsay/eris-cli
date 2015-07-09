#!/usr/bin/env bash

# Test whether the mounted directory is writable for us
if ( touch $IPFS_PATH/write_test 2>/dev/null ); then
  rm $IPFS_PATH/write_test
else
  echo "ERR: $IPFS_PATH is not writable for user 'eris' (UID 1000)"
  exit 1
fi

printf "Starting IPFS:\n\n$(ipfs version)\n\nThe marmot says wheeeeeeeeeeee....\n\n"

if [ -e $IPFS_PATH/config ]; then
  echo "Found ipfs repository. Not initializing."
else
  ipfs init
  ipfs config Addresses.Gateway /ip4/${IP_ADDR:=0.0.0.0}/${GATE_PROTO:=tcp}/${GATE_PORT:=8080}
  ipfs config Addresses.API /ip4/${IP_ADDR:=0.0.0.0}/${API_PROTO:=tcp}/${API_PORT:=5001}
fi

ipfs daemon --writable --unrestricted-api