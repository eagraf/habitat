export SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
export INPUT_PATH=$1

echo "connecting to community $INPUT_PATH"
IPFS_PATH=$INPUT_PATH ipfs daemon