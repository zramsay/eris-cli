#!/usr/bin/env bash
# [Silas] glide-vc didn't work for me (just failed exit status 1, no useful error message)
# so I have experimented with Govend
go get -u github.com/govend/govend
# govend has different behaviour (seems to skip things if vendor is there
rm -r vendor
# get everything at locked versions (so we can grab libsecp256k1 at the locked version)
govend -v -r
cp -r vendor/github.com/ethereum/go-ethereum/crypto/secp256k1/libsecp256k1 .
# Same again but pruned
govend -v -r --prune
# Restore lib
cp -r ./libsecp256k1 vendor/github.com/ethereum/go-ethereum/crypto/secp256k1/
rm -rf libsecp256k1