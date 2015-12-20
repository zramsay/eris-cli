# NOTE => this dockerfile is used **ONLY** for testing
# please do not use this dockerfile for anything other
# than sandboxed testing of the cli
FROM ubuntu:14.04
MAINTAINER Eris Industries <support@erisindustries.com>

ENV DEBIAN_FRONTEND noninteractive
ENV DEBIAN_PRIORITY critical
ENV DEBCONF_NOWARNINGS yes
ENV TERM linux
RUN echo 'debconf debconf/frontend select Noninteractive' | debconf-set-selections

# Where to install binaries
ENV INSTALL_BASE /usr/local/bin

# DEPS
RUN apt-get update && apt-get install -y \
  curl wget gcc libc6-dev make ca-certificates \
  lxc apt-transport-https supervisor jq \
  --no-install-recommends \
  && rm -rf /var/lib/apt/lists/*

# GOLANG
ENV GOLANG_VERSION 1.4.2
RUN curl -sSL https://golang.org/dl/go$GOLANG_VERSION.src.tar.gz \
  | tar -v -C /usr/src -xz
RUN cd /usr/src/go/src && ./make.bash --no-clean 2>&1
ENV PATH /usr/src/go/bin:$PATH
RUN mkdir -p /go/src /go/bin && chmod -R 777 /go
ENV GOPATH /go
ENV PATH /go/bin:$PATH
WORKDIR /go

# GO WRAPPER
ENV GO_WRAPPER_VERSION 1.4
RUN curl -sSL -o $INSTALL_BASE/go-wrapper https://raw.githubusercontent.com/docker-library/golang/master/$GO_WRAPPER_VERSION/wheezy/go-wrapper
RUN chmod +x $INSTALL_BASE/go-wrapper

# DOCKER
RUN mkdir -p /var/log/docker
RUN echo deb https://apt.dockerproject.org/repo ubuntu-trusty main > /etc/apt/sources.list.d/docker.list \
  && apt-key adv --keyserver hkp://p80.pool.sks-keyservers.net:80 --recv-keys 58118E89F3A912897C070ADBF76221572C52609D \
  && apt-get update -qq \
  && apt-get install -qqy docker-engine

# DOCKER WRAPPER
RUN curl -sSL -o $INSTALL_BASE/wrapdocker https://raw.githubusercontent.com/jpetazzo/dind/master/wrapdocker
RUN chmod +x $INSTALL_BASE/wrapdocker

# DOCKER-MACHINE (for testing)
ENV DOCKER_MACHINE_VERSION 0.4.0
RUN curl -sSL -o $INSTALL_BASE/docker-machine \
  https://github.com/docker/machine/releases/download/v$DOCKER_MACHINE_VERSION/docker-machine_linux-amd64 && \
  chmod +x $INSTALL_BASE/docker-machine

# INSTALL CLI
ENV REPO github.com/eris-ltd/eris-cli
ENV BASE $GOPATH/src/$REPO
ENV NAME eris
RUN mkdir --parents $BASE
COPY . $BASE/
RUN cd $BASE/cmd/eris && go build -o $INSTALL_BASE/$NAME

# SETUP USER
ENV USER eris
ENV ERIS /home/$USER/.eris
RUN groupadd --system $USER && \
  useradd --system --create-home --uid 1000 --gid $USER $USER && \
  usermod -a -G docker $USER
RUN mkdir $ERIS && \
  mkdir /home/$USER/.docker
RUN mv $BASE/tests/* /home/$USER
RUN chown --recursive $USER:$USER /home/$USER
RUN chown --recursive $USER:$USER /go

USER $USER
WORKDIR /home/$USER
ENTRYPOINT ["eris"]
