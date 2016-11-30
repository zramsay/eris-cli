FROM quay.io/eris/tools
MAINTAINER Eris Industries <support@erisindustries.com>

ENV NAME zcash
ENV LOCATION /var/zcash

# Install dependencies
RUN apt-add-repository -y ppa:bitcoin/bitcoin && \
  apt-get update -qq && \
  apt-get install -qq -y --no-install-recommends \
    g++ python-dev autotools-dev libicu-dev libbz2-dev wget \
    automake autoconf libtool bsdmainutils libdb4.8-dev \
    libdb4.8++-dev libssl-dev libevent-dev pkg-config libgmp-dev \
    libprocps3-dev libgtest-dev python-markdown unzip m4 g++-multilib \
    ncurses-dev zlib1g-dev && \
  apt-get upgrade -y -qq && \
  rm -rf /var/lib/apt/lists/*

# Install component
RUN git clone https://github.com/Electric-Coin-Company/$NAME.git /$LOCATION
WORKDIR /$LOCATION
RUN /$LOCATION/zcutil/fetch-params.sh
RUN /$LOCATION/zcutil/build.sh

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
ENV DATA_DIR /home/$USER/.zcash
RUN mkdir -p /home/$USER/.zcash && \
  touch $DATA_DIR/bitcoin.conf && \
  echo testnet=1 >> $DATA_DIR/bitcoin.conf && \
  echo addnode=alphatestnet.z.cash >> $DATA_DIR/bitcoin.conf && \
  echo rpcuser=eris >> $DATA_DIR/bitcoin.conf && \
  echo rpcpassword=Hze7btPVjmo6iJmUCdB5DXppCzZWaK8PEDHUJYYAYuFC >> $DATA_DIR/bitcoin.conf && \
  echo rpcconnect=zcash >> $DATA_DIR/bitcoin.conf && \
  chmod 600 $DATA_DIR/bitcoin.conf && \
  mkdir /root/.zcash && \
  cp $DATA_DIR/bitcoin.conf /root/.zcash/bitcoin.conf && \
  mv /root/.zcash-params /home/$USER/ && \
  chown $USER -R /home/$USER/
VOLUME $DATA_DIR
VOLUME /root/.zcash/

USER $USER
CMD $LOCATION/src/zcashd -conf=$DATA_DIR/bitcoin.conf -datadir=$DATA_DIR -daemon
