#!/bin/bash
# order of arguments is ipfs path, ip, port, peer id of new peer to add, swarm key
export SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
export INPUT_PATH=$SCRIPT_DIR/$1

if [ ! -e "$INPUT_PATH" ]; then
    echo "no ipfs node at this path, exiting"
else 
    echo "ipfs initialized at $INPUT_PATH"
    echo "/key/swarm/psk/1.0.0/" >> $INPUT_PATH/swarm.key
    echo "/base16/" >> $INPUT_PATH/swarm.key
    echo $5 >> $INPUT_PATH/swarm.key
    IPFS_PATH=$INPUT_PATH ipfs bootstrap add /ip4/$2/tcp/$3/ipfs/$4
fi