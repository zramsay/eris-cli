FROM golang:1.4
MAINTAINER Eris Industries <support@erisindustries.com>

# User Creation
ENV user eris
RUN groupadd --system $user && useradd --system --create-home --gid $user $user
RUN mkdir --parents /home/$user/.eris
RUN chown --recursive $user /home/$user/.eris
# USER eris