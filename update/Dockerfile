FROM quay.io/eris/base
MAINTAINER Eris Industries <support@erisindustries.com>

ENV NAME         eris-cli
ENV REPO 	 eris-ltd/$NAME
ENV BRANCH       $BRANCH_TO_UPDATE #passed it when cli does `docker run`
ENV BINARY_PATH  $NAME
ENV CLONE_PATH   $GOPATH/src/github.com/$REPO

RUN mkdir --parents $CLONE_PATH

RUN git clone -q https://github.com/$REPO $CLONE_PATH
RUN git checkout -q $BRANCH
RUN cd $CLONE_PATH/cmd/eris && go build -o $INSTALL_BASE/eris

USER $USER
WORKDIR $ERIS
VOLUME $ERIS
