#!/usr/bin/env bash
# assumes golang build from source or golang >= 1.5

go get github.com/laher/goxc
# read version from version file
goxc -wd cmd/eris -n eris -d $HOME/.eris/builds