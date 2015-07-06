FROM golang:1.4
MAINTAINER Eris Industries <support@erisindustries.com>

# User Creation
ENV USER eris

# Choose the user id number 1000 to work well with Boot2Docker volumes.  See:
# https://github.com/boot2docker/boot2docker/issues/581#issuecomment-62491280
RUN groupadd --system $USER \
  && useradd --system --create-home --uid 1000 --gid $USER $USER

RUN mkdir /home/$USER/.eris
RUN chown --recursive $USER /home/$USER/.eris
# USER eris
