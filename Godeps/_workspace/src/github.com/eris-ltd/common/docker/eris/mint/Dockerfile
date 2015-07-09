# Pull base image.
FROM eris/base
MAINTAINER Eris Industries <support@erisindustries.com>

# grab deps (gmp)
RUN apt-get update && \
  apt-get install -y --no-install-recommends \
    libgmp3-dev && \
  rm -rf /var/lib/apt/lists/*

# configure install
ENV NAME         tendermint
ENV REPO         tendermint/tendermint
ENV BRANCH       develop
ENV BINARY_PATH  ./cmd/$NAME
ENV CLONE_PATH   $GOPATH/src/github.com/$REPO
ENV INSTALL_PATH $INSTALL_BASE/$NAME

# install
WORKDIR $CLONE_PATH
RUN git clone -q https://github.com/$REPO $CLONE_PATH && \
  git checkout -q $BRANCH && \
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
ENV TMROOT $ERIS/blockchains/tendermint
RUN mkdir --parents $TMROOT

# boot
VOLUME $ERIS
EXPOSE 46656 46657
CMD ["start"]

