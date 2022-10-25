package state

import (
	"testing"

	"github.com/eagraf/habitat/structs/community"
	"github.com/stretchr/testify/assert"
)

func TestJSONState(t *testing.T) {
	initState := []byte(`{
		"community_id": "abc"
	}`)
	jsonState, err := NewJSONState(community.CommunityStateSchema, initState)
	assert.Nil(t, err)

	patch := []byte(`[{
		"op": "add",
		"path": "/counter",
		"value": 1
	}]`)

	err = jsonState.ApplyPatch(patch)
	assert.Nil(t, err)
}
