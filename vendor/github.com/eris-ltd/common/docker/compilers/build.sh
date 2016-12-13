#!/usr/bin/env bash

start=`pwd`
name=compilers
eris_name=eris-$name
repo=https://github.com/eris-ltd/$eris_name

cd /tmp
git clone $repo $name
cd $name
tests/build_tool.sh

cd .. && rm -rf $name
cd $start