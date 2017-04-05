FROM alpine:3.5
MAINTAINER Monax | Platform | Tools <support@monax.io>

# Install bash and coreutils
RUN apk --no-cache --update add bash coreutils

# Where to install binaries
ENV INSTALL_BASE /usr/local/bin

# User Creation
# Choose the user id number 1000 to work well with Boot2Docker volumes.  See:
# https://github.com/boot2docker/boot2docker/issues/581#issuecomment-62491280
ENV USER monax
ENV MONAX_PATH /home/$USER/.monax
RUN addgroup -S $USER && adduser -s /bin/bash -S -D -u 1000 $USER $USER
RUN echo "alias ls='ls --color=never'" > /home/monax/.profile

# All common/go/common/dirs_and_files.go MajorDirs should be available for dumping into.
# except the scratch directories.
RUN mkdir $MONAX_PATH && \
  mkdir $MONAX_PATH/apps && \
  mkdir $MONAX_PATH/bundles && \
  mkdir $MONAX_PATH/chains && \
  mkdir $MONAX_PATH/chains/account-types && \
  mkdir $MONAX_PATH/chains/chain-types && \
  mkdir $MONAX_PATH/keys && \
  mkdir $MONAX_PATH/keys/data && \
  mkdir $MONAX_PATH/keys/names && \
  mkdir $MONAX_PATH/services && \
  chown --recursive $USER:$USER $MONAX_PATH