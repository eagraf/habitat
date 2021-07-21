#!/bin/bash
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
docker run -v $SCRIPT_DIR/bin:/root/bin --net=host --expose=8000 debian:bullseye-slim /root/bin/notes-api
