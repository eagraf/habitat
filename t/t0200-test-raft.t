#!/usr/bin/env bash

TESTDIR=$(realpath "$(dirname "$0")")
. "$TESTDIR/../tools/lib.bash"

raft-simple() {
    testing::desc "test raft consensus protocol"
    echo "HOLA"
    local ret
    { $TESTDIR/t0200/raft-simple.bash ; ret=$? ; } || true

    echo "HEHE"

    return $ret
}

raft-follower-restart() {
    testing::desc "test a raft cluser where a node fails and restarts"
    echo "HEHE"

    local ret
    { $TESTDIR/t0200/follower-restart.bash ; ret=$? ; } || true

    return $ret
}

testing::register raft-simple
testing::register raft-follower-restart
testing::run
