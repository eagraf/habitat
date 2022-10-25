package community

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/qri-io/jsonschema"
	"github.com/stretchr/testify/assert"
)

func TestCommunityStateSchema(t *testing.T) {
	rs := &jsonschema.Schema{}
	err := json.Unmarshal(CommunityStateSchema, rs)
	assert.Nil(t, err)

	// Test that an empty CommunityState struct matches the schema
	cs := CommunityState{
		Members: []*Member{},
		Nodes:   []*Node{},
	}
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
