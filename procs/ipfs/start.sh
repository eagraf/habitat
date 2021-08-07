#!/bin/bash

export SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
export IPFS_PATH=$SCRIPT_DIR/ipfs

if [ ! -f "$IPFS_PATH" ]; then
    a="initializing IPFS in "
    echo $a$IPFS_PATH
    ipfs init
else 
    echo "ipfs already initialized"
fi

ipfs daemon
