FROM quay.io/eris/tools
MAINTAINER Eris Industries <support@erisindustries.com>

RUN apt-get update -qq && \
  apt-get install -qq -y --no-install-recommends \
    geth && \
  rm -rf /var/lib/apt/lists/*

# start script
COPY start.sh $INSTALL_BASE/start
COPY new_account.sh $INSTALL_BASE/new_account

# set user
USER $USER
WORKDIR $ERIS

# configure $ETH_PATH is a default, will usually not go to that for test chains
ENV ETH_PATH $ERIS/chains/ethereum
RUN mkdir -p $ETH_PATH

# boot
VOLUME $ERIS
EXPOSE 8545 30303
CMD ["start"]

# default is to use keys from the host
# docker run --user eris -e CHAIN_ID=geth_test -e ACCOUNT=0 -e KEY=$(cat ~/.eris/keys/ethereum/whatever) -e PASSWORD="whatever" quay.io/eris/geth

# you'll need to make a key from inside the container and export it to your host if you have not make a key yet...
# docker run --rm --user eris -e PASSWORD=whatever quay.io/eris/geth new_account

# OR..... you can always just make it make an account and go for you
# docker run --user eris -e ACCOUNT=0 -e PASSWORD=whatever -e CHAIN_ID=geth_test -e MAKEACCT=true quay.io/eris/geth
