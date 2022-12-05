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
	case state.TransitionTypeInitializeIPFSSwarm:
		return e.InitializeIPFSSwarm
	case state.TransitionTypeStartProcess:
		return e.StartProcess
	case state.TransitionTypeStopProcess:
		return e.StopProcess
	case state.TransitionTypeStartProcessInstance:
		return e.StartProcessInstance
	case state.TransitionTypeStopProcessInstance:
		return e.StartProcessInstance
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

func (e *CommunityExecutor) InitializeIPFSSwarm(update *state.StateUpdate) error {
	/*newState, err := update.State()
	if err != nil {
		return err
	}
	communityIPFSConfig, err := json.Marshal(newState.IPFSConfig)
	if err != nil {
		return fmt.Errorf("error marshaling IPFS config: %s", err)
	}

	communityIPFSConfigB64 := base64.StdEncoding.EncodeToString(communityIPFSConfig)
	if err != nil {
		return fmt.Errorf("error base64 encoding IPFS config: %s", err)
	}

	ipfsPath := filepath.Join(compass.CommunitiesPath(), newState.CommunityID, "ipfs")
	args := []string{ipfsPath}
	flags := []string{"-c", communityIPFSConfigB64}

	_, err = e.node.ProcessManager.StartProcessInstance("ipfs-driver", newState.CommunityID, args, []string{}, flags)
	if err != nil {
		return fmt.Errorf("error starting IPFS driver process: %s", err)
	}*/
	return nil
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

	if transition.ProcessInstance.NodeID == e.node.ID {
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
