#!/usr/bin/env bash

# Test whether the mounted directory is writable for us
if ( touch $ETH_PATH/write_test 2>/dev/null ); then
  rm $ETH_PATH/write_test
else
  echo "ERR: $ETH_PATH is not writable for user 'eris' (UID 1000)"
  exit 1
fi

# Start geth
printf "Starting ETH:\n\n$(geth version)\n\nThe marmot says wheeeeeeeeeeee....\n\n"

exec geth --rpc \
  --rpcaddr ${RPC_ADDR:=0.0.0.0} \
  --rpcport ${RPC_PORT:=8545} \
  --identity ${NODE_ID:="geth-eris"} \
  --maxpeers ${MAX_PEERS:=25} \
  --port ${PEER_PORT:=30303} \
  --datadir $ETH_PATH \
  --verbosity ${LOG_LEVEL:=1}
