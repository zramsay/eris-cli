# Pull base image.
FROM quay.io/eris/base:arm
MAINTAINER Eris Industries <support@erisindustries.com>


# configure install
ENV NAME         tendermint
ENV REPO 	tendermint/tendermint
ENV BRANCH       master
ENV BINARY_PATH  ./cmd/$NAME
ENV CLONE_PATH   $GOPATH/src/github.com/$REPO

# install
WORKDIR $CLONE_PATH
RUN git clone -q https://github.com/$REPO $CLONE_PATH && \
  git checkout -q $BRANCH && \
  go get $BINARY_PATH && \ 
  go install $BINARY_PATH

# set user
USER $USER
WORKDIR $ERIS

# configure
ENV TMROOT $ERIS/chains/tendermint
RUN mkdir --parents $TMROOT

# boot
VOLUME $ERIS
EXPOSE 46656 46657
ENTRYPOINT ["tendermint"]
