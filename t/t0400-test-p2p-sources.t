#!/usr/bin/env bash

TESTDIR=$(realpath "$(dirname "$0")")
. "$TESTDIR/../tools/lib.bash"

sources-p2p() {
    testing::desc "read sources remotely over p2p"

    local ret
    { $TESTDIR/t0400/p2p-sources.bash ; ret=$? ; } || true

    return $ret
}

testing::register sources-p2p
testing::run
