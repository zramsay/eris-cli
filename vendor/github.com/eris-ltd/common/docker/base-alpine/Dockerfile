FROM alpine@sha256:4b7f27ae8ce4ce6019ce41fa4275296f31b7b730b3eeb5fecf80f1b60959343d
MAINTAINER Eris Industries <support@erisindustries.com>

RUN apk update; apk add go git bash coreutils

# Add vendor support to Go
ENV GO15VENDOREXPERIMENT 1

# Go stuff
ENV GOPATH /go
ENV PATH $GOPATH/bin:/usr/local/go/bin:$PATH

RUN mkdir -p "$GOPATH/src" "$GOPATH/bin" && chmod -R 777 "$GOPATH"

# Where to install binaries
ENV INSTALL_BASE /usr/local/bin

# User Creation
# Choose the user id number 1000 to work well with Boot2Docker volumes.  See:
# https://github.com/boot2docker/boot2docker/issues/581#issuecomment-62491280
ENV USER eris
ENV ERIS /home/$USER/.eris
RUN addgroup -S $USER && adduser -s /bin/bash -S -D -u 1000 $USER $USER

RUN echo "alias ls='ls --color=never'" > /home/eris/.profile

# All common/go/common/dirs_and_files.go MajorDirs should be available for dumping into.
# except the scratch directories.
RUN mkdir $ERIS && \
  mkdir $ERIS/apps && \
  mkdir $ERIS/bundles && \
  mkdir $ERIS/chains && \
  mkdir $ERIS/chains/default && \
  mkdir $ERIS/chains/account-types && \
  mkdir $ERIS/chains/chain-types && \
  mkdir $ERIS/keys && \
  mkdir $ERIS/keys/data && \
  mkdir $ERIS/keys/names && \
  mkdir $ERIS/remotes && \
  mkdir $ERIS/services && \
  mkdir $ERIS/services/global && \
  chown --recursive $USER:$USER $ERIS

WORKDIR $ERIS

VOLUME /home/eris/.eris
