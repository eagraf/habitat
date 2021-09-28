package main

import (
	"fmt"
	"testing"
)

func TestCommunities(*testing.T) {
	err, key, id1, addrs := CreateCommunity("testcomm", "ipfs1")
	fmt.Print("Error: ", err, "\nkey: ", key, "\nid1: ", id1, "\naddrs: ", addrs, "\n")

	s := make([]string, 0)
	err, id2 := JoinCommunity("testcomm", "ipfs2", key, addrs, s)
	fmt.Print("Error: ", err, "\nid2: ", id2, "\n")

}
