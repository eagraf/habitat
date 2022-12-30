#!/bin/bash

set -e 

TESTDIR=$(realpath "$(dirname "$0")")
. "$TESTDIR/../../tools/lib.bash"
. "$TESTDIR/../setup-community.bash"

setup_community true

bob_and_charlie_join

testfile="$(temp::file)"
CONTENTS="ABC123"
echo $CONTENTS > tempff

HASH=`$ALICE_NODE_CLI_CMD fs add tempff | jq -r .content_id`
GETA=`$ALICE_NODE_CLI_CMD fs get $HASH`
GETB=`$BOB_NODE_CLI_CMD fs get $HASH`
GETC=`$CHARLIE_NODE_CLI_CMD fs get $HASH`

[[ "$CONTENTS" == "$GETA" ]] || log::fatal "Alice could not get the correct file back"
[[ "$CONTENTS" == "$GETB" ]] || log::fatal "Bob could not get the correct file back"
[[ "$CONTENTS" == "$GETC" ]] || log::fatal "Charlie could not get the correct file back"
