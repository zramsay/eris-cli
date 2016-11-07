FROM quay.io/eris/build
MAINTAINER Monax <support@monax.io>

# Install eris-pm, a go app that manages packages
ENV NAME eris-pm
ENV REPO $GOPATH/src/github.com/eris-ltd/$NAME

COPY ./glide.yaml $REPO/glide.yaml
COPY ./glide.lock $REPO/glide.lock
WORKDIR $REPO
RUN glide install

COPY . $REPO/.
RUN cd $REPO/cmd/$NAME && \
  go build --ldflags '-extldflags "-static"' -o $INSTALL_BASE/$NAME && \
  unset NAME && \
  unset REPO
