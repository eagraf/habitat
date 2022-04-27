package state

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/qri-io/jsonschema"
	"github.com/stretchr/testify/assert"
)

func TestCommunityStateSchema(t *testing.T) {
	rs := &jsonschema.Schema{}
	err := json.Unmarshal(communityStateSchema, rs)
	assert.Nil(t, err)

	// Test that an empty CommunityState struct matches the schema
	cs := CommunityState{}
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
	jsonState, err := NewJSONState(communityStateSchema, initState)
	assert.Nil(t, err)

	patch := []byte(`[{
		"op": "add",
		"path": "/counter",
		"value": 1
	}]`)

	err = jsonState.ApplyPatch(patch)
	assert.Nil(t, err)
}

func TestCounterIncrement(t *testing.T) {
	jsonState, err := NewJSONState(communityStateSchema, []byte(`{"community_id":"abc"}`))
	assert.Nil(t, err)

	csm := NewCommunityStateMachine(jsonState, &LocalDispatcher{jsonState: jsonState})
	state, err := csm.State()
	assert.Nil(t, err)
	assert.Equal(t, state.CommunityID, "abc")

	initCounter := &InitializeCounterTransition{
		InitialCount: 1,
	}
	err = csm.ProposeTransition(initCounter)
	assert.Nil(t, err)

	state, err = csm.State()
	assert.Nil(t, err)
	assert.Equal(t, state.Counter, 1)

	incCounter := &IncrementCounterTransition{}
	err = csm.ProposeTransition(incCounter)
	assert.Nil(t, err)

	state, err = csm.State()
	assert.Nil(t, err)
	assert.Equal(t, state.Counter, 2)
}
