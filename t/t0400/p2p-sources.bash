#!/bin/bash

set -e 

TESTDIR=$(realpath "$(dirname "$0")")
. "$TESTDIR/../../tools/lib.bash"
. "$TESTDIR/../setup-community.bash"

setup_community true
bob_join

geo_id='https://json-schema.org/learn/examples/geographical-location.schema.json'
alice_data='{"latitude":48,"longitude":98}'
bob_data='{"latitude":78,"longitude":99}'

$ALICE_NODE_CLI_CMD sources write --id $geo_id --data $alice_data
ALICE_READ=`$ALICE_NODE_CLI_CMD sources read --id $geo_id`
BOB_READ=`$BOB_NODE_CLI_CMD sources read --id $geo_id --node $ALICE_NODE_ID`

[[ "$ALICE_READ" == "$alice_data" ]] || log::fatal "Alice sources read did not match the write $ALICE_READ $alice_data"
[[ "$BOB_READ" == "$alice_data" ]] || log::fatal "Bob sources remote read did not match Alice sources write $ALICE_READ $alice_data"

$BOB_NODE_CLI_CMD sources write --id $geo_id --data $bob_data
BOB_READ=`$BOB_NODE_CLI_CMD sources read --id $geo_id`
ALICE_READ=`$ALICE_NODE_CLI_CMD sources read --id $geo_id --node $BOB_NODE_ID`

[[ "$BOB_READ" == "$bob_data" ]] || log::fatal "Bob sources read did not match the write"
[[ "$ALICE_READ" == "$bob_data" ]] || log::fatal "Alice sources remote read did not match Bob sources write"
