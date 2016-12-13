FROM quay.io/eris/node
MAINTAINER Eris Industries <support@erisindustries.com>

RUN npm install --global gulp

# set user
USER $USER
WORKDIR $ERIS
VOLUME $ERIS
ENTRYPOINT "gulp"