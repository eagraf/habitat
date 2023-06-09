#!/bin/bash

set -e 

TESTDIR=$(realpath "$(dirname "$0")")
. "$TESTDIR/../../tools/lib.bash"
. "$TESTDIR/../setup-community.bash"

setup

#bob_and_charlie_join

#testfile="$(temp::file)"
#CONTENTS="ABC123"
#echo $CONTENTS > tempff

schema_hash=`$ALICE_NODE_CLI_CMD dex schema write '{"foo":"bar"}' | jq -r .hash`
schema_res=`$ALICE_NODE_CLI_CMD dex schema $schema_hash | jq -r .foo`
iface_hash=`$ALICE_NODE_CLI_CMD dex interface write '{"schema_hash":"'$schema_hash'","description":"desc"}' | jq -r .hash`
iface_res=`$ALICE_NODE_CLI_CMD dex interface $iface_hash | jq -r .description`


[[ "$schema_res" == "bar" ]] || log::fatal "Schema read/write failed"
[[ "$iface_res" == "desc" ]] || log::fatal "Interface read/write failed"