FROM eris/base
MAINTAINER Eris Industries <support@erisindustries.com>

# dependencies => gmp, godeps
RUN apt-get update -qq && \
  apt-get install -y -qq --no-install-recommends \
    libgmp3-dev && \
  rm -rf /var/lib/apt/lists/*
RUN go get github.com/tools/godep

# configure install
ENV NAME         geth
ENV REPO         ethereum/go-ethereum
ENV BRANCH       develop
ENV BINARY_PATH  ./cmd/$NAME
ENV CLONE_PATH   $GOPATH/src/github.com/$REPO
ENV INSTALL_PATH $INSTALL_BASE/$NAME

# install
WORKDIR $CLONE_PATH
RUN git clone -q https://github.com/$REPO $CLONE_PATH && \
  git checkout -q $BRANCH && \
  godep restore && \
  cd $BINARY_PATH && go get -d . && cd ../../ && \
  go build -o $INSTALL_PATH $BINARY_PATH

# cleanup install
RUN rm -rf $GOPATH/src/* && \
  unset NAME && \
  unset INSTALL_BASE && \
  unset REPO && \
  unset CLONE_PATH && \
  unset BINARY_PATH && \
  unset INSTALL_PATH && \
  unset BRANCH

# start script
COPY start.sh $INSTALL_BASE/start

# set user
USER $USER
WORKDIR $ERIS

# configure
ENV ETH_PATH $ERIS/blockchains/ethereum
RUN mkdir --parents $ETH_PATH

# boot
VOLUME $ERIS
EXPOSE 8545 30303
CMD ["start"]
