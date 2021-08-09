#!/bin/bash

export SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
export CONTENT_DIR=$SCRIPT_DIR/content

docker stop habitat_nginx
docker rm habitat_nginx
docker run --rm -d -p 8081:80 -v $CONTENT_DIR:/usr/share/nginx/html --name habitat_nginx nginx
