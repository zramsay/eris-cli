FROM quay.io/eris/gulp
MAINTAINER Eris Industries <support@erisindustries.com>

USER root

RUN npm install --global s-unit

# set user
USER $USER
WORKDIR $ERIS
VOLUME $ERIS
ENTRYPOINT "sunit"