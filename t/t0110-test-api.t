#!/usr/bin/env bash

TESTDIR=$(realpath "$(dirname "$0")")
. "$TESTDIR/../tools/lib.bash"

habitat-api() {
    testing::desc "reach Habitat API via various methods"

    local ret
    { $TESTDIR/t0110/habitat-api.bash ; ret=$? ; } || true

    return $ret
}

sources-p2p() {
    testing::desc "read sources remotely over p2p"

    local ret
    { $TESTDIR/t0400/p2p-sources.bash ; ret=$? ; } || true

    return $ret
}

testing::register habitat-api
testing::register sources-p2p
testing::run
