FROM quay.io/eris/build
MAINTAINER Monax <support@monax.io>

# Install eris-keys, a go app for development signing
ENV TARGET eris-keys
ENV REPO $GOPATH/src/github.com/eris-ltd/$TARGET

ADD ./glide.yaml $REPO/
ADD ./glide.lock $REPO/
WORKDIR $REPO
RUN glide install

COPY . $REPO/.
RUN cd $REPO/cmd/$TARGET && \
  go build --ldflags '-extldflags "-static"' -o $INSTALL_BASE/$TARGET

# build customizations start here
# install mint-key [to be deprecated]
ENV ERIS_KEYS_MINT_REPO github.com/eris-ltd/mint-client
ENV ERIS_KEYS_MINT_SRC_PATH $GOPATH/src/$ERIS_KEYS_MINT_REPO

WORKDIR $ERIS_KEYS_MINT_SRC_PATH

RUN git clone --quiet https://$ERIS_KEYS_MINT_REPO . \
  && git checkout --quiet master \
  && go build --ldflags '-extldflags "-static"' -o $INSTALL_BASE/mintkey ./mintkey \
  && unset ERIS_KEYS_MINT_REPO \
  && unset ERIS_KEYS_MINT_SRC_PATH
