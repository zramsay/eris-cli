FROM eris/base:latest
MAINTAINER Eris Industries <support@erisindustries.com>

ENV source $GOPATH/src/github.com/eris-ltd/eris-cli
COPY . $source
WORKDIR $source/cmd/eris
RUN go install

USER $USER
WORKDIR /home/$USER
