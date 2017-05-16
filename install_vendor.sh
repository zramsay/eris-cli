#!/usr/bin/env bash
# This scripts installs a pruned vendor directory based of the vendor.conf lockfile
# list of paths to preserve from full vendor tree (even if they would be pruned)
preserve=(
github.com/ethereum/go-ethereum/crypto/secp256k1/libsecp256k1
github.com/stretchr/testify
)
# install trash
go get -u github.com/rancher/trash
# start fresh
rm -rf vendor
# trash will use this directory for its cache (although this is its default
# anyway), but we also need it for preserve action later so doing this makes
# sure we're in sync
export TRASH_CACHE="$HOME/.trash-cache"
# get everything at locked versions
trash
# Restore vendor paths from preserve
for p in "${preserve[@]}"
do
  echo "Restoring '$p' to vendor..."
  rel_path="./vendor/${p}"
  parent=`dirname "${rel_path}"`
  mkdir -p "${parent}"
  cp -r "${TRASH_CACHE}/src/${p}" "${parent}"
  rm -rf "$rel_path/.git"*
done
