FROM quay.io/eris/tools
MAINTAINER Eris Industries <support@erisindustries.com>

# export rust LIBRARY_PATH
ENV LIBRARY_PATH /usr/local/lib

# show backtraces
ENV RUST_BACKTRACE 1

RUN curl -sf https://raw.githubusercontent.com/brson/multirust/master/quick-install.sh | sh -s -- --yes
RUN git clone https://github.com/ethcore/parity /tmp/parity && \
  cd /tmp/parity && \
  cargo build --release && \
  cargo install parity --path . --root /usr/local && \
  rm -rf /tmp/parity
RUN curl -sf https://raw.githubusercontent.com/brson/multirust/master/blastoff.sh | sh -s -- --uninstall --yes

# set user
USER $USER
WORKDIR $ERIS

# boot
VOLUME $ERIS
CMD parity