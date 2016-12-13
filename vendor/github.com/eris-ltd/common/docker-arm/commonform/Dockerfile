FROM quay.io/eris/node:arm

RUN npm --global install commonform-cli

# set user
USER $USER
WORKDIR $ERIS
VOLUME $ERIS
ENTRYPOINT commonform
