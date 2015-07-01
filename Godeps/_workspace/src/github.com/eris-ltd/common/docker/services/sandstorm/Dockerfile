FROM eris/base
MAINTAINER Eris Industries <support@erisindustries.com>

RUN apt-get update
RUN apt-get install -y curl xz-utils

COPY install.sh /install.sh
RUN chown $USER:$USER /install.sh

USER $USER

RUN /install.sh -d -u

RUN echo 'SERVER_USER=eris\n\
PORT=6080\n\
MONGO_PORT=6081\n\
BIND_IP=0.0.0.0\n\
BASE_URL=http://localhost:6080\n\
WILDCARD_HOST=*.local.sandstorm.io:6080\n\
MAIL_URL=\n' > $HOME/sandstorm/sandstorm.conf

RUN echo 'export PATH=$PATH:$HOME/sandstorm' >> $HOME/.bashrc
RUN ln -s /home/eris/sandstorm/var /home/eris/.eris/sandstorm

VOLUME $HOME/.eris/data/sandstorm

EXPOSE 6080
CMD $HOME/sandstorm/sandstorm start && sleep infinity