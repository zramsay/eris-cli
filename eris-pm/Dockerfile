FROM quay.io/eris/build
MAINTAINER Monax <support@monax.io>

# Install eris-pm, a go app that manages packages
ENV TARGET eris-pm
ENV REPO $GOPATH/src/github.com/eris-ltd/$TARGET

COPY ./glide.yaml $REPO/glide.yaml
COPY ./glide.lock $REPO/glide.lock
WORKDIR $REPO
RUN glide install

COPY . $REPO/.
RUN cd $REPO/cmd/$TARGET && \
  go build --ldflags '-extldflags "-static"' -o $INSTALL_BASE/$TARGET && \
  unset TARGET && \
  unset REPO
