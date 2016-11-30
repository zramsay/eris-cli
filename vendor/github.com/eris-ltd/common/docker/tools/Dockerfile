FROM quay.io/eris/ubuntu
MAINTAINER Eris Industries <support@erisindustries.com>

RUN apt-get install -qq -y software-properties-common && \
  add-apt-repository -y ppa:ethereum/ethereum > /dev/null && \
  add-apt-repository -y ppa:ethereum/ethereum-dev > /dev/null && \
  apt-get update -qq && \
  apt-get install -qq -y --no-install-recommends \
    build-essential \
    curl \
    gcc \
    g++ \
    git \
    make \
    libyaml-dev \
    python-dev \
    lllc \
    jq \
    sc \
    solc && \
  curl -sL https://bootstrap.pypa.io/get-pip.py | python - && \
  rm -rf /var/lib/apt/lists/*
