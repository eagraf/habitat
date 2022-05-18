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
	Type  string `json:"type"`
	Patch []byte `json:"patch"`
}

type CommunityStateTransition interface {
	Type() string
	JSON(oldState *community.CommunityState) ([]byte, error)
	Validate(oldState *community.CommunityState) error
}

func serializeTransition(patch []byte, transitionType string) ([]byte, error) {
	wrapper := &TransitionWrapper{
		Type:  transitionType,
		Patch: patch,
	}
	return json.Marshal(wrapper)
}

type InitializeCounterTransition struct {
	InitialCount int
}

func (t *InitializeCounterTransition) Type() string {
	return TransitionTypeInitializeCounter
}

func (t *InitializeCounterTransition) JSON(oldState *community.CommunityState) ([]byte, error) {
	patch := []byte(fmt.Sprintf(`[{
		"op": "add",
		"path": "/counter",
		"value": %d
	}]`, t.InitialCount))

	return serializeTransition(patch, TransitionTypeInitializeCounter)
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

func (t *IncrementCounterTransition) JSON(oldState *community.CommunityState) ([]byte, error) {
	patch := []byte(fmt.Sprintf(`[{
		"op": "replace",
		"path": "/counter",
		"value": %d 
	}]`, oldState.Counter+1))

	return serializeTransition(patch, TransitionTypeIncrementCounter)
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

func (t *InitializeIPFSSwarmTransition) JSON(oldState *community.CommunityState) ([]byte, error) {
	configBytes, err := json.Marshal(t.IPFSConfig)
	if err != nil {
		return nil, err
	}

	patch := []byte(fmt.Sprintf(`[{
		"op": "add",
		"path": "/ipfs_config",
		"value": %s
	}]`, configBytes))

	return serializeTransition(patch, TransitionTypeInitializeIPFSSwarm)
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
