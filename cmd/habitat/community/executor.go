package community

import (
	"encoding/json"

	"github.com/eagraf/habitat/cmd/habitat/community/state"
	"github.com/eagraf/habitat/cmd/habitat/node"
	"github.com/eagraf/habitat/cmd/habitat/procs"
	"github.com/eagraf/habitat/structs/community"
	"github.com/rs/zerolog/log"
)

type TransitionExecutor func(update *state.StateUpdate) error

// CommunityExecutor implements the state.Executor interface.
// It receives incoming state transitions from the replicated state machine,
// and executes the commands on the running node.
type CommunityExecutor struct {
	node *node.Node
}

func NewCommunityExecutor(n *node.Node) *CommunityExecutor {
	return &CommunityExecutor{
		node: n,
	}
}

// GetTransitionExecutor maps transition types to executor functions
// TODO this is pretty silly, lets find a way to do this with reflection or equivalent
func (e *CommunityExecutor) GetTransitionExecutor(transitionType string) TransitionExecutor {
	switch transitionType {
	case state.TransitionTypeStartProcess:
		return e.StartProcess
	case state.TransitionTypeStopProcess:
		return e.StopProcess
	case state.TransitionTypeStartProcessInstance:
		return e.StartProcessInstance
	case state.TransitionTypeStopProcessInstance:
		return e.StopProcessInstance
	case state.TransitionTypeAddNode:
		return e.AddNode
	default:
		return nil
	}
}

func (e *CommunityExecutor) Execute(update *state.StateUpdate) {
	log.Info().Msgf("executing %s state transition", update.TransitionType)

	transitionExecutor := e.GetTransitionExecutor(update.TransitionType)
	if transitionExecutor != nil {
		err := transitionExecutor(update)
		if err != nil {
			log.Error().Err(err).Msgf("error executing %s", update.TransitionType)
		}
	}
}

func (e *CommunityExecutor) StartProcess(update *state.StateUpdate) error {
	return nil
}

func (e *CommunityExecutor) StopProcess(update *state.StateUpdate) error {
	return nil
}

func (e *CommunityExecutor) StartProcessInstance(update *state.StateUpdate) error {

	var transition state.StartProcessInstanceTransition
	err := json.Unmarshal(update.Transition, &transition)
	if err != nil {
		return err
	}

	state, err := update.State()
	if err != nil {
		return err
	}

	var process *community.Process
	for _, p := range state.Processes {
		if p.ID == transition.ProcessInstance.ProcessID {
			process = p
		}
	}

	// Hash together the process ID, community ID and node ID to get a process instance ID

	if transition.ProcessInstance.NodeID == e.node.ID {
		_, err := e.node.ProcessManager.StartProcessInstance(state.CommunityID, process.ID, process.AppName, process.Args, process.Env, process.Flags)
		if err != nil {
			return err
		}
	}

	return nil
}

func (e *CommunityExecutor) StopProcessInstance(update *state.StateUpdate) error {
	var transition state.StopProcessInstanceTransition
	err := json.Unmarshal(update.Transition, &transition)
	if err != nil {
		return err
	}

	state, err := update.State()
	if err != nil {
		return err
	}

	var process *community.Process
	for _, p := range state.Processes {
		if p.ID == transition.ProcessID {
			process = p
		}
	}

	if transition.NodeID == e.node.ID {
		processInstanceID, err := procs.GetProcessInstanceID(state.CommunityID, e.node.ID, process.ID)
		if err != nil {
			return err
		}
		err = e.node.ProcessManager.StopProcessInstance(processInstanceID)
		if err != nil {
			return err
		}
	}

	return nil
}

func (e *CommunityExecutor) AddNode(update *state.StateUpdate) error {
	var transition state.AddNodeTransition
	err := json.Unmarshal(update.Transition, &transition)
	if err != nil {
		return err
	}

	// Check if the node is this instance
	if transition.Node.ID == e.node.ID {
		return nil
	}

	// HAX: req IPFS to add node as peer
	_, err = e.node.IPFSClient.AddPeer(transition.Node.IPFSSwarmAddress)
	if err != nil {
		return err
	}

	// Add node to list of data proxy peer nodes
	err = e.node.DataProxy.AddPeerNode(transition.Node.Address)
	if err != nil {
		return err
	}

	return nil
}
