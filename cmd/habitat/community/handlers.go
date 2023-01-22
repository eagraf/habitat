package community

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/eagraf/habitat/cmd/habitat/api"
	"github.com/eagraf/habitat/cmd/habitat/community/state"
	"github.com/eagraf/habitat/cmd/habitat/procs"
	"github.com/eagraf/habitat/pkg/compass"
	"github.com/eagraf/habitat/pkg/identity"
	"github.com/eagraf/habitat/structs/community"
	"github.com/eagraf/habitat/structs/ctl"
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"
)

func signKeyExchange(conn *websocket.Conn, finalMsg ctl.WebsocketMessage, nodeID string) (*community.Member, *community.Node, error) {
	pubMsg := &ctl.SigningPublicKeyMsg{}

	// Generate the private key first
	pub, _, err := identity.GenerateMemberNodeKeypair()
	if err != nil {
		api.WriteWebsocketError(conn, err, pubMsg)
		return nil, nil, err
	}

	// Send public key to client to be signed by user private key
	pubMsg.PublicKey = pub
	pubMsg.NodeID = nodeID

	err = conn.WriteJSON(pubMsg)
	if err != nil {
		api.WriteWebsocketError(conn, err, pubMsg)
		return nil, nil, err
	}

	// Wait for client response with signed cert
	var certMsg ctl.SigningCertMsg
	err = conn.ReadJSON(&certMsg)
	if err != nil {
		api.WriteWebsocketError(conn, err, finalMsg)
		return nil, nil, err
	}

	cert, err := identity.GetCertFromPEM(certMsg.NodeCertificate)
	if err != nil {
		api.WriteWebsocketError(conn, err, finalMsg)
		return nil, nil, err
	}

	userID, err := identity.GetUIDFromName(&cert.Issuer)
	if err != nil {
		api.WriteWebsocketError(conn, err, finalMsg)
		return nil, nil, err
	}

	returnedNodeID, err := identity.GetUIDFromName(&cert.Subject)
	if err != nil {
		api.WriteWebsocketError(conn, err, finalMsg)
		return nil, nil, err
	}

	// sanity check that the nodeID we generated is now encoded in the certificate
	if returnedNodeID != nodeID {
		api.WriteWebsocketError(conn, fmt.Errorf("node ID in returned certificate does not match one generated: %s, %s", returnedNodeID, nodeID), finalMsg)
		return nil, nil, err
	}

	member := &community.Member{
		ID:          userID,
		Username:    cert.Issuer.CommonName,
		Certificate: certMsg.UserCertificate,
	}

	node := &community.Node{
		ID:          nodeID,
		P2PID:       compass.PeerID().String(),
		MemberID:    userID,
		Certificate: certMsg.NodeCertificate,
	}

	return member, node, nil
}

// TODO: make request args a map
// for now: always send [name, address]
func (m *Manager) CommunityCreateHandler(w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.Upgrader{}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		api.WriteError(w, http.StatusInternalServerError, err)
	}
	defer api.WriteWebsocketClose(conn)

	var commRes ctl.CommunityCreateResponse
	member, node, err := signKeyExchange(conn, &commRes, compass.NodeID())
	if err != nil {
		// signKeyExchange should have already sent an error back
		return
	}

	var commReq ctl.CommunityCreateRequest
	err = conn.ReadJSON(&commReq)
	if err != nil {
		api.WriteWebsocketError(conn, err, &commRes)
		return
	}

	publicMa, err := compass.PublicLibP2PMultiaddr()
	if err != nil {
		api.WriteWebsocketError(conn, err, &commRes)
		return
	}

	ipfsSwarmMa, err := compass.IPFSSwarmAddr()
	if err != nil {
		api.WriteWebsocketError(conn, err, &commRes)
		return
	}

	node.Address = publicMa.String()
	node.IPFSSwarmAddress = ipfsSwarmMa.String()
	community, err := m.CreateCommunity(commReq.CommunityName, commReq.CreateIPFSCluster, member, node)
	if err != nil {
		api.WriteWebsocketError(conn, err, &commRes)
		return
	}

	joinInfo := &ctl.JoinInfo{
		CommunityID: community.CommunityID,
		Address:     publicMa.String(),
	}

	marshaled, err := json.Marshal(joinInfo)
	if err != nil {
		api.WriteWebsocketError(conn, err, &commRes)
		return
	}

	encoded := base64.StdEncoding.EncodeToString([]byte(marshaled))

	commRes.CommunityID = community.CommunityID
	commRes.JoinToken = encoded

	err = conn.WriteJSON(&commRes)
	if err != nil {
		// Client is not expecting any more messages, so we just close the connection
		log.Error().Msgf("error writing final community creation response to websocket: %s", err)
		return
	}
}

// TODO: make request args a map
// for now: always send [name, swarmkey, bootstrap peer (one only), address, communityid]
func (m *Manager) CommunityJoinHandler(w http.ResponseWriter, r *http.Request) {

	upgrader := websocket.Upgrader{}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		api.WriteError(w, http.StatusInternalServerError, err)
		return
	}
	defer api.WriteWebsocketClose(conn)

	var commRes ctl.CommunityJoinResponse
	newMember, newNode, err := signKeyExchange(conn, &commRes, compass.NodeID())
	if err != nil {
		// signKeyExchange should have already sent an error back
		return
	}

	var commReq ctl.CommunityJoinRequest
	conn.ReadJSON(&commReq)
	if err != nil {
		api.WriteWebsocketError(conn, err, &commRes)
		return
	}

	_, err = m.JoinCommunity(commReq.CommunityName, commReq.SwarmKey, commReq.BootstrapPeers, commReq.AcceptingNodeAddr, commReq.CommunityID)
	if err != nil {
		api.WriteWebsocketError(conn, err, &commRes)
		return
	}

	publicMa, err := compass.PublicLibP2PMultiaddr()
	if err != nil {
		api.WriteWebsocketError(conn, err, &commRes)
		return
	}

	ipfsSwarmMa, err := compass.IPFSSwarmAddr()
	if err != nil {
		api.WriteWebsocketError(conn, err, &commRes)
		return
	}

	newNode.Address = publicMa.String()
	newNode.IPFSSwarmAddress = ipfsSwarmMa.String()
	addMemberReq := &ctl.CommunityAddMemberRequest{
		CommunityID:        commReq.CommunityID,
		NodeID:             compass.PeerID().String(),
		JoiningNodeAddress: publicMa.String(),
		Member:             newMember,
		Node:               newNode,
	}
	var addMemberRes ctl.CommunityAddMemberResponse

	reqBody, err := json.Marshal(addMemberReq)
	if err != nil {
		api.WriteWebsocketError(conn, err, &commRes)
		return
	}

	p2pReq, err := http.NewRequest("POST", "", bytes.NewReader(reqBody))
	if err != nil {
		api.WriteWebsocketError(conn, err, &commRes)
		return
	}

	bytes, err := m.node.P2PNode.PostRequestToPeer(commReq.AcceptingNodeAddr, "/habitat"+ctl.GetRoute(ctl.CommandCommunityAddMember), p2pReq)
	if err != nil {
		api.WriteWebsocketError(conn, err, &commRes)
		return
	} else if err := json.Unmarshal(bytes, &addMemberRes); err != nil {
		api.WriteWebsocketError(conn, err, &commRes)
		return
	}

	err = conn.WriteJSON(&commRes)
	if err != nil {
		// Client is not expecting any more messages, so we just close the connection
		log.Error().Msgf("error writing final community join response to websocket: %s", err)
		return
	}

}

func (m *Manager) CommunityStateHandler(w http.ResponseWriter, r *http.Request) {
	var commReq ctl.CommunityStateRequest
	err := api.BindPostRequest(r, &commReq)
	if err != nil {
		api.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	communityID := commReq.CommunityID

	state, err := m.GetState(communityID)
	if err != nil {
		api.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	marshaled, err := json.Marshal(state)
	if err != nil {
		api.WriteError(w, http.StatusInternalServerError, err)
		return
	}
	commRes := &ctl.CommunityStateResponse{
		State: marshaled,
	}

	api.WriteResponse(w, commRes)
}

func (m *Manager) CommunityAddMemberHandler(w http.ResponseWriter, r *http.Request) {
	var commReq ctl.CommunityAddMemberRequest
	err := api.BindPostRequest(r, &commReq)
	if err != nil {
		api.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	communityID := commReq.CommunityID
	nodeID := commReq.NodeID
	address := commReq.JoiningNodeAddress

	err = m.clusterManager.AddNode(communityID, nodeID, address)
	if err != nil {
		api.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	_, err = m.AddMemberNode(communityID, commReq.Member, commReq.Node)
	if err != nil {
		api.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	commRes := &ctl.CommunityAddMemberResponse{}
	api.WriteResponse(w, commRes)
}

func (m *Manager) CommunityProposeHandler(w http.ResponseWriter, r *http.Request) {
	var commReq ctl.CommunityProposeRequest
	err := api.BindPostRequest(r, &commReq)
	if err != nil {
		api.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	err = m.ProposeTransitions(commReq.CommunityID, commReq.StateTransition)
	if err != nil {
		api.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	commRes := &ctl.CommunityProposeResponse{}
	api.WriteResponse(w, commRes)
}

func (m *Manager) CommunityListHandler(w http.ResponseWriter, r *http.Request) {
	var commReq ctl.CommunityListRequest
	err := api.BindPostRequest(r, &commReq)
	if err != nil {
		api.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	communityDir := compass.CommunitiesPath()

	_, err = os.Stat(communityDir)
	if errors.Is(err, os.ErrNotExist) {
		api.WriteError(w, http.StatusInternalServerError, err)
		return
	} else if err != nil {
		api.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	communities, err := ioutil.ReadDir(communityDir)
	if err != nil {
		api.WriteError(w, http.StatusInternalServerError, err)
		return
	}
	commRes := &ctl.CommunityListResponse{
		NodeID:      string(compass.PeerID().Pretty()),
		Communities: make([]string, 0),
	}
	for _, c := range communities {
		commRes.Communities = append(commRes.Communities, c.Name())
	}

	api.WriteResponse(w, commRes)
}

func (m *Manager) CommunityPSHandler(w http.ResponseWriter, r *http.Request) {
	var commReq ctl.CommunityPSRequest
	err := api.BindPostRequest(r, &commReq)
	if err != nil {
		api.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	communityID := commReq.CommunityID

	state, err := m.GetState(communityID)
	if err != nil {
		api.WriteError(w, http.StatusInternalServerError, err)
		return
	}
	commRes := &ctl.CommunityPSResponse{
		Processes: []*ctl.CommunityPSProcess{},
	}

	processInstances := make(map[string][]*community.ProcessInstance)
	for _, i := range state.ProcessInstances {
		if instances, ok := processInstances[i.ProcessID]; ok {
			processInstances[i.ProcessID] = append(instances, i)
		} else {
			processInstances[i.ProcessID] = []*community.ProcessInstance{
				i,
			}
		}
	}

	for _, p := range state.Processes {
		instances, ok := processInstances[p.ID]
		if !ok {
			instances = []*community.ProcessInstance{}
		}
		commRes.Processes = append(commRes.Processes, &ctl.CommunityPSProcess{
			Process:   p,
			Instances: instances,
		})

	}

	api.WriteResponse(w, commRes)
}

func (m *Manager) CommunityStartProcessHandler(w http.ResponseWriter, r *http.Request) {
	var commReq ctl.CommunityStartProcessRequest
	err := api.BindPostRequest(r, &commReq)
	if err != nil {
		api.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	communityID := commReq.CommunityID

	stateMachine, ok := m.communities[communityID]
	if !ok {
		api.WriteError(w, http.StatusInternalServerError, fmt.Errorf("community %s is not on this instance", communityID))
		return
	}

	processID := procs.RandomProcessID()

	transitions := []state.CommunityStateTransition{
		&state.StartProcessTransition{
			Process: &community.Process{
				ID:      processID,
				AppName: commReq.App,
				Args:    commReq.Args,
				Flags:   commReq.Flags,
				Env:     commReq.Env,
			},
		},
	}

	for _, i := range commReq.InstancesNodes {
		transitions = append(transitions, &state.StartProcessInstanceTransition{
			ProcessInstance: &community.ProcessInstance{
				ProcessID: processID,
				NodeID:    i,
			},
		})
	}

	_, err = stateMachine.ProposeTransitions(transitions)
	if err != nil {
		api.WriteError(w, http.StatusInternalServerError, err)
	}

	api.WriteResponse(w, &ctl.CommunityStartProcessResponse{})
}

func (m *Manager) CommunityStopProcessHandler(w http.ResponseWriter, r *http.Request) {
	var commReq ctl.CommunityStopProcessRequest
	err := api.BindPostRequest(r, &commReq)
	if err != nil {
		api.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	communityID := commReq.CommunityID

	stateMachine, ok := m.communities[communityID]
	if !ok {
		api.WriteError(w, http.StatusInternalServerError, fmt.Errorf("community %s is not on this instance", communityID))
		return
	}

	curState, err := stateMachine.State()
	if err != nil {
		api.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	transitions := []state.CommunityStateTransition{}

	if len(commReq.InstancesNodes) == 0 {
		for _, i := range curState.ProcessInstances {
			if i.ProcessID == commReq.ProcessID {
				transitions = append(transitions, &state.StopProcessInstanceTransition{
					ProcessID: i.ProcessID,
					NodeID:    i.NodeID,
				})
			}
		}
		transitions = append(transitions, &state.StopProcessTransition{
			ProcessID: commReq.ProcessID,
		})
	} else {
		for _, i := range commReq.InstancesNodes {
			transitions = append(transitions, &state.StopProcessInstanceTransition{
				ProcessID: commReq.ProcessID,
				NodeID:    i,
			})
		}
	}

	_, err = stateMachine.ProposeTransitions(transitions)
	if err != nil {
		api.WriteError(w, http.StatusInternalServerError, err)
	}

	api.WriteResponse(w, &ctl.CommunityStartProcessResponse{})
}
