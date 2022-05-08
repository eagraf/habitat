package state

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/eagraf/habitat/structs/community"
)

const (
	TransitionTypeInitializeCounter   = "initialize_counter"
	TransitionTypeIncrementCounter    = "increment_counter"
	TransitionTypeInitializeIPFSSwarm = "initialize_ipfs_swarm"
)

type TransitionWrapper struct {
	Type       string `json:"type"`
	Transition []byte `json:"transition"`
}

type CommunityStateTransition interface {
	Type() string
	Patch(oldState *community.CommunityState) ([]byte, error)
	Validate(oldState *community.CommunityState) error
}

type InitializeCounterTransition struct {
	InitialCount int
}

func (t *InitializeCounterTransition) Type() string {
	return TransitionTypeInitializeCounter
}

func (t *InitializeCounterTransition) Patch(oldState *community.CommunityState) ([]byte, error) {
	return []byte(fmt.Sprintf(`[{
		"op": "add",
		"path": "/counter",
		"value": %d
	}]`, t.InitialCount)), nil
}

func (t *InitializeCounterTransition) Validate(oldState *community.CommunityState) error {
	if oldState.Counter != 0 {
		return errors.New("counter is not 0")
	}
	return nil
}

type IncrementCounterTransition struct{}

func (t *IncrementCounterTransition) Type() string {
	return TransitionTypeIncrementCounter
}

func (t *IncrementCounterTransition) Patch(oldState *community.CommunityState) ([]byte, error) {
	return []byte(fmt.Sprintf(`[{
		"op": "replace",
		"path": "/counter",
		"value": %d 
	}]`, oldState.Counter+1)), nil
}

func (t *IncrementCounterTransition) Validate(oldState *community.CommunityState) error {
	return nil
}

type InitializeIPFSSwarmTransition struct {
	IPFSConfig *community.IPFSConfig
}

func (t *InitializeIPFSSwarmTransition) Type() string {
	return TransitionTypeInitializeIPFSSwarm
}

func (t *InitializeIPFSSwarmTransition) Patch(oldState *community.CommunityState) ([]byte, error) {
	configBytes, err := json.Marshal(t.IPFSConfig)
	if err != nil {
		return nil, err
	}

	return []byte(fmt.Sprintf(`[{
		"op": "add",
		"path": "/ipfs_config",
		"value": %s
	}]`, configBytes)), nil
}

func (t *InitializeIPFSSwarmTransition) Validate(oldState *community.CommunityState) error {
	if oldState.IPFSConfig != nil {
		return errors.New("state already has an initialized IPFS config")
	}
	if len(t.IPFSConfig.BootstrapAddresses) == 0 {
		return errors.New("no bootstrap addresses supplied")
	}
	return nil
}
