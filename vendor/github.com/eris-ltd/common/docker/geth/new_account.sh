#!/usr/bin/env bash
set -e

# NOTE: I should be ran with the following command:
#
# docker run --rm --user eris -e PASSWORD=whatever eris/eth new_account
#
# or
#
# docker run -it --rm --user eris eris/eth new_account (if you want to type it in manually)
#
# then copy the exported key to a file on your host.

# ------------------------------------------------
# set and export directories

if [ "$CHAIN_ID" = "" ]
then
  echo "CHAIN_ID not set. Using eth defaults"
else
  CHAIN_DIR="$ERIS/chains/$CHAIN_ID"
  export ETH_PATH=$CHAIN_DIR
fi

if [ ! -d "$ETH_PATH" ]
then
  mkdir -p $ETH_PATH
fi

# ------------------------------------------------
# formulating the new accounts command

geth_cmd="geth --datadir $ETH_PATH"

if [ -z "$PASSWORD" ]
then
  echo "No password given."
else
  echo "$PASSWORD" > $ETH_PATH/password
  geth_cmd="$geth_cmd --password $ETH_PATH/password"
fi

geth_cmd="$geth_cmd account new"

# ------------------------------------------------
# running new accounts command and displaying the files

$geth_cmd

file=$(ls $ETH_PATH/keystore/)
echo -e "\n\nCopy the below and paste into a file on your host.\n\n"
cat $ETH_PATH/keystore/$file
echo -e "\n\n"
