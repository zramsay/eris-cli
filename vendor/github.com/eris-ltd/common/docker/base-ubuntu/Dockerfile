FROM golang:1.5
MAINTAINER Eris Industries <support@erisindustries.com>

# shut up debian
ENV DEBIAN_FRONTEND noninteractive
ENV DEBIAN_PRIORITY critical
ENV DEBCONF_NOWARNINGS yes
ENV TERM linux
RUN echo 'debconf debconf/frontend select Noninteractive' | debconf-set-selections

# Where to install binaries
ENV INSTALL_BASE /usr/local/bin

# User Creation
# Choose the user id number 1000 to work well with Boot2Docker volumes.  See:
# https://github.com/boot2docker/boot2docker/issues/581#issuecomment-62491280
ENV USER eris
ENV ERIS /home/$USER/.eris
RUN groupadd --system $USER \
  && useradd --system --create-home --uid 1000 --gid $USER $USER

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
