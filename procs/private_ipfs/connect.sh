export SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
export INPUT_PATH=$SCRIPT_DIR/$1

IPFS_PATH=$INPUT_PATH ipfs daemon