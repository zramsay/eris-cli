FROM quay.io/eris/base:arm
MAINTAINER Monax Industries <support@monax.io>

# configure install
ENV NAME         ipfs
ENV REPO         ipfs/go-ipfs
ENV TAG          v0.4.0
ENV BINARY_PATH  ./cmd/$NAME
ENV CLONE_PATH   $GOPATH/src/github.com/$REPO
ENV INSTALL_PATH $INSTALL_BASE/$NAME
#dir files was rm from common -shouldn't be a problem unless base is updated...?
ENV IPFS_PATH $ERIS/files/ipfs_data
ENV GX_IPFS ""

# install
WORKDIR $CLONE_PATH

# install gx
RUN go get -u github.com/whyrusleeping/gx
RUN go get -u github.com/whyrusleeping/gx-go
RUN git clone -q https://github.com/$REPO $CLONE_PATH

# Point gx to a specific IPFS API
RUN ([ -z "$GX_IPFS" ] || echo $GX_IPFS > $IPFS_PATH/api)

# Invoke gx
RUN cd $CLONE_PATH && \
  git checkout -q tags/$TAG && \
  gx --verbose install --global

RUN go build -o $INSTALL_PATH $BINARY_PATH

# We get the current commit using this hack,
# so that we don't have to copy all of .git/ into the build context.
# This saves us quite a bit of image size.
#RUN ref="$(cat .git/HEAD | cut -d' ' -f2)"
#RUN commit="$(cat .git/$ref | head -c 7)"
#RUN echo "ldflags=-X github.com/ipfs/go-ipfs/repo/config.CurrentCommit=$commit"
# Build and install IPFS and entrypoint script
#RUN cd $CLONE_PATH/cmd/ipfs
#RUN go build -ldflags "-X github.com/ipfs/go-ipfs/repo/config.CurrentCommit=$commit"

# cleanup install
RUN rm -rf $GOPATH/src/* && \
  unset NAME && \
  unset INSTALL_BASE && \
  unset REPO && \
  unset CLONE_PATH && \
  unset BINARY_PATH && \
  unset INSTALL_PATH && \
  unset TAG

# start script
COPY start.sh $INSTALL_BASE/start

# set user
USER $USER
WORKDIR $ERIS

# configure
RUN mkdir -p $IPFS_PATH

# boot
VOLUME $ERIS
EXPOSE 4001 5001 8080
CMD ["start"]
