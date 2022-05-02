package community

import (
	"errors"
	"io/ioutil"
	"os"

	"github.com/eagraf/habitat/pkg/compass"
	"github.com/eagraf/habitat/structs/ctl"
)

type CTLHandler func(req *ctl.RequestWrapper) (*ctl.ResponseWrapper, error)

// TODO: make request args a map
// for now: always send [name, address]
func (m *Manager) CommunityCreateHandler(req *ctl.RequestWrapper) (*ctl.ResponseWrapper, error) {
	var commReq ctl.CommunityCreateRequest
	err := req.Deserialize(&commReq)
	if err != nil {
		return nil, err
	}

	community, err := m.CreateCommunity(commReq.CommunityName, commReq.CreateIPFSCluster)
	if err != nil {
		return nil, err
	}

	commRes := &ctl.CommunityCreateResponse{
		CommunityID: community.Id,
	}
	res, err := ctl.NewResponseWrapper(commRes, ctl.StatusOK, "")
	if err != nil {
		return nil, err
	}
	return res, nil
}

// TODO: make request args a map
// for now: always send [name, swarmkey, bootstrap peer (one only), address, communityid]
func (m *Manager) CommunityJoinHandler(req *ctl.RequestWrapper) (*ctl.ResponseWrapper, error) {
	var commReq ctl.CommunityJoinRequest
	err := req.Deserialize(&commReq)
	if err != nil {
		return nil, err
	}

	_, err = m.JoinCommunity(commReq.CommunityName, commReq.SwarmKey, commReq.BootstrapPeers, commReq.AcceptingNodeAddr, commReq.CommunityID)
	if err != nil {
		return nil, err
	}

	commRes := &ctl.CommunityJoinResponse{}
	res, err := ctl.NewResponseWrapper(commRes, ctl.StatusOK, "")
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (m *Manager) CommunityStateHandler(req *ctl.RequestWrapper) (*ctl.ResponseWrapper, error) {
	var commReq ctl.CommunityStateRequest
	err := req.Deserialize(&commReq)
	if err != nil {
		return nil, err
	}

	communityID := commReq.CommunityID

	state, err := m.GetState(communityID)
	if err != nil {
		return nil, err
	}
	commRes := &ctl.CommunityStateResponse{
		State: state,
	}
	res, err := ctl.NewResponseWrapper(commRes, ctl.StatusOK, "")
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (m *Manager) CommunityAddMemberHandler(req *ctl.RequestWrapper) (*ctl.ResponseWrapper, error) {
	var commReq ctl.CommunityAddMemberRequest
	err := req.Deserialize(&commReq)
	if err != nil {
		return nil, err
	}

	communityID := commReq.CommunityID
	nodeID := commReq.NodeID
	address := commReq.JoiningNodeAddress

	err = m.clusterManager.AddNode(communityID, nodeID, address)
	if err != nil {
		return nil, err
	}
	commRes := &ctl.CommunityAddMemberResponse{}
	res, err := ctl.NewResponseWrapper(commRes, ctl.StatusOK, "")
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (m *Manager) CommunityProposeHandler(req *ctl.RequestWrapper) (*ctl.ResponseWrapper, error) {
	var commReq ctl.CommunityProposeRequest
	err := req.Deserialize(&commReq)
	if err != nil {
		return nil, err
	}

	err = m.ProposeTransition(commReq.CommunityID, commReq.StateTransition)
	if err != nil {
		return nil, err
	}

	commRes := &ctl.CommunityProposeResponse{}
	res, err := ctl.NewResponseWrapper(commRes, ctl.StatusOK, "")
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (m *Manager) CommunityListHandler(req *ctl.RequestWrapper) (*ctl.ResponseWrapper, error) {
	var commReq ctl.CommunityListRequest
	err := req.Deserialize(&commReq)
	if err != nil {
		return nil, err
	}

	communityDir := compass.CommunitiesPath()

	_, err = os.Stat(communityDir)
	if errors.Is(err, os.ErrNotExist) {
		return nil, err
	} else if err != nil {
		return nil, err
	}

	communities, err := ioutil.ReadDir(communityDir)
	if err != nil {
		return nil, err
	}
	commRes := &ctl.CommunityListResponse{
		NodeID:      string(compass.PeerID().Pretty()),
		Communities: make([]string, 0),
	}
	for _, c := range communities {
		commRes.Communities = append(commRes.Communities, c.Name())
	}

	res, err := ctl.NewResponseWrapper(commRes, ctl.StatusOK, "")
	if err != nil {
		return nil, err
	}
	return res, nil
}
