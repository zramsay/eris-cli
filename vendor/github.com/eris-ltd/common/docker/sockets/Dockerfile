FROM quay.io/eris/base
MAINTAINER Eris Industries <support@erisindustries.com>

ENV SOCKS_DIR /home/$USER/sockets

RUN mkdir $SOCKS_DIR
RUN chown -R $USER:$USER $SOCKS_DIR

WORKDIR $SOCKS_DIR
VOLUME $SOCKS_DIR
