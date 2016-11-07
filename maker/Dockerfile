FROM quay.io/eris/build
MAINTAINER Monax <support@monax.io>

# Install eris-cm, a go app that manages chains
ENV TARGET eris-cm
ENV REPO $GOPATH/src/github.com/eris-ltd/$TARGET

COPY ./glide.yaml $REPO/glide.yaml
COPY ./glide.lock $REPO/glide.lock
WORKDIR $REPO
RUN glide install

COPY . $REPO/.
RUN cd $REPO/cmd/$TARGET && \
  go build --ldflags '-extldflags "-static"' -o $INSTALL_BASE/$TARGET && \
  unset TARGET && \
  unset REPO

#-----------------------------------------------------------------------------
# install mintgen [to be deprecated]
ENV ERIS_GEN_MINT_REPO github.com/eris-ltd/mint-client
ENV ERIS_GEN_MINT_SRC_PATH $GOPATH/src/$ERIS_GEN_MINT_REPO

WORKDIR $ERIS_GEN_MINT_SRC_PATH

RUN git clone --quiet https://$ERIS_GEN_MINT_REPO . \
  && git checkout --quiet master \
  && go build --ldflags '-extldflags "-static"' -o $INSTALL_BASE/mintgen ./mintgen \
  && unset ERIS_GEN_MINT_REPO \
  && unset ERIS_GEN_MINT_SRC_PATH
# [end to be deprecated]
#-----------------------------------------------------------------------------
