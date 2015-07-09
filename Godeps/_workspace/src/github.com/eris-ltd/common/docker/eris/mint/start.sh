#! /bin/bash

if [ "$FAST_SYNC" = "true" ]; then
	tendermint node --fast_sync
else 
	tendermint node
fi

