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

docker run --net=host --name=habitat_ipfs --mount type=bind,src=$IPFS_PATH,dst=/data/ipfs --mount type=bind,src=$SCRIPT_DIR/start_ipfs_with_config.sh,dst=/usr/local/bin/start.sh --entrypoint /usr/local/bin/start.sh -p 4001:4001 -p 5001:5001 -p 8080:8080 ipfs/go-ipfs  daemon --migrate=true
