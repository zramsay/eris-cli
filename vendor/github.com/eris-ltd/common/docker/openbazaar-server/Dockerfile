FROM quay.io/eris/tools
MAINTAINER Eris Industries <support@erisindustries.com>

# Dependencies
RUN touch /usr/local/man
RUN add-apt-repository ppa:chris-lea/libsodium && \
  echo "deb http://ppa.launchpad.net/chris-lea/libsodium/ubuntu trusty main" >> /etc/apt/sources.list && \
  apt-get -qq update && apt-get -y upgrade && \
  apt-get install -y -qq \
    autoconf \
    libffi-dev \
    libssl-dev \
    libtool \
    libzmq3-dev \
    openssl \
    pkg-config \
    python-pip \
    pylint \
    libsodium-dev && \
  pip install cryptography

RUN git clone https://github.com/OpenBazaar/OpenBazaar-Server.git /OpenBazaar-Server
WORKDIR /OpenBazaar-Server

RUN pip install -r requirements.txt && \
  chown --recursive $USER /OpenBazaar-Server

# set user
USER $USER
WORKDIR $ERIS

# boot
VOLUME $ERIS
CMD python openbazaard.py start -a 0.0.0.0
