FROM quay.io/eris/build
MAINTAINER Monax <support@monax.io>

# Install eris-db, a go app that manages relationships
ENV TARGET eris-db
ENV REPO $GOPATH/src/github.com/eris-ltd/$TARGET

ADD ./glide.yaml $REPO/
ADD ./glide.lock $REPO/
WORKDIR $REPO
RUN glide install

COPY . $REPO/.
RUN cd $REPO/cmd/$TARGET && \
  go build --ldflags '-extldflags "-static"' -o $INSTALL_BASE/$TARGET

# build customizations start here
RUN cd $REPO/client/cmd/eris-client && \
  go build --ldflags '-extldflags "-static"' -o $INSTALL_BASE/eris-client
