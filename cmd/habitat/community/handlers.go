package community

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/eagraf/habitat/cmd/habitat/api"
	"github.com/eagraf/habitat/pkg/compass"
	"github.com/eagraf/habitat/pkg/identity"
	"github.com/eagraf/habitat/structs/community"
	"github.com/eagraf/habitat/structs/ctl"
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"
)

func signKeyExchange(conn *websocket.Conn, finalMsg ctl.WebsocketMessage) (*community.Member, *community.Node, error) {
	pubMsg := &ctl.SigningPublicKeyMsg{}

	// Generate the private key first
	pub, _, err := identity.GenerateMemberNodeKeypair()
	if err != nil {
		api.WriteWebsocketError(conn, err, pubMsg)
		return nil, nil, err
	}

	// Send public key to client to be signed by user private key
	pubMsg.PublicKey = pub

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

	nodeID, err := identity.GetUIDFromName(&cert.Subject)
	if err != nil {
		api.WriteWebsocketError(conn, err, finalMsg)
		return nil, nil, err
	}

	member := &community.Member{
		ID:          userID,
		Username:    cert.Issuer.CommonName,
		Certificate: certMsg.UserCertificate,
	}

	node := &community.Node{
		ID:          nodeID,
		MemberID:    userID,
		Certificate: certMsg.NodeCertificate,
	}

	return member, node, nil
}

// TODO: make request args a map
// for now: always send [name, address]
func (m *Manager) CommunityCreateHandler(w http.ResponseWriter, r *http.Request) {
	var commReq ctl.CommunityCreateRequest
	err := api.BindPostRequest(r, &commReq)
	if err != nil {
		api.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	publicMa, err := compass.PublicRaftMultiaddr()
	if err != nil {
		api.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	community, err := m.CreateCommunity(commReq.CommunityName, commReq.CreateIPFSCluster)
	if err != nil {
		api.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	joinInfo := &ctl.JoinInfo{
		CommunityID: community.CommunityID,
		Address:     publicMa.String(),
	}

	marshaled, err := json.Marshal(joinInfo)
	if err != nil {
		api.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	encoded := base64.StdEncoding.EncodeToString([]byte(marshaled))

	commRes := &ctl.CommunityCreateResponse{
		CommunityID: community.CommunityID,
		JoinToken:   encoded,
	}

	api.WriteResponse(w, commRes)
}

// TODO: make request args a map
// for now: always send [name, swarmkey, bootstrap peer (one only), address, communityid]
func (m *Manager) CommunityJoinHandler(w http.ResponseWriter, r *http.Request) {
	var commReq ctl.CommunityJoinRequest
	err := api.BindPostRequest(r, &commReq)
	if err != nil {
		api.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	community, err := m.JoinCommunity(commReq.CommunityName, commReq.SwarmKey, commReq.BootstrapPeers, commReq.AcceptingNodeAddr, commReq.CommunityID)
	if err != nil {
		api.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	publicMa, err := compass.PublicRaftMultiaddr()
	if err != nil {
		api.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	addInfo := &ctl.AddMemberInfo{
		CommunityID: community.CommunityID,
		Address:     publicMa.String(),
		NodeID:      compass.PeerID().Pretty(),
	}

	marshaled, err := json.Marshal(addInfo)
	if err != nil {
		api.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	encoded := base64.StdEncoding.EncodeToString([]byte(marshaled))

	commRes := &ctl.CommunityJoinResponse{
		AddMemberToken: encoded,
	}

	api.WriteResponse(w, commRes)
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
	commRes := &ctl.CommunityStateResponse{
		State: state,
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
