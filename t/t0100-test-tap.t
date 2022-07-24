#!/usr/bin/env bash

TESTDIR=$(realpath "$(dirname "$0")")
. "$TESTDIR/../tools/lib.bash"

echo "WHUTT"
test-tap() {
    testing::desc verify that tap testing works
    return 0
}
echo "HELLLLOOOO"
testing::register test-tap
testing::run
