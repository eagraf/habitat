package state

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/eagraf/habitat/structs/community"
)

const (
	TransitionTypeInitializeCommunity = "initialize_community"
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
	Patch(oldState *community.CommunityState) ([]byte, error)
	Validate(oldState *community.CommunityState) error
}

func wrapTransition(t CommunityStateTransition, oldState *community.CommunityState) (*TransitionWrapper, error) {
	patch, err := t.Patch(oldState)
	if err != nil {
		return nil, err
	}

	return &TransitionWrapper{
		Type:  t.Type(),
		Patch: patch,
	}, nil
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

type InitializeCommunityTransition struct {
	CommunityID string
}

func (t *InitializeCommunityTransition) Type() string {
	return TransitionTypeInitializeCommunity
}

func (t *InitializeCommunityTransition) Patch(oldState *community.CommunityState) ([]byte, error) {
	return []byte(fmt.Sprintf(`[{
		"op": "add",
		"path": "/community_id",
		"value": "%s"
	}]`, t.CommunityID)), nil
}

func (t *InitializeCommunityTransition) Validate(oldState *community.CommunityState) error {
	if oldState.CommunityID != "" {
		return errors.New("community_id already initialized")
	}
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
		"op": "replace",
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
