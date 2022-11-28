package state

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/eagraf/habitat/structs/community"
)

const (
	TransitionTypeInitializeCommunity = "initialize_community"
	TransitionTypeAddMember           = "add_member"
	TransitionTypeAddNode             = "add_node"

	TransitionTypeInitializeCounter = "initialize_counter"
	TransitionTypeIncrementCounter  = "increment_counter"

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

type AddMemberTransition struct {
	Member *community.Member
}

func (t *AddMemberTransition) Type() string {
	return TransitionTypeAddMember
}

func (t *AddMemberTransition) Patch(oldState *community.CommunityState) ([]byte, error) {
	marshaledMember, err := json.Marshal(t.Member)
	if err != nil {
		return nil, err
	}

	return []byte(fmt.Sprintf(`[{
		"op": "add",
		"path": "/members/-",
		"value": %s
	}]`, string(marshaledMember))), nil
}

func (t *AddMemberTransition) Validate(oldState *community.CommunityState) error {
	// TODO cryptographically verify all of this stuff
	// Make sure member is not already a part of the community
	for _, m := range oldState.Members {
		if m.ID == t.Member.ID {
			return fmt.Errorf("member with ID %s already in community", m.ID)
		}
	}

	return nil
}

type AddNodeTransition struct {
	Node *community.Node
}

func (t *AddNodeTransition) Type() string {
	return TransitionTypeAddNode
}

func (t *AddNodeTransition) Patch(oldState *community.CommunityState) ([]byte, error) {
	marshaledNode, err := json.Marshal(t.Node)
	if err != nil {
		return nil, err
	}

	return []byte(fmt.Sprintf(`[{
		"op": "add",
		"path": "/nodes/-",
		"value": %s
	}]`, string(marshaledNode))), nil
}

func (t *AddNodeTransition) Validate(oldState *community.CommunityState) error {
	found := false
	for _, m := range oldState.Members {
		if t.Node.MemberID == m.ID {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("associated member %s is not in the community", t.Node.MemberID)
	}

	for _, n := range oldState.Nodes {
		if n.ID == t.Node.ID {
			return fmt.Errorf("node with ID %s already in community", n.ID)
		}
	}

	return nil
}
