FROM quay.io/monax/base:0.16
MAINTAINER Monax | Platform | Tools <support@monax.io>

# Install go, git, and glide
RUN apk --no-cache --update add go git
ENV GOPATH /go
ENV PATH $GOPATH/bin:/usr/local/go/bin:$PATH
RUN mkdir -p "$GOPATH/src" "$GOPATH/bin" && chmod -R 777 "$GOPATH"

# Install jq and c libraries
RUN apk --no-cache --update add jq gmp-dev gcc musl-dev
RUN go get github.com/Masterminds/glide && \
    go get github.com/sgotti/glide-vc