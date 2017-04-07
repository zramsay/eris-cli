#!/usr/bin/env bash

cp -r vendor/github.com/ethereum/go-ethereum/crypto/secp256k1/libsecp256k1 .
go get -u github.com/sgotti/glide-vc
glide vc
cp -r ./libsecp256k1 vendor/github.com/ethereum/go-ethereum/crypto/secp256k1/
rm -rf libsecp256k1