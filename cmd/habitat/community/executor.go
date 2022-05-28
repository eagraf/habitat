package community

import (
	"github.com/eagraf/habitat/cmd/habitat/community/state"
	"github.com/rs/zerolog/log"
)

type CommunityExecutor struct {
}

func (e *CommunityExecutor) Execute(update *state.StateUpdate) {
	log.Info().Msgf("executing %s state transition", update.TransitionType)
	newState, err := update.State()
	if err != nil {
		log.Error().Msgf("error getting new state: %s", err)
	}
	switch update.TransitionType {
	case state.TransitionTypeInitializeIPFSSwarm:
		err := joinIPFSSwarm(newState.IPFSConfig)
		if err != nil {
			log.Error().Msgf("error joining IPFS swarm: %s", err)
		}
	}

}
