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

	TransitionTypeStartProcess         = "start_process"
	TransitionTypeStopProcess          = "stop_process"
	TransitionTypeStartProcessInstance = "start_process_instance"
	TransitionTypeStopProcessInstance  = "stop_process_instance"

	TransitionTypeInitializeCounter = "initialize_counter"
	TransitionTypeIncrementCounter  = "increment_counter"

	TransitionTypeInitializeIPFSSwarm = "initialize_ipfs_swarm"

	TransitionTypeAddImplementation    = "add_implementation"
	TransitionTypeRemoveImplementation = "remove_implementation"
)

type TransitionWrapper struct {
	Type       string `json:"type"`
	Patch      []byte `json:"patch"`      // The JSON patch generated from the transition struct
	Transition []byte `json:"transition"` // JSON encoded transition struct
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

	transition, err := json.Marshal(t)
	if err != nil {
		return nil, err
	}

	return &TransitionWrapper{
		Type:       t.Type(),
		Patch:      patch,
		Transition: transition,
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

type StartProcessTransition struct {
	Process *community.Process
}

func (t *StartProcessTransition) Type() string {
	return TransitionTypeStartProcess
}

func (t *StartProcessTransition) Patch(oldState *community.CommunityState) ([]byte, error) {
	marshaledProcess, err := json.Marshal(t.Process)
	if err != nil {
		return nil, err
	}

	return []byte(fmt.Sprintf(`[{
		"op": "add",
		"path": "/processes/-",
		"value": %s
	}]`, string(marshaledProcess))), nil
}

func (t *StartProcessTransition) Validate(oldState *community.CommunityState) error {
	// make sure no process with same process id and app name is already started
	for _, p := range oldState.Processes {
		if p.ID == t.Process.ID {
			return fmt.Errorf("process ID %s already used in community %s", t.Process.ID, oldState.CommunityID)
		}
		if p.AppName == t.Process.AppName {
			return fmt.Errorf("community %s already has app %s running", oldState.CommunityID, t.Process.ID)
		}
	}

	return nil
}

type StopProcessTransition struct {
	ProcessID string
}

func (t *StopProcessTransition) Type() string {
	return TransitionTypeStopProcess
}

func (t *StopProcessTransition) Patch(oldState *community.CommunityState) ([]byte, error) {
	for i, p := range oldState.Processes {
		if p.ID == t.ProcessID {
			return []byte(fmt.Sprintf(`[{
				"op": "remove",
				"path": "/processes/%d"
			}]`, i)), nil
		}
	}

	return nil, fmt.Errorf("process ID %s not found in processes list", t.ProcessID)
}

func (t *StopProcessTransition) Validate(oldState *community.CommunityState) error {
	// ensure no process_instances with the process ID are still around
	for _, i := range oldState.ProcessInstances {
		if i.ProcessID == t.ProcessID {
			return fmt.Errorf("there are still process instances with process ID %s", t.ProcessID)
		}
	}

	for _, p := range oldState.Processes {
		if p.ID == t.ProcessID {
			return nil
		}
	}

	return fmt.Errorf("process ID %s not found in processes list", t.ProcessID)
}

type StartProcessInstanceTransition struct {
	ProcessInstance *community.ProcessInstance
}

func (t *StartProcessInstanceTransition) Type() string {
	return TransitionTypeStartProcessInstance
}

func (t *StartProcessInstanceTransition) Patch(oldState *community.CommunityState) ([]byte, error) {
	marshaledProcessInstance, err := json.Marshal(t.ProcessInstance)
	if err != nil {
		return nil, err
	}

	return []byte(fmt.Sprintf(`[{
		"op": "add",
		"path": "/process_instances/-",
		"value": %s
	}]`, string(marshaledProcessInstance))), nil
}

func (t *StartProcessInstanceTransition) Validate(oldState *community.CommunityState) error {
	found := false
	for _, p := range oldState.Processes {
		if p.ID == t.ProcessInstance.ProcessID {
			found = true
		}
	}
	if !found {
		return fmt.Errorf("no process with ID %s found in community %s", t.ProcessInstance.ProcessID, oldState.CommunityID)
	}

	found = false
	for _, n := range oldState.Nodes {
		if n.ID == t.ProcessInstance.NodeID {
			found = true
		}
	}
	if !found {
		return fmt.Errorf("no node with ID %s found in community %s", t.ProcessInstance.NodeID, oldState.CommunityID)
	}

	for _, i := range oldState.ProcessInstances {
		if i.NodeID == t.ProcessInstance.NodeID && i.ProcessID == t.ProcessInstance.ProcessID {
			return fmt.Errorf("process instane with matching node and process IDs found")
		}
	}
	return nil
}

type StopProcessInstanceTransition struct {
	ProcessID string
	NodeID    string
}

func (t *StopProcessInstanceTransition) Type() string {
	return TransitionTypeStopProcessInstance
}

func (t *StopProcessInstanceTransition) Patch(oldState *community.CommunityState) ([]byte, error) {
	for i, p := range oldState.ProcessInstances {
		if p.ProcessID == t.ProcessID && p.NodeID == t.NodeID {
			return []byte(fmt.Sprintf(`[{
				"op": "remove",
				"path": "/process_instances/%d"
			}]`, i)), nil
		}
	}

	return nil, fmt.Errorf("process ID %s not found in processes list", t.ProcessID)
}

func (t *StopProcessInstanceTransition) Validate(oldState *community.CommunityState) error {
	for _, p := range oldState.ProcessInstances {
		if p.ProcessID == t.ProcessID && p.NodeID == t.NodeID {
			return nil
		}
	}

	return fmt.Errorf("process instance with matching process and node IDs not found")
}

type AddImplementationTransition struct {
	InterfaceHash string
	DatastoreID   string
	// We need to decide whether to keep the CID in the state machine. Clients should
	// never need to know about the CID, so why keep it in the state machine. Each datastore
	// could maintain the mapping of interface to CIDs instead.
	ContentIdentifier string
}

func (t *AddImplementationTransition) Type() string {
	return TransitionTypeAddImplementation
}

func (t *AddImplementationTransition) Patch(oldState *community.CommunityState) ([]byte, error) {
	// First try to see if there are other impls for the same hash that we can append to the list of
	for _, impls := range oldState.DexImplementations {
		if impls.InterfaceHash == t.InterfaceHash {
			for _, impl := range impls.Implementations {
				if impl.DatastoreID == t.DatastoreID {
					return nil, fmt.Errorf("implementation with datastore %s already exists", t.DatastoreID)
				}
			}
			return []byte(fmt.Sprintf(`[{
				"op": "add",
				"path": "/dex_implementations/%s/implementations/-",
				"value": {
					"datastore_id": "%s",
					"content_identifier": "%s"
				}
			}]`, t.InterfaceHash, t.DatastoreID, t.ContentIdentifier)), nil
		}
	}

	// If not available, create a new list
	return []byte(fmt.Sprintf(`[{
		"op": "add",
		"path": "/dex_implementations/%s",
		"value": {
			"interface_hash": "%s",
			"implementations": [
				{
					"datastore_id": "%s",
					"content_identifier": "%s" 
				}
			]
		}
	}]`, t.InterfaceHash, t.InterfaceHash, t.DatastoreID, t.ContentIdentifier)), nil
}

func (t *AddImplementationTransition) Validate(oldState *community.CommunityState) error {
	// Check that datastore_id matches an existing process_id that has the is_datastore flag set, or that datastore_id is "ipfs"
	validDatastore := false
	if t.DatastoreID == "ipfs" {
		validDatastore = true
	}

	// This validations should be removed eventually. If the datastore is restarted with a different
	// process ID, we will enter a bad state.
	// Option 1: cascade process id changes to datastore ids in impl mappings
	// Option 2: create a persistent identifier for datastores that persists across restarts
	for _, p := range oldState.Processes {
		if p.ID == t.DatastoreID && p.IsDatastore {
			validDatastore = true
		}
	}

	if !validDatastore {
		return fmt.Errorf("datastore ID %s not found in processes list", t.DatastoreID)
	}

	// Check that all interface hashes are unique
	for _, impls := range oldState.DexImplementations {
		if impls.InterfaceHash == t.InterfaceHash {
			for _, impl := range impls.Implementations {
				if impl.DatastoreID == t.DatastoreID {
					return fmt.Errorf("interface hash %s already exists", t.InterfaceHash)
				}
			}
		}
	}
	return nil
}

type RemoveImplementationTransition struct {
	InterfaceHash string
	DatastoreID   string
}

func (t *RemoveImplementationTransition) Type() string {
	return TransitionTypeRemoveImplementation
}

func (t *RemoveImplementationTransition) Patch(oldState *community.CommunityState) ([]byte, error) {
	for i, impls := range oldState.DexImplementations {
		if impls.InterfaceHash == t.InterfaceHash {
			for j, impl := range impls.Implementations {
				if impl.DatastoreID == t.DatastoreID {
					return []byte(fmt.Sprintf(`[{
						"op": "remove",
						"path": "/dex_implementations/%s/implementations/%d"
					}]`, i, j)), nil
				}
			}
		}
	}

	return nil, fmt.Errorf("implementation with datastore %s not found", t.DatastoreID)
}

func (t *RemoveImplementationTransition) Validate(oldState *community.CommunityState) error {
	for _, impls := range oldState.DexImplementations {
		if impls.InterfaceHash == t.InterfaceHash {
			for _, impl := range impls.Implementations {
				if impl.DatastoreID == t.DatastoreID {
					return nil
				}
			}
		}
	}

	return fmt.Errorf("implementation with datastore %s not found", t.DatastoreID)
}
