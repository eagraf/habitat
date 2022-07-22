#!/bin/bash

set -e

TESTDIR=$(realpath "$(dirname "$0")")
. "$TESTDIR/../../tools/lib.bash"

function wrapTransition() {
base64 -w 0 <<EOF
{
    "type": "$1",
    "patch": "$2"
}
EOF
}

TRANSITION1=$(base64 -w 0 <<EOF
[{
    "op": "add",
    "path": "/counter",
    "value": 1
}]
EOF
)

TRANSITION2=$(base64 -w 0 <<EOF
[{
    "op": "replace",
    "path": "/counter",
    "value": 2
}]
EOF
)

TRANSITION3=$(base64 -w 0 <<EOF
[{
    "op": "replace",
    "path": "/counter",
    "value": 3
}]
EOF
)
docker-compose -f docker-compose-raft.yml up -V 2> /dev/null &

echo $(wrapTransition "initialize_counter" $TRANSITION1)

sleep 10

ALICE_NODE_ID=`./bin/habitatctl -p 2000 community ls | head -n1 | awk '{print $3}'`
BOB_NODE_ID=`./bin/habitatctl -p 2001 community ls | head -n1 | awk '{print $3}'`
CHARLIE_NODE_ID=`./bin/habitatctl -p 2002 community ls | head -n1 | awk '{print $3}'`

COMMUNITY_UUID=`./bin/habitatctl -p 2000 community create | head -n1 | awk '{print $1}'`

sleep 2

./bin/habitatctl -p 2000 community propose -c $COMMUNITY_UUID $(wrapTransition "initialize_counter" $TRANSITION1)
./bin/habitatctl -p 2000 community propose -c $COMMUNITY_UUID $(wrapTransition "increment_counter" $TRANSITION2)

ALICE_IP=`docker inspect -f '{{range.NetworkSettings.Networks}}{{.IPAddress}}{{end}}' habitat_alice_1`
BOB_IP=`docker inspect -f '{{range.NetworkSettings.Networks}}{{.IPAddress}}{{end}}' habitat_bob_1`
CHARLIE_IP=`docker inspect -f '{{range.NetworkSettings.Networks}}{{.IPAddress}}{{end}}' habitat_charlie_1`

./bin/habitatctl -p 2001 community join -c $COMMUNITY_UUID -a /ip4/$ALICE_IP/tcp/6000
./bin/habitatctl -p 2002 community join -c $COMMUNITY_UUID -a /ip4/$ALICE_IP/tcp/6000
sleep 3
./bin/habitatctl -p 2000 community add -c $COMMUNITY_UUID -n $BOB_NODE_ID -a /ip4/$BOB_IP/tcp/6000
./bin/habitatctl -p 2000 community add -c $COMMUNITY_UUID -n $CHARLIE_NODE_ID -a /ip4/$CHARLIE_IP/tcp/6000

#sleep 1
./bin/habitatctl -p 2000 community propose -c $COMMUNITY_UUID $(wrapTransition "increment_counter" $TRANSITION3)

sleep 1

COUNTER1=`./bin/habitatctl -p 2000 community state -c $COMMUNITY_UUID | jq .counter`
COUNTER2=`./bin/habitatctl -p 2001 community state -c $COMMUNITY_UUID | jq .counter`
COUNTER3=`./bin/habitatctl -p 2002 community state -c $COMMUNITY_UUID | jq .counter`

docker-compose -f docker-compose-raft.yml down 2> /dev/null
docker-compose rm -f 2> /dev/null

[[ $COUNTER1 -eq 3 ]] || log::fatal "alice's counter should be 3, not $COUNTER1"
[[ $COUNTER2 -eq 3 ]] || log::fatal "bob's counter should be 3, not $COUNTER2"
[[ $COUNTER3 -eq 3 ]] || log::fatal "charlie's counter should be 3, not $COUNTER3"
