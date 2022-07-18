package fs

import (
	"fmt"
	"os"
	"testing"
)

func TestIPFSReadDir(t *testing.T) {
	var ipfsURL string
	if ipfsURL = os.Getenv("IPFS_URL"); ipfsURL == "" {
		t.Skip()
	}
	ipfs, err := NewFS(IPFSType, ipfsURL)
	if err != nil {
		t.Error(err)
	}

	res, err := ipfs.ReadDir("/")
	if err != nil {
		t.Error(err)
	}

	fmt.Println(res)
}
