#!/bin/bash
# t is for TCP port, g is for Gateway port, s is for Swarm port
# TODO later (not obvious how to change the ports programmatically on start up)
while getopts t:g:s flag
do
    case "${flag}" in
        t) tcp=${OPTARG};;
        g) gate=${OPTARG};;
        s) swarm=${OPTARG};;
    esac
done

export SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
export INPUT_PATH=$1
export EDITOR=/usr/bin/vim

if [ ! -e $INPUT_PATH ]; then
    echo "initializing IPFS at $INPUT_PATH"
    IPFS_PATH=$INPUT_PATH ipfs init
    IPFS_PATH=$INPUT_PATH ipfs bootstrap rm --all
    # echo "check to make sure the Bootstrap field is null:"
    IPFS_PATH=$INPUT_PATH ipfs config profile apply randomports
    # IPFS_PATH=$INPUT_PATH ipfs config show
    # IPFS_PATH=$INPUT_PATH ipfs config edit
else 
    echo "ipfs already initialized at $INPUT_PATH, showing config"
    # IPFS_PATH=$INPUT_PATH ipfs config show
fi