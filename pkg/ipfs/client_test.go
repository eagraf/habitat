package ipfs

import (
	"fmt"
	"testing"
)

// These tests don't actually validate results, they just run the api calls

func TestVersion(t *testing.T) {
	client, err := NewClient("http://localhost:5001/api/v0")
	if err != nil {
		t.Error(err)
	}

	version, err := client.GetVersion()
	if err != nil {
		t.Error(err)
	}

	fmt.Println(version)
}
