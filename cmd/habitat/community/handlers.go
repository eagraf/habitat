package community

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/eagraf/habitat/pkg/compass"
	"github.com/eagraf/habitat/structs/ctl"
	"github.com/rs/zerolog/log"
)

type CTLHandler func(req *ctl.Request) (*ctl.Response, error)

// TODO: make request args a map
// for now: always send [name, address]
func (m *Manager) CommunityCreateHandler(req *ctl.Request) (*ctl.Response, error) {
	community, err := m.CreateCommunity(req.Args[0], req.Args[1], req.Args[2] == "true")
	if err != nil {
		log.Err(err)
		return &ctl.Response{
			Status:  500,
			Message: err.Error(),
		}, err
	}

	bytes, err := json.Marshal(community)
	if err != nil {
		log.Err(err)
		return &ctl.Response{
			Status:  500,
			Message: err.Error(),
		}, err
	}
	return &ctl.Response{
		Status:  ctl.StatusOK,
		Message: string(bytes),
	}, nil
}

// TODO: make request args a map
// for now: always send [name, swarmkey, bootstrap peer (one only), address, communityid]
func (m *Manager) CommunityJoinHandler(req *ctl.Request) (*ctl.Response, error) {
	// validate args
	if len(req.Args) < 2 {
		return nil, errors.New("need atleast 2 arguments to join community")
	}

	community, err := m.JoinCommunity(req.Args[0], req.Args[1], []string{req.Args[2]}, req.Args[3], req.Args[4])
	if err != nil {
		log.Err(err)
		return &ctl.Response{
			Status:  500,
			Message: err.Error(),
		}, err
	}

	bytes, err := json.Marshal(community)
	if err != nil {
		log.Err(err)
		return &ctl.Response{
			Status:  500,
			Message: err.Error(),
		}, err
	}
	return &ctl.Response{
		Status:  ctl.StatusOK,
		Message: string(bytes),
	}, nil
}

func (m *Manager) CommunityStateHandler(req *ctl.Request) (*ctl.Response, error) {
	// validate args
	if len(req.Args) != 1 {
		return nil, errors.New("need 1 argument to get community state")
	}

	communityID := req.Args[0]

	state, err := m.GetState(communityID)
	if err != nil {
		return nil, err
	}

	return &ctl.Response{
		Status:  ctl.StatusOK,
		Message: string(state),
	}, nil
}

func (m *Manager) CommunityAddMemberHandler(req *ctl.Request) (*ctl.Response, error) {
	if len(req.Args) != 3 {
		return &ctl.Response{
			Status:  ctl.StatusInternalServerError,
			Message: errors.New("need 3 arguments to get add community member").Error(),
		}, nil
	}

	communityID := req.Args[0]
	nodeID := req.Args[1]
	address := req.Args[2]

	err := m.clusterManager.AddNode(communityID, nodeID, address)
	if err != nil {
		log.Err(err)
		return &ctl.Response{
			Status:  ctl.StatusInternalServerError,
			Message: err.Error(),
		}, nil
	}

	return &ctl.Response{
		Status:  ctl.StatusOK,
		Message: fmt.Sprintf("added node %s to community %s", nodeID, communityID),
	}, nil
}

func (m *Manager) CommunityProposeHandler(req *ctl.Request) (*ctl.Response, error) {
	// validate args
	if len(req.Args) != 2 {
		return nil, errors.New("need 2 arguments to make proposal")
	}

	err := m.ProposeTransition(req.Args[0], []byte(req.Args[1]))
	if err != nil {
		return nil, err
	}

	return &ctl.Response{
		Status:  ctl.StatusOK,
		Message: fmt.Sprintf("proposed transition"),
	}, nil
}

func (m *Manager) CommunityListHandler(req *ctl.Request) (*ctl.Response, error) {
	b := strings.Builder{}
	b.WriteString(fmt.Sprintf("node id: %s\n", compass.PeerID()))

	communityDir := compass.CommunitiesPath()

	_, err := os.Stat(communityDir)
	if errors.Is(err, os.ErrNotExist) {
		return &ctl.Response{
			Status:  ctl.StatusOK,
			Message: b.String(),
		}, nil
	} else if err != nil {
		return nil, err
	}

	communities, err := ioutil.ReadDir(communityDir)
	if err != nil {
		return nil, err
	}

	for _, c := range communities {
		b.WriteString(fmt.Sprintf("%s\n", c.Name()))
	}

	return &ctl.Response{
		Status:  ctl.StatusOK,
		Message: b.String(),
	}, nil
}
