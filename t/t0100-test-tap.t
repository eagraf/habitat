#!/usr/bin/env bash

TESTDIR=$(realpath "$(dirname "$0")")
. "$TESTDIR/../tools/lib.bash"

test-tap() {
    testing::desc verify that tap testing works
    return 0
}

testing::register test-tap
testing::run
