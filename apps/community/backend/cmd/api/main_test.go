package main

import (
	"fmt"
	"testing"
)

func TestCommunities(*testing.T) {
	res, err := CreateCommunity("testcomm", "ipfs1")
	fmt.Println(string(res), err)
}
