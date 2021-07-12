#!/bin/bash

export SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
export CONTENT_DIR=$SCRIPT_DIR/content

docker run -d -p 8080:80 -v $CONTENT_DIR:/usr/share/nginx/html --name nginx nginx
