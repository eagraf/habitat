#!/bin/bash

set -e

TESTDIR=$(realpath "$(dirname "$0")")
. "$TESTDIR/../../tools/lib.bash"

TRANSITION1=$(base64 <<EOF
[{
    "op": "add",
    "path": "/counter",
    "value": 1
}]
EOF
)

TRANSITION2=$(base64 <<EOF
[{
    "op": "replace",
    "path": "/counter",
    "value": 2
}]
EOF
)

TRANSITION3=$(base64 <<EOF
[{
    "op": "replace",
    "path": "/counter",
    "value": 3
}]
EOF
)
docker-compose -f docker-compose-raft.yml up -V 2> /dev/null &

sleep 10 

ALICE_NODE_ID=`./bin/habitatctl -p 2000 community ls | awk '{print $3}'`
BOB_NODE_ID=`./bin/habitatctl -p 2001 community ls | awk '{print $3}'`
CHARLIE_NODE_ID=`./bin/habitatctl -p 2002 community ls | awk '{print $3}'`

COMMUNITY_UUID=`./bin/habitatctl -p 2000 community create | awk '{print $NF}'`

sleep 2

./bin/habitatctl -p 2000 community propose -c $COMMUNITY_UUID $TRANSITION1

ALICE_IP=`docker inspect -f '{{range.NetworkSettings.Networks}}{{.IPAddress}}{{end}}' habitat_alice_1`
BOB_IP=`docker inspect -f '{{range.NetworkSettings.Networks}}{{.IPAddress}}{{end}}' habitat_bob_1`
CHARLIE_IP=`docker inspect -f '{{range.NetworkSettings.Networks}}{{.IPAddress}}{{end}}' habitat_charlie_1`

sleep 1
./bin/habitatctl -p 2001 community join -c $COMMUNITY_UUID -a /ip4/$ALICE_IP/tcp/6000
./bin/habitatctl -p 2002 community join -c $COMMUNITY_UUID -a /ip4/$ALICE_IP/tcp/6000
sleep 3
./bin/habitatctl -p 2000 community add -c $COMMUNITY_UUID -n $BOB_NODE_ID -a /ip4/$BOB_IP/tcp/6000
./bin/habitatctl -p 2000 community add -c $COMMUNITY_UUID -n $CHARLIE_NODE_ID -a /ip4/$CHARLIE_IP/tcp/6000

sleep 1

docker restart habitat_bob_1

./bin/habitatctl -p 2000 community propose -c $COMMUNITY_UUID $TRANSITION2

sleep 6

COUNTER1=`./bin/habitatctl -p 2000 community state -c $COMMUNITY_UUID | jq .counter`
COUNTER2=`./bin/habitatctl -p 2001 community state -c $COMMUNITY_UUID | jq .counter`
COUNTER3=`./bin/habitatctl -p 2002 community state -c $COMMUNITY_UUID | jq .counter`

docker-compose -f docker-compose-raft.yml down 2> /dev/null
docker-compose rm 2> /dev/null

sleep 1

[[ $COUNTER1 -eq 2 ]] || log::fatal "alice's counter should be 2, not $COUNTER1"
[[ $COUNTER2 -eq 2 ]] || log::fatal "bob's counter should be 2, not $COUNTER2"
[[ $COUNTER3 -eq 2 ]] || log::fatal "charlie's counter should be 2, not $COUNTER3"
