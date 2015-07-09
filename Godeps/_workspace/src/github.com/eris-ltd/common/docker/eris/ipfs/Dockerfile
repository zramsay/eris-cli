FROM eris/base
MAINTAINER Eris Industries <support@erisindustries.com>

# configure install
ENV NAME         ipfs
ENV REPO         ipfs/go-ipfs
ENV BRANCH       master
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
ENV IPFS_PATH $ERIS/files/ipfs_data
RUN mkdir --parents $IPFS_PATH

# boot
VOLUME $ERIS
EXPOSE 4001 5001 8080
CMD ["start"]
