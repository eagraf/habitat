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

$ALICE_NODE_CLI_CMD start dex-ipfs server

schema_hash=`docker exec -it habitat_alice_1 /habitat/apps/dex-ipfs/bin/amd64-linux/dex-ipfs client schema write '{"foo":"bar"}' | jq -r .hash`
schema_res=`docker exec -it habitat_alice_1 /habitat/apps/dex-ipfs/bin/amd64-linux/dex-ipfs client schema $schema_hash | jq -r .foo`
iface_hash=`docker exec -it habitat_alice_1 /habitat/apps/dex-ipfs/bin/amd64-linux/dex-ipfs client interface write '{"schema_hash":"'$schema_hash'","description":"desc"}' | jq -r .hash`
iface_res=`docker exec -it habitat_alice_1 /habitat/apps/dex-ipfs/bin/amd64-linux/dex-ipfs client interface $iface_hash | jq -r .description`


[[ "$schema_res" == "bar" ]] || log::fatal "Schema read/write failed"
[[ "$iface_res" == "desc" ]] || log::fatal "Interface read/write failed"
#[[ "$CONTENTS" == "$GETC" ]] || log::fatal "Charlie could not get the correct file back"
