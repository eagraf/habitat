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

type CommunityExecutor struct {
	processManager *procs.Manager
}

func NewCommunityExecutor(p *procs.Manager) *CommunityExecutor {
	return &CommunityExecutor{
		processManager: p,
	}
}

func (e *CommunityExecutor) Execute(update *state.StateUpdate) {
	log.Info().Msgf("executing %s state transition", update.TransitionType)
	newState, err := update.State()
	if err != nil {
		log.Error().Msgf("error getting new state: %s", err)
	}
	switch update.TransitionType {
	case state.TransitionTypeInitializeIPFSSwarm:
		communityIPFSConfig, err := json.Marshal(newState.IPFSConfig)
		if err != nil {
			log.Error().Msgf("error marshaling IPFS config")
		}

		communityIPFSConfigB64 := base64.StdEncoding.EncodeToString(communityIPFSConfig)
		if err != nil {
			log.Error().Msgf("error base64 encoding IPFS config")
		}

		ipfsPath := filepath.Join(compass.CommunitiesPath(), newState.CommunityID, "ipfs")
		args := []string{ipfsPath}
		flags := []string{"-c", communityIPFSConfigB64}

		fmt.Println(newState.CommunityID)
		procID, err := e.processManager.StartProcess("ipfs-driver", newState.CommunityID, args, []string{}, flags)
		if err != nil {
			log.Error().Msgf("error starting IPFS driver process")
		}
		fmt.Println(procID)
	}
}
