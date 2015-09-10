#! /bin/bash

ifExit(){
	if [ $? -ne 0 ]; then
		echo "ifExit"
		echo "$1"
		exit 1
	fi
}


if [ "$1" == "clean" ]; then
eris chains stop -rxf etcb
eris chains stop -rxf mychain
docker rm -vf eris_chain_mychain_2 eris_data_mychain_2
exit

fi

# this test starts an etcb chain, then generates a new chain, registers it on the etcb, and then installs it from the etcb

eris services start keys

# first, generate a couple keys
ADDR1=`eris services exec keys eris-keys gen`
ADDR1=${ADDR1%?}
echo "addr1: $ADDR1"
PUB1=`eris services exec keys "eris-keys pub --addr $ADDR1"`
PUB1=${PUB1%?}
echo "pub1: $PUB1"

ADDR2=`eris services exec keys eris-keys gen`
ADDR2=${ADDR2%?}
echo "addr2: $ADDR2"
PUB2=`eris services exec keys "eris-keys pub --addr $ADDR2"`
PUB2=${PUB2%?}
echo "pub2: $PUB2"

# genesis csvs
echo "$PUB1,10000" > vals.csv
echo "$PUB2,1000000" > accs.csv

# get the priv val
docker run --rm --volumes-from eris_data_keys_1 -t --entrypoint mintkey eris/erisdb:0.10.3 mint $ADDR1 > priv1.json
docker run --rm --volumes-from eris_data_keys_1 -t --entrypoint mintkey eris/erisdb:0.10.3 mint $ADDR2 > priv2.json

# now create the etcb chain
eris chains new --priv=priv1.json --csv="vals.csv,accs.csv" --options="skip-upnp=true" etcb -p --api
ifExit "failed to create etcb chain"
echo "new etcb chain container"

# set some vars for finding etcb
ETCB_PORT=46657
ETCB_HOST=etcb_host

# boot a new chain 
CHAIN_ID=mychain
eris chains new $CHAIN_ID --csv=accs.csv --priv=priv2.json  --api
ifExit "failed to create new chain"
echo "created new chain $CHAIN_ID"

# cleanup
rm vals.csv accs.csv priv1.json priv2.json

# now we register the chain on etcb
# the seed is the running new container (assume container doing "install" is linked to it)
NEW_SEED="the_seed"
eris chains register $CHAIN_ID "$NEW_SEED:46656" --pub=$PUB2 --links="eris_chain_etcb_1:$ETCB_HOST" --etcb-host="$ETCB_HOST:$ETCB_PORT" --etcb-chain=etcb
ifExit "failed to register chain"
echo "registered chain $CHAIN_ID"


# now lets install the chain
eris chains install -p $CHAIN_ID -N=2 --links="eris_chain_etcb_1:$ETCB_HOST,eris_chain_${CHAIN_ID}_1:$NEW_SEED" --etcb-host="$ETCB_HOST:$ETCB_PORT"
ifExit "failed to install chain"
echo "installed chain $CHAIN_ID"

# let it boot
sleep 3

# ensure both chains have the same genesis
G1=`docker run -it --rm --link eris_chain_mychain_1:eris1 --entrypoint mintinfo eris/erisdb:0.10.3 --node-addr=eris1:46657 genesis`
G2=`docker run -it --rm --link eris_chain_mychain_2:eris2 --entrypoint mintinfo eris/erisdb:0.10.3 --node-addr=eris2:46657 genesis`

if [ "$G1" != "$G2" ]; then
 echo "genesis files from the two chains dont match"
 echo "GENESIS 1"
 echo "$G1"
 echo ""
 echo "GENESIS 2"
 echo "$G2"

 exit 1
fi

# sleep a little to get some blocks mined
sleep 3

# ensure both chains have some positive block height
BH1=`docker run -it --rm --link eris_chain_mychain_1:eris1 --entrypoint mintinfo eris/erisdb:0.10.3 --node-addr=eris1:46657 status latest_block_height`
BH2=`docker run -it --rm --link eris_chain_mychain_2:eris2 --entrypoint mintinfo eris/erisdb:0.10.3 --node-addr=eris2:46657 status latest_block_height`

echo "block heights"
ATLEAST=2

if  [[ ("$BH1" < "$ATLEAST") ]]; then
	echo "expected block height to be greater than 1. Got $BH1"
	exit 1
fi

if  [[ ("$BH2" < "$ATLEAST") ]]; then
	echo "expected block height to be greater than 1. Got $BH2"
	exit 1
fi

echo "REGISTER/INSTALL SUCCESS!"

#eris chains stop -rxf etcb
#eris chains stop -rxf mychain
