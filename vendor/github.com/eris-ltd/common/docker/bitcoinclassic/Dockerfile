FROM quay.io/eris/tools
MAINTAINER Eris Industries <support@erisindustries.com>

ENV NAME bitcoinclassic

RUN apt-add-repository -y ppa:bitcoin/bitcoin && \
  apt-get update -qq && \
  apt-get install -qq -y --no-install-recommends \
    automake autoconf libtool bsdmainutils libboost1.54-all-dev libdb4.8-dev libdb4.8++-dev libssl-dev libevent-dev pkg-config && \
  rm -rf /var/lib/apt/lists/*

RUN git clone https://github.com/bitcoinclassic/bitcoinclassic.git /tmp/$NAME && \
  cd /tmp/$NAME && \
  ./autogen.sh && \
  ./configure --without-qt && \
  make && \
  make install && \
  rm -rf /tmp/$NAME

# Expose mainnet wallet rpc port.
EXPOSE 8332

# Expose mainnet peer port.
EXPOSE 8333

# Expose mainnet rpc port.
EXPOSE 8334

# Expose testnet wallet rpc port.
EXPOSE 18332

# Expose testnet peer port.
EXPOSE 18333

# Expose testnet rpc port.
EXPOSE 18334

# persist data, set user
WORKDIR $ERIS
ENV DATA_DIR $ERIS/scratch/data/$NAME
RUN mkdir -p $DATA_DIR && \
  touch $DATA_DIR/bitcoin.conf && \
  echo rpcuser=bitcoinrpc >> $DATA_DIR/bitcoin.conf && \
  echo rpcpassword=Hze7btPVjmo6iJmUCdB5DXppCzZWaK8PEDHUJYYAYuFC >> $DATA_DIR/bitcoin.conf && \
  chmod 600 $DATA_DIR/bitcoin.conf && \
  chown $USER -R /home/$USER/

VOLUME /home/$USER/.eris

USER $USER
CMD bitcoind -conf=$DATA_DIR/bitcoin.conf -datadir=$DATA_DIR -daemon
