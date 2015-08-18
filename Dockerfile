# NOTE => this dockerfile is used **ONLY** for testing
# please do not use this dockerfile for anything other
# than sandboxed testing of the cli
FROM ubuntu:14.04
MAINTAINER Eris Industries <support@erisindustries.com>

ENV DEBIAN_FRONTEND noninteractive

# DEPS
RUN apt-get update && apt-get install -y \
  curl wget gcc libc6-dev make ca-certificates \
  lxc apt-transport-https supervisor \
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

RUN curl -sSL -o /usr/local/bin/go-wrapper https://raw.githubusercontent.com/docker-library/golang/master/1.4/wheezy/go-wrapper
RUN chmod +x /usr/local/bin/go-wrapper

# DOCKER
RUN mkdir -p /var/log/supervisor
RUN mkdir -p /var/log/docker
COPY ./agent/supervisord.conf /etc/supervisor/conf.d/supervisord.conf

RUN echo deb https://get.docker.io/ubuntu docker main > /etc/apt/sources.list.d/docker.list \
  && apt-key adv --keyserver hkp://keyserver.ubuntu.com:80 --recv-keys 36A1D7869245C8950F966E92D8576A8BA88D21E9 \
  && apt-get update -qq \
  && apt-get install -qqy lxc-docker

RUN curl -sSL -o /usr/local/bin/wrapdocker https://raw.githubusercontent.com/jpetazzo/dind/master/wrapdocker
RUN chmod +x /usr/local/bin/wrapdocker

# SETUP USER
ENV USER eris
RUN groupadd --system $USER && \
  useradd --system --create-home --uid 1000 --gid $USER $USER && \
  usermod -a -G docker $USER

RUN mkdir /home/$USER/.eris
RUN chown --recursive $USER /home/$USER/.eris

# INSTALL CLI
RUN mkdir --parents /go/src/github.com/eris-ltd/eris-cli
COPY . /go/src/github.com/eris-ltd/eris-cli/
RUN cd /go/src/github.com/eris-ltd/eris-cli/cmd/eris && go install

COPY ./tests/test.sh /home/$USER/
WORKDIR /home/$USER

# CMD eris agent