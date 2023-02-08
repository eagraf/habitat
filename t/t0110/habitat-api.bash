#!/usr/bin/env bash

set -e

TESTDIR=$(realpath "$(dirname "$0")")
. "$TESTDIR/../../tools/lib.bash"

docker-compose up -V -- alice 2> /dev/null &
atexit "docker-compose down"
atexit "docker-compose rm -f"

sleep 5

curl -f http://localhost:2000/ps || log::fatal "failed to reach Habitat API over HTTP"
curl -f http://localhost:3000/habitat/ps || log::fatal "failed to reach Habitat API via reverse proxy over HTTP"

LIBP2P_PROXY_ADDR=`$HABITATCTL_PATH -p 2000 inspect | jq -r .libp2p_proxy_multiaddr`

$HABITATCTL_PATH -p 2000 --libp2p-proxy $LIBP2P_PROXY_ADDR inspect || log::fatal "failed to reach Habitat API via LibP2P reverse proxy"
