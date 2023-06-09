#!/bin/bash


setup () {
    docker-compose up -V 2> /dev/null &
    atexit "docker-compose down"
    atexit "docker-compose rm -f"

    sleep 10

    ID_PATH="$(temp::dir)"
    export HABITATCTL_IDENTITY_PATH=$ID_PATH

    $HABITATCTL_PATH id init
    ALICE_NODE_CLI_CMD="$HABITATCTL_PATH -p 2000"
    BOB_NODE_CLI_CMD="$HABITATCTL_PATH -p 2001"
    CHARLIE_NODE_CLI_CMD="$HABITATCTL_PATH -p 2002"

    ALICE_NODE_ID=`$ALICE_NODE_CLI_CMD community ls | head -n1 | awk '{print $3}'`
    BOB_NODE_ID=`$BOB_NODE_CLI_CMD community ls | head -n1 | awk '{print $3}'`
    CHARLIE_NODE_ID=`$CHARLIE_NODE_CLI_CMD community ls | head -n1 | awk '{print $3}'`
}

setup_community() {
    $ALICE_NODE_CLI_CMD id create --username alice --password abc
    $BOB_NODE_CLI_CMD id create --username bob --password abc
    $CHARLIE_NODE_CLI_CMD id create --username charlie --password abc

    ALICE_CMD="$ALICE_NODE_CLI_CMD --username alice --password abc"
    BOB_CMD="$BOB_NODE_CLI_CMD --username bob --password abc"
    CHARLIE_CMD="$CHARLIE_NODE_CLI_CMD --username charlie --password abc"

    local community_create_command="$ALICE_CMD community create"
    if $1 ; then
        community_create_command+=" --ipfs" 
    fi
    echo $community_create_command

    COMMUNITY_CREATE_RES=`$community_create_command` || true
    echo $COMMUNITY_CREATE_RES

    COMMUNITY_UUID=`echo $COMMUNITY_CREATE_RES | jq -r .community_id`
    COMMUNITY_JOIN_CODE=`echo $COMMUNITY_CREATE_RES | jq -r .join_code`

    sleep 2
}

bob_join() {
    $BOB_CMD community join --token $COMMUNITY_JOIN_CODE
}

bob_and_charlie_join() {
    $BOB_CMD community join --token $COMMUNITY_JOIN_CODE
    $CHARLIE_CMD community join --token $COMMUNITY_JOIN_CODE
}