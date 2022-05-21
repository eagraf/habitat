package state

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/eagraf/habitat/structs/community"
	"github.com/qri-io/jsonschema"
	"github.com/stretchr/testify/assert"
)

func TestCommunityStateSchema(t *testing.T) {
	rs := &jsonschema.Schema{}
	err := json.Unmarshal(community.CommunityStateSchema, rs)
	assert.Nil(t, err)

	// Test that an empty CommunityState struct matches the schema
	cs := community.CommunityState{}
	marshaled, err := json.Marshal(&cs)
	assert.Nil(t, err)
	keyErrs, err := rs.ValidateBytes(context.Background(), marshaled)
	assert.Nil(t, err)
	if len(keyErrs) != 0 {
		for _, e := range keyErrs {
			t.Log(e)
		}
		t.Error()
	}
}

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
