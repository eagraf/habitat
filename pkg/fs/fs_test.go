package fs

import (
	"fmt"
	"testing"
)

func TestIPFSReadDir(t *testing.T) {
	ipfs, err := NewFS(IPFSType)
	if err != nil {
		t.Error(err)
	}

	res, err := ipfs.ReadDir("/")
	if err != nil {
		t.Error(err)
	}

	fmt.Println(res)
}
