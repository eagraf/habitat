package state

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/eagraf/habitat/structs/community"
	"github.com/stretchr/testify/assert"
)

func testTransitions(oldState *community.CommunityState, transitions []CommunityStateTransition) (*community.CommunityState, error) {
	var state *community.CommunityState
	if oldState == nil {
		state = community.NewCommunityState()
	} else {
		state = oldState
	}
	for _, t := range transitions {

		err := t.Validate(state)
		if err != nil {
			return state, fmt.Errorf("transition validation failed: %s", err)
		}

		patch, err := t.Patch(state)
		if err != nil {
			return state, err
		}

		marshaledState, err := json.Marshal(state)
		if err != nil {
			return state, err
		}

		jsonState, err := NewJSONState(community.CommunityStateSchema, marshaledState)
		if err != nil {
			return state, err
		}

		newStateBytes, err := jsonState.ValidatePatch(patch)
		if err != nil {
			return state, err
		}

		var newState community.CommunityState
		err = json.Unmarshal(newStateBytes, &newState)
		if err != nil {
			return state, err
		}

		state = &newState
	}

	return state, nil
}

// use this if you expect the transitions to cause an error
func testTransitionsOnCopy(oldState *community.CommunityState, transitions []CommunityStateTransition) (*community.CommunityState, error) {
	marshaled, err := json.Marshal(oldState)
	if err != nil {
		return nil, err
	}

	var copy community.CommunityState
	err = json.Unmarshal(marshaled, &copy)
	if err != nil {
		return nil, err
	}

	return testTransitions(&copy, transitions)
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

	state, err := testTransitions(nil, transitions)
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

	_, err = testTransitions(nil, transitions)
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

	_, err = testTransitions(nil, transitions)
	assert.NotNil(t, err)
}

func TestProcessManagement(t *testing.T) {
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
				ID:          "node1",
				MemberID:    "jorts",
				Certificate: []byte("mycert"),
			},
		},
		&AddNodeTransition{
			Node: &community.Node{
				ID:          "node2",
				MemberID:    "jorts",
				Certificate: []byte("mycert"),
			},
		},
		&StartProcessTransition{
			Process: &community.Process{
				ID:      "proc1",
				AppName: "app1",
				Args:    []string{},
				Env:     []string{},
				Flags:   []string{},
			},
		},
		&StartProcessTransition{
			Process: &community.Process{
				ID:      "proc2",
				AppName: "app2",
				Args:    []string{},
				Env:     []string{},
				Flags:   []string{},
			},
		},
		&StartProcessInstanceTransition{
			ProcessInstance: &community.ProcessInstance{
				ProcessID: "proc1",
				NodeID:    "node1",
			},
		},
		&StartProcessInstanceTransition{
			ProcessInstance: &community.ProcessInstance{
				ProcessID: "proc1",
				NodeID:    "node2",
			},
		},
	}
	state, err := testTransitions(nil, transitions)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(state.Processes))
	assert.Equal(t, 2, len(state.ProcessInstances))

	// try starting a process with the same id
	_, err = testTransitionsOnCopy(state, []CommunityStateTransition{
		&StartProcessTransition{
			Process: &community.Process{
				ID:      "proc1",
				AppName: "app1",
			},
		},
	})
	assert.NotNil(t, err)

	// try starting a process instance on a node that does not exist
	_, err = testTransitionsOnCopy(state, []CommunityStateTransition{
		&StartProcessInstanceTransition{
			ProcessInstance: &community.ProcessInstance{
				ProcessID: "proc1",
				NodeID:    "node3",
			},
		},
	})
	assert.NotNil(t, err)

	// try starting a process instance for a process that does not exist
	_, err = testTransitionsOnCopy(state, []CommunityStateTransition{
		&StartProcessInstanceTransition{
			ProcessInstance: &community.ProcessInstance{
				ProcessID: "proc3",
				NodeID:    "node1",
			},
		},
	})
	assert.NotNil(t, err)

	// try starting a duplicate process instance
	_, err = testTransitionsOnCopy(state, []CommunityStateTransition{
		&StartProcessInstanceTransition{
			ProcessInstance: &community.ProcessInstance{
				ProcessID: "proc1",
				NodeID:    "node1",
			},
		},
	})
	assert.NotNil(t, err)

	// try stopping process with instances still running
	_, err = testTransitionsOnCopy(state, []CommunityStateTransition{
		&StopProcessTransition{
			ProcessID: "proc1",
		},
	})
	assert.NotNil(t, err)

	// Stop proc1
	state, err = testTransitions(state, []CommunityStateTransition{
		&StopProcessInstanceTransition{
			ProcessID: "proc1",
			NodeID:    "node1",
		},
		&StopProcessInstanceTransition{
			ProcessID: "proc1",
			NodeID:    "node2",
		},
		&StopProcessTransition{
			ProcessID: "proc1",
		},
	})
	assert.Nil(t, err)
	assert.Equal(t, 1, len(state.Processes))
}

func TestImplementations(t *testing.T) {
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
				ID:          "node1",
				MemberID:    "jorts",
				Certificate: []byte("mycert"),
			},
		},
		&StartProcessTransition{
			Process: &community.Process{
				ID:          "proc1",
				AppName:     "app1",
				Args:        []string{},
				Env:         []string{},
				Flags:       []string{},
				IsDatastore: true,
			},
		},
		&StartProcessTransition{
			Process: &community.Process{
				ID:          "proc2",
				AppName:     "app2",
				Args:        []string{},
				Env:         []string{},
				Flags:       []string{},
				IsDatastore: true,
			},
		},
		&AddImplementationTransition{
			InterfaceHash: "abc",
			DatastoreID:   "proc1",
		},
		&AddImplementationTransition{
			InterfaceHash: "def",
			DatastoreID:   "proc1",
		},
		&AddImplementationTransition{
			InterfaceHash: "abc",
			DatastoreID:   "proc2",
		},
		&AddImplementationTransition{
			InterfaceHash: "def",
			DatastoreID:   "ipfs",
		},
	}

	state, err := testTransitions(nil, transitions)
	fmt.Printf("%+v", state)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(state.DexImplementations))
	assert.NotNil(t, state.DexImplementations["abc"])
	assert.Equal(t, 2, len(state.DexImplementations["abc"].Implementations))
	assert.Equal(t, 2, len(state.DexImplementations["def"].Implementations))

	// Test adding and removing impl to proc1 datastore
	state, err = testTransitionsOnCopy(state, []CommunityStateTransition{
		&AddImplementationTransition{
			InterfaceHash: "abc",
			DatastoreID:   "proc1",
		},
	})
	assert.NotNil(t, err)
	assert.Equal(t, 2, len(state.DexImplementations["abc"].Implementations))

	state, err = testTransitionsOnCopy(state, []CommunityStateTransition{
		&RemoveImplementationTransition{
			InterfaceHash: "abc",
			DatastoreID:   "proc1",
		},
	})
	assert.Nil(t, err)
	assert.NotNil(t, state.DexImplementations["abc"])
	assert.Equal(t, 1, len(state.DexImplementations["abc"].Implementations))

	_, err = testTransitionsOnCopy(state, []CommunityStateTransition{
		&RemoveImplementationTransition{
			InterfaceHash: "abc",
			DatastoreID:   "ipfs",
		},
	})
	assert.NotNil(t, err)
	assert.Equal(t, 2, len(state.DexImplementations["def"].Implementations))
}
