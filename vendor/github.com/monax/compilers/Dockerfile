FROM quay.io/monax/build:0.16
MAINTAINER Monax <support@monax.io>

# build customizations start here
ENV SOLC_VERSION 0.4.4
ENV JSONCPP_VERSION 1.7.7

# install build depenedencies
RUN apk --no-cache --update add build-base cmake boost-dev file

# stop boost complaining about sys/poll.h
RUN sed -i -E -e 's/include <sys\/poll.h>/include <poll.h>/' /usr/include/boost/asio/detail/socket_types.hpp

# get correct repositories
WORKDIR /src
RUN git clone https://github.com/open-source-parsers/jsoncpp
RUN git clone https://github.com/ethereum/solidity

# alpine has jsoncpp-dev, but it doesn't provide static libs
WORKDIR /src/jsoncpp
RUN git checkout $JSONCPP_VERSION \
  && cmake -DBUILD_STATIC_LIBS=ON -DBUILD_SHARED_LIBS=OFF . \
  && make jsoncpp_lib_static \
  && make install

# build solidity
WORKDIR /src/solidity/build
RUN git checkout v$SOLC_VERSION \
  && cmake -DCMAKE_BUILD_TYPE=Release \
          -DTESTS=1 \
          -DSTATIC_LINKING=1 \
          .. \
  && make --jobs=2 solc soltest \
  && install -s solc/solc /usr/local/bin \
  && install -s test/soltest /usr/local/bin
# build customizations end here

# Install monax-compilers, a go app that serves compilation results
ENV TARGET compilers
ENV REPO $GOPATH/src/github.com/monax/compilers

ADD ./glide.yaml $REPO/
ADD ./glide.lock $REPO/
WORKDIR $REPO
RUN glide install

COPY . $REPO/.
RUN cd $REPO/cmd/$TARGET && \
  go build --ldflags '-extldflags "-static"' -o $INSTALL_BASE/$TARGET
