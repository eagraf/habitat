#!/bin/bash

set -e

TESTDIR=$(realpath "$(dirname "$0")")
. "$TESTDIR/../../tools/lib.bash"
. "$TESTDIR/../setup-community.bash"

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

setup_community false

$ALICE_CMD community propose -c $COMMUNITY_UUID $(wrapTransition "initialize_counter" $TRANSITION1)

sleep 1
$BOB_CMD community join --token $COMMUNITY_JOIN_CODE
$CHARLIE_CMD community join --token $COMMUNITY_JOIN_CODE
sleep 3

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
