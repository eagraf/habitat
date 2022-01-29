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
docker-compose up 2> /dev/null &

sleep 2

ALICE_NODE_ID=`./bin/habitatctl -p 2000 community ls | awk '{print $3}'`
BOB_NODE_ID=`./bin/habitatctl -p 2001 community ls | awk '{print $3}'`
CHARLIE_NODE_ID=`./bin/habitatctl -p 2002 community ls | awk '{print $3}'`

COMMUNITY_UUID=`./bin/habitatctl -p 2000 community create | awk '{print $NF}'`

sleep 1

./bin/habitatctl -p 2000 community propose -c $COMMUNITY_UUID $TRANSITION1

ALICE_CONTAINER_ID=`docker ps | grep 'habitat_alice_1' | awk '{print $1}'`
BOB_CONTAINER_ID=`docker ps | grep 'habitat_bob_1' | awk '{print $1}'`
CHARLIE_CONTAINER_ID=`docker ps | grep 'habitat_charlie_1' | awk '{print $1}'`

sleep 1
./bin/habitatctl -p 2000 community add -c $COMMUNITY_UUID -n $BOB_NODE_ID -a http://$BOB_CONTAINER_ID:2041/raft/msg/$COMMUNITY_UUID
./bin/habitatctl -p 2000 community add -c $COMMUNITY_UUID -n $CHARLIE_NODE_ID -a http://$CHARLIE_CONTAINER_ID:2041/raft/msg/$COMMUNITY_UUID
./bin/habitatctl -p 2001 community join -c $COMMUNITY_UUID -a http://$ALICE_CONTAINER_ID:2041/raft/rpc/add
./bin/habitatctl -p 2002 community join -c $COMMUNITY_UUID -a http://$ALICE_CONTAINER_ID:2041/raft/rpc/add

sleep 1

docker restart $BOB_CONTAINER_ID
./bin/habitatctl -p 2000 community propose -c $COMMUNITY_UUID $TRANSITION2

sleep 1

#./bin/habitatctl -p 2001 start raft

sleep 3

COUNTER1=`./bin/habitatctl -p 2000 community state -c $COMMUNITY_UUID | jq .counter`
COUNTER2=`./bin/habitatctl -p 2001 community state -c $COMMUNITY_UUID | jq .counter`
COUNTER3=`./bin/habitatctl -p 2002 community state -c $COMMUNITY_UUID | jq .counter`

docker-compose down 2> /dev/null
docker-compose rm 2> /dev/null

[[ $COUNTER1 -eq 2 ]] || log::fatal "alice's counter should be 2, not $COUNTER1"
[[ $COUNTER2 -eq 2 ]] || log::fatal "bob's counter should be 2, not $COUNTER2"
[[ $COUNTER3 -eq 2 ]] || log::fatal "charlie's counter should be 2, not $COUNTER3"
