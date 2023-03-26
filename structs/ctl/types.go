package ctl

import (
	"encoding/json"

	"github.com/eagraf/habitat/structs/community"
	"github.com/qri-io/jsonschema"
)

type InspectRequest struct{}

type InspectResponse struct {
	LibP2PPeerID         string `json:"libp2p_peer_id"`
	LibP2PProxyMultiaddr string `json:"libp2p_proxy_multiaddr"`
}

type StartRequest struct {
	App         string   `json:"app"`
	CommunityID string   `json:"community_id"`
	Args        []string `json:"args"`
	Env         []string `json:"env"`
}

type StartResponse struct {
	ProcessInstanceID string `json:"process_instance_id"`
}

type StopRequest struct {
	ProcID string `json:"process_id"`
}

type StopResponse struct {
}

type PSRequest struct {
}

type PSResponse struct {
	ProcIDs []string `json:"procs"`
}

type CommunityCreateRequest struct {
	CommunityName     string `json:"community_name"`
	CreateIPFSCluster bool   `json:"create_ipfs_cluster"`
}

type CommunityCreateResponse struct {
	CommunityID string `json:"community_id"`
	JoinToken   string `json:"join_code"`

	WebsocketControl
}

type CommunityJoinRequest struct {
	CommunityID       string   `json:"community_id"`
	CommunityName     string   `json:"community_name"`
	AcceptingNodeAddr string   `json:"accepting_node_addr"`
	SwarmKey          string   `json:"swarm_key"`
	BootstrapPeers    []string `json:"bootstrap_peers"`
}

type CommunityJoinResponse struct {
	WebsocketControl
}

type CommunityAddMemberRequest struct {
	CommunityID        string            `json:"community_id"`
	NodeID             string            `json:"peer_id"`
	JoiningNodeAddress string            `json:"joining_node_address"`
	Member             *community.Member `json:"member"`
	Node               *community.Node   `json:"node"`
}

type CommunityAddMemberResponse struct {
}

type CommunityProposeRequest struct {
	CommunityID     string `json:"community_id"`
	StateTransition []byte `json:"state_transition"`
}

type CommunityProposeResponse struct {
}

type CommunityStateRequest struct {
	CommunityID string `json:"community_id"`
}

type CommunityStateResponse struct {
	State []byte `json:"community_state"`
}

type CommunityListRequest struct {
}

type CommunityListResponse struct {
	NodeID      string   `json:"node_id"`
	Communities []string `json:"communities"`
}

type CommunityPSRequest struct {
	CommunityID string `json:"community_id"`
}

type CommunityPSProcess struct {
	*community.Process
	Instances []*community.ProcessInstance `json:"instances"`
}

type CommunityPSResponse struct {
	Processes []*CommunityPSProcess `json:"processes"`
}

type CommunityStartProcessRequest struct {
	CommunityID string `json:"community_id"`

	App  string   `json:"app"`
	Args []string `json:"args"`
	Env  []string `json:"env"`

	InstancesNodes []string `json:"instance_nodes"`
}

type CommunityStartProcessResponse struct {
}

type CommunityStopProcessRequest struct {
	CommunityID    string   `json:"community_id"`
	ProcessID      string   `json:"process_id"`
	InstancesNodes []string `json:"instance_nodes"`
}

type CommunityStopProcessResponse struct {
}

// The following message types are for key signing exchanges

type JoinInfo struct {
	CommunityID string `json:"community_id"`
	Address     string `json:"address"`
}

// Sent by server back to client
type SigningPublicKeyMsg struct {
	NodeID    string `json:"node_id"`
	PublicKey []byte `json:"public_key"`

	WebsocketControl
}

// Sent by client back to server
type SigningCertMsg struct {
	UserCertificate []byte `json:"user_certificate"`
	NodeCertificate []byte `json:"node_certificate"`

	WebsocketControl
}

type dataType string

const (
	SourcesRequest dataType = "sources"
)

type DataReadRequest struct {
	Type        dataType        `json:"data_type"`
	NodeID      string          `json:"node_id"`
	CommunityID string          `json:"community_id"`
	Token       string          `json:"token"`
	Body        json.RawMessage `json:"body"`
}

type DataReadResponse struct {
	Data []byte `json:"data"`
}

type DataWriteRequest struct {
	Type        dataType        `json:"data_type"`
	NodeID      string          `json:"node_id"`
	CommunityID string          `json:"community_id"`
	Token       string          `json:"token"`
	Body        json.RawMessage `json:"body"`
	Data        []byte          `json:"data"`
}

type DataWriteResponse struct{}

type AddSchemaRequest struct {
	Schema *jsonschema.Schema `json:"schema"`
}

type LookupSchemaRequest struct {
	ID string `json:"id"`
}

type DeleteSchemaRequest struct {
	ID string `json:"id"`
}

type AddFileRequest struct {
}

type AddFileResponse struct {
	ContentID string `json:"content_id"`
}

type GetFileRequest struct {
	ContentID string `json:"content_id"`
}

type GetFileResponse struct {
}
