package ipfs

import (
	"os"
	"testing"
)

// These tests don't actually validate results, they just run the api calls

func TestVersion(t *testing.T) {
	var ipfsURL string
	if ipfsURL = os.Getenv("IPFS_URL"); ipfsURL == "" {
		t.Skip()
	}
	client, err := NewClient(ipfsURL)
	if err != nil {
		t.Error(err)
	}

	_, err = client.GetVersion()
	if err != nil {
		t.Error(err)
	}
}

func TestListFiles(t *testing.T) {
	var ipfsURL string
	if ipfsURL = os.Getenv("IPFS_URL"); ipfsURL == "" {
		t.Skip()
	}
	client, err := NewClient(ipfsURL)
	if err != nil {
		t.Error(err)
	}

	_, err = client.ListFiles()
	if err != nil {
		t.Error(err)
	}
}
