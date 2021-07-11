#!/bin/bash

export SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
export IPFS_PATH=$SCRIPT_DIR/ipfs
echo $IPFS_PATH

if [ -f "$IPFS_PATH" ]; then
    ipfs init
else 
    echo "ipfs already initialized"
fi

ipfs daemon
