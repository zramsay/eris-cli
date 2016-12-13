FROM quay.io/eris/base
MAINTAINER Eris Industries <support@erisindustries.com>

RUN go get -u github.com/btcsuite/btcd/...

# Expose mainnet listening port.
EXPOSE 8333

# Expose mainnet rpc port.
EXPOSE 8334

# Expose mainnet wallet rpc port.
EXPOSE 8332

# Expose testnet listening port.
EXPOSE 18333

# Expose testnet rpc port.
EXPOSE 18334

# Expose testnet wallet rpc port.
EXPOSE 18332

# persist data, set user
WORKDIR /home/$USER/
RUN mkdir .btcd && \
  mkdir .btcctl && \
  chown $USER -R /home/$USER/

VOLUME /home/$USER/.btcd
VOLUME /home/$USER/.btcctl
VOLUME /home/$USER/.eris

USER $USER
CMD ["/go/bin/btcd"]
