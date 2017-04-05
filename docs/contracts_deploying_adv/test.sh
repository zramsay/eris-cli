#!/usr/bin/env bash
set -e
chains_dir=$HOME/.monax/chains
chain_name=advchain2
name_full="$chain_name"_full_000
name_part="$chain_name"_participant_000
chain_dir=$chains_dir/$chain_name
monax chains make --account-types=Full:1,Participant:1 $chain_name
key1_addr=$(cat $chain_dir/addresses.csv | grep $name_full | cut -d ',' -f 1)
monax chains start $chain_name --init-dir $chain_dir/$name_full 1>/dev/null

monax pkgs do --chain "$chain_name" --address "$key1_addr" --set "setStorageBase=5"

ls ./abi

monax clean --yes
