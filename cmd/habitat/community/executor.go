package community

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/eagraf/habitat/cmd/habitat/community/state"
	"github.com/eagraf/habitat/cmd/habitat/procs"
	"github.com/eagraf/habitat/pkg/compass"
	"github.com/rs/zerolog/log"
)

type TransitionExecutor func(update *state.StateUpdate) error

// CommunityExecutor implements the state.Executor interface.
// It receives incoming state transitions from the replicated state machine,
// and executes the commands on the running node.
type CommunityExecutor struct {
	processManager *procs.Manager
}

func NewCommunityExecutor(p *procs.Manager) *CommunityExecutor {
	return &CommunityExecutor{
		processManager: p,
	}
}

// GetTransitionExecutor maps transition types to executor functions
// TODO this is pretty silly, lets find a way to do this with reflection or equivalent
func (e *CommunityExecutor) GetTransitionExecutor(transitionType string) TransitionExecutor {
	switch transitionType {
	case state.TransitionTypeInitializeIPFSSwarm:
		return e.InitializeIPFSSwarm
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
	newState, err := update.State()
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

	_, err = e.processManager.StartProcess("ipfs-driver", newState.CommunityID, args, []string{}, flags)
	if err != nil {
		return fmt.Errorf("error starting IPFS driver process: %s", err)
	}
	return nil
}
