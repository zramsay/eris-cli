FROM quay.io/eris/tools
MAINTAINER Eris Industries <support@erisindustries.com>

RUN echo "setting up node, npm, and io.js" && \
  curl -sL https://deb.nodesource.com/setup_4.x | bash - > /dev/null && \
  apt-get install -y -qq --no-install-recommends \
    nodejs && \
  rm -rf /var/lib/apt/lists/*
