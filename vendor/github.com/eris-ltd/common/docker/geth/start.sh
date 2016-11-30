#!/usr/bin/env bash
set -e

# Note, most of this is based on this tutorial -> http://adeduke.com/2015/08/how-to-create-a-private-ethereum-chain/
#   ether-ites, please let us know if and when this needs to be updated for your purposes. <3 Eris.

#------------------------------------------------
# set and export directories

if [ "$CHAIN_ID" = "" ]; then
  echo "eris requires CHAIN_ID be set"
  exit 1
fi

# TODO: deal with chain numbers
# and eg. $CONTAINER_NAME
CHAIN_DIR="$ERIS/chains/$CHAIN_ID"

# set the eth directory
export ETH_PATH=$CHAIN_DIR

if [ ! -d "$ETH_PATH" ]; then
  mkdir -p $ETH_PATH
fi

# Test whether the mounted directory is writable for us
echo "Testing permissions to write to the eth directory: $ETH_PATH"
if ( touch $ETH_PATH/write_test 2>/dev/null ); then
  rm $ETH_PATH/write_test
else
  echo "ERR: $ETH_PATH is not writable for user 'eris' (UID 1000)"
  exit 1
fi

# ------------------------------------------------
# formulate the geth command properly

geth_cmd="geth --rpc \
  --datadir $ETH_PATH \
  --rpcaddr ${RPC_ADDR:=0.0.0.0} \
  --rpcport ${RPC_PORT:=8545} \
  --port ${PEER_PORT:=30303} \
  --identity ${NODE_ID:="geth-eris"} \
  --networkid ${NETWORKID:=1} \
  --gasprice ${GASPRICE:=50000000000} \
  --gpomin ${GASPRICEMIN:=50000000000} \
  --gpomax ${GASPRICEMAX:=500000000000} \
  --maxpeers ${MAX_PEERS:=25} \
  --maxpendpeers ${MAX_PEND_PEERS:=0} \
  --verbosity ${LOG_LEVEL:=3}"

if [ -z "$BOOTNODES" ]
then
  echo "No bootnodes given. Using geth defaults."
else
  echo "Setting bootnodes to: $BOOTNODES"
  geth_cmd="$geth_cmd --bootnodes $BOOTNODES"
fi

if [ -z "$BLOCKCHAINVERSION" ]
then
  echo "No blockchain version given. Using geth defaults."
else
  echo "Setting blockchain version to: $BLOCKCHAINVERSION"
  geth_cmd="$geth_cmd --blockchainversion $BLOCKCHAINVERSION"
fi

if [ -z "$MINE" ]
then
  echo "Leaving miner off."
else
  echo "Turning miner on."
  geth_cmd="$geth_cmd --mine"
fi

# Generally, for test chains, this should be on
if [ -z "$NODISCOVER" ]
then
  echo "Normal listening."
else
  echo "Setting to no discover. Manual peer entry required."
  geth_cmd="$geth_cmd --nodiscover"
fi

# “connections between nodes are valid only if peers have identical protocol version and network id”
if [ -z "$PROTOCOLVERSION" ]
then
  echo "No peer protocol version given. Using geth defaults."
else
  echo "Setting protocol version to: $PROTOCOLVERSION"
  geth_cmd="$geth_cmd --eth $PROTOCOLVERSION"
fi

# ------------------------------------------------
# Dump approriate files, if given

if [ -z "$GENESIS" ]
then
  echo "No genesis given. Using default."
else
  echo "Genesis given. Writing genesis.json"
  echo $GENESIS > $ETH_PATH/genesis.json
  geth_cmd="$geth_cmd --genesis $ETH_PATH/genesis.json"
fi

if [ -z "$KEY" ]
then
  echo "No Key Given. Checking if I'm to make an account."
  if [ ! -z "$MAKEACCT" ]
  then
    if [ -z "$PASSWORD" ]
    then
      echo "I cannot make an account without a password. Please rerun the container with a PASSWORD given."
      exit 1
    fi
    echo "Making and account file"
    export PASSWORD
    export CHAIN_ID
    export ETH_PATH
    new_account
  fi
else
  echo "Key Given. Writing keyfile"
  echo "$KEY" > $ERIS/key
  ADDR=$(cat $ERIS/key | jq '.address')
  ADDR=$(echo "$ADDR" | tr -d '"')
  mkdir -p $ETH_PATH/keystore/$ADDR
  mv $ERIS/key $ETH_PATH/keystore/$ADDR/$ADDR
fi

if [ -z "$ACCOUNT" ]
then
  echo "No account given."
  geth_cmd="$geth_cmd --etherbase 0"
else
  echo "Account given. $ACCOUNT"
  if [ -z "$PASSWORD" ]
  then
    echo "I cannot unlock the account without a password file. Please rerun the container with a PASSWORD given."
    exit 1
  fi
  echo "$PASSWORD" > $ETH_PATH/password
  geth_cmd="$geth_cmd --unlock $ACCOUNT --password $ETH_PATH/password --etherbase $ACCOUNT"
fi

if [ -z "$NODEKEY" ]
then
  echo "No nodekey given. Letting geth figure that out."
else
  echo "Node key given. Writing to proper place."
  echo "$NODEKEY" > $ETH_PATH/nodekey
fi

# ------------------------------------------------
# Start geth

echo -e "Starting ETH:\n\n$(geth version)\n\nThe marmot says wheeeeeeeeeeee....\n\n"
exec $geth_cmd
