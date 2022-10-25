package state

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/eagraf/habitat/structs/community"
	"github.com/stretchr/testify/assert"
)

func testTransitions(transitions []CommunityStateTransition) (*community.CommunityState, error) {
	state := community.NewCommunityState()
	for _, t := range transitions {

		err := t.Validate(state)
		if err != nil {
			return nil, fmt.Errorf("transition validation failed: %s", err)
		}

		patch, err := t.Patch(state)
		if err != nil {
			return nil, err
		}

		marshaledState, err := json.Marshal(state)
		if err != nil {
			return nil, err
		}

		jsonState, err := NewJSONState(community.CommunityStateSchema, marshaledState)
		if err != nil {
			return nil, err
		}

		newStateBytes, err := jsonState.ValidatePatch(patch)
		if err != nil {
			return nil, err
		}

		var newState community.CommunityState
		err = json.Unmarshal(newStateBytes, &newState)
		if err != nil {
			return nil, err
		}

		state = &newState
	}

	return state, nil
}

func TestCommunityInitialization(t *testing.T) {
	transitions := []CommunityStateTransition{
		&InitializeCommunityTransition{
			CommunityID: "abc",
		},
		&AddMemberTransition{
			Member: &community.Member{
				ID:          "jorts",
				Certificate: []byte("mycert"),
			},
		},
		&AddNodeTransition{
			Node: &community.Node{
				ID:          "george",
				MemberID:    "jorts",
				Certificate: []byte("mycert"),
			},
		},
	}

	state, err := testTransitions(transitions)
	assert.Nil(t, err)
	assert.NotNil(t, state)
	assert.Equal(t, "abc", state.CommunityID)
	assert.Equal(t, 1, len(state.Members))
	assert.Equal(t, 1, len(state.Nodes))

	// test adding multiple members with the same id

	transitions = []CommunityStateTransition{
		&InitializeCommunityTransition{
			CommunityID: "abc",
		},
		&AddMemberTransition{
			Member: &community.Member{
				ID: "jorts",
			},
		},
		&AddMemberTransition{
			Member: &community.Member{
				ID: "jorts",
			},
		},
	}

	_, err = testTransitions(transitions)
	assert.NotNil(t, err)

	// test adding a node not associated with a current member

	transitions = []CommunityStateTransition{
		&InitializeCommunityTransition{
			CommunityID: "abc",
		},
		&AddNodeTransition{
			Node: &community.Node{
				ID:       "george",
				MemberID: "jorts",
			},
		},
		&AddMemberTransition{
			Member: &community.Member{
				ID: "jorts",
			},
		},
	}

	_, err = testTransitions(transitions)
	assert.NotNil(t, err)
}
