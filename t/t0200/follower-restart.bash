#!/bin/bash

set -e

TESTDIR=$(realpath "$(dirname "$0")")
. "$TESTDIR/../../tools/lib.bash"

function wrapTransition() {
base64 -w 0 <<EOF
[{
    "type": "$1",
    "patch": "$2"
}]
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
atexit "docker-compose -f docker-compose-raft.yml down"
atexit "docker-compose rm -f"

sleep 10 

ID_PATH="$(temp::dir)"
export HABITATCTL_IDENTITY_PATH=$ID_PATH

$HABITATCTL_PATH id init
ALICE_NODE_CLI_CMD="$HABITATCTL_PATH -p 2000"
BOB_NODE_CLI_CMD="$HABITATCTL_PATH -p 2001"
CHARLIE_NODE_CLI_CMD="$HABITATCTL_PATH -p 2002"

ALICE_NODE_ID=`$ALICE_NODE_CLI_CMD community ls | head -n1 | awk '{print $3}'`
BOB_NODE_ID=`$BOB_NODE_CLI_CMD community ls | head -n1 | awk '{print $3}'`
CHARLIE_NODE_ID=`$CHARLIE_NODE_CLI_CMD community ls | head -n1 | awk '{print $3}'`


$ALICE_NODE_CLI_CMD id create --username alice --password abc
$BOB_NODE_CLI_CMD id create --username bob --password abc
$CHARLIE_NODE_CLI_CMD id create --username charlie --password abc

ALICE_CMD="$ALICE_NODE_CLI_CMD --username alice --password abc"
BOB_CMD="$BOB_NODE_CLI_CMD --username bob --password abc"
CHARLIE_CMD="$CHARLIE_NODE_CLI_CMD --username charlie --password charlie"

COMMUNITY_UUID=`$ALICE_CMD community create | head -n1 | awk '{print $1}'`

sleep 2

$ALICE_CMD community propose -c $COMMUNITY_UUID $(wrapTransition "initialize_counter" $TRANSITION1)

ALICE_IP=`docker inspect -f '{{range.NetworkSettings.Networks}}{{.IPAddress}}{{end}}' habitat_alice_1`
BOB_IP=`docker inspect -f '{{range.NetworkSettings.Networks}}{{.IPAddress}}{{end}}' habitat_bob_1`
CHARLIE_IP=`docker inspect -f '{{range.NetworkSettings.Networks}}{{.IPAddress}}{{end}}' habitat_charlie_1`

sleep 1
$BOB_CMD community join -c $COMMUNITY_UUID -a /ip4/$ALICE_IP/tcp/6000
$CHARLIE_CMD community join -c $COMMUNITY_UUID -a /ip4/$ALICE_IP/tcp/6000
sleep 3
$ALICE_CMD community add -c $COMMUNITY_UUID -n $BOB_NODE_ID -a /ip4/$BOB_IP/tcp/6000
$ALICE_CMD community add -c $COMMUNITY_UUID -n $CHARLIE_NODE_ID -a /ip4/$CHARLIE_IP/tcp/6000

sleep 1

docker restart habitat_bob_1

$ALICE_CMD community propose -c $COMMUNITY_UUID $(wrapTransition "increment_counter" $TRANSITION2)

sleep 6

COUNTER1=`$ALICE_CMD community state -c $COMMUNITY_UUID | jq .counter`
COUNTER2=`$BOB_CMD community state -c $COMMUNITY_UUID | jq .counter`
COUNTER3=`$CHARLIE_CMD community state -c $COMMUNITY_UUID | jq .counter`

sleep 1

[[ $COUNTER1 -eq 2 ]] || log::fatal "alice's counter should be 2, not $COUNTER1"
[[ $COUNTER2 -eq 2 ]] || log::fatal "bob's counter should be 2, not $COUNTER2"
[[ $COUNTER3 -eq 2 ]] || log::fatal "charlie's counter should be 2, not $COUNTER3"
