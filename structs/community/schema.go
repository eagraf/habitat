package community

import (
	"encoding/json"

	"github.com/libp2p/go-libp2p/core/peer"
	ma "github.com/multiformats/go-multiaddr"
)

var CommunityStateSchema = []byte(`{
	"$defs": {
		"member": {
			"type": "object",
			"properties": {
				"id": { "type": "string" },
				"username": { "type": "string" },
				"certificate": { "type": "string" }
			},
			"required": [ "username", "certificate" ]
		},
		"node": {
			"type": "object",
			"properties": {
				"id": { "type": "string" },
				"peer_id": { "type": "string" },
				"addresses": {
					"type": "array",
					"items": { "type": "string" }
				},
				"certificate": { "type": "string" },
				"member_id": { "type": "string" },
				"reachability": { "type": "string" },
				"relay": { "type": "boolean" }
			},
			"required": [ "id", "peer_id", "addresses", "certificate", "member_id", "reachability", "relay" ]
		},
		"process": {
			"type": "object",
			"properties": {
				"id": { "type": "string" },
				"app_name": { "type": "string" },
				"env": {
					"type": "array",
					"items": { "type": "string" }
				},
				"flags": {
					"type": "array",
					"items": { "type": "string" }
				},
				"args": {
					"type": "array",
					"items": { "type": "string" }
				},
				"config": {
					"type": [ "null", "object" ]
				}
			},
			"required": [ "id", "app_name", "env", "flags", "args" ]
		},
		"process_instance": {
			"type": "object",
			"properties": {
				"process_id": { "type": "string" },
				"node_id": { "type": "string" }
			},
			"required": [ "process_id", "node_id" ]
		}
	},
	"title": "community state schema",
	"type": "object",
	"properties": {
		"community_id": {
			"type": "string"
		},
		"counter": {
			"type": "integer",
			"minmum": 0
		},
		"ipfs_config": {
			"type": [ "object", "null" ],
			"properties": {
				"swarm_key": {
					"type": "string"
				},
				"bootstrap_addresses": {
					"type": "array",
					"items": {
						"type": "string"
					}
				}
			}
		},
		"members": {
			"type": "array",
			"items": {
				"$ref": "#/$defs/member"
			}
		},
		"nodes": {
			"type": "array",
			"items": {
				"$ref": "#/$defs/node"
			}
		},
		"processes": {
			"type": "array",
			"items": {
				"$ref": "#/$defs/process"
			}
		},
		"process_instances": {
			"type": "array",
			"items": {
				"$ref": "#/$defs/process_instance"
			}
		}
	},
	"required": [ "community_id", "members", "nodes", "processes", "process_instances" ]
}`)

// CommunityState is a Go struct that correspons to the community state JSON schema
// TODO look at ways to generate this from the schema or vice versa so there is a single
// source of truth
type CommunityState struct {
	CommunityID      string             `json:"community_id"`
	Counter          int                `json:"counter,omitempty"`
	IPFSConfig       *IPFSConfig        `json:"ipfs_config"`
	Members          []*Member          `json:"members"`
	Nodes            []*Node            `json:"nodes"`
	Processes        []*Process         `json:"processes"`
	ProcessInstances []*ProcessInstance `json:"process_instances"`
}

type IPFSConfig struct {
	SwarmKey           string   `json:"swarm_key"`
	BootstrapAddresses []string `json:"bootstrap_addresses"`
}

type Member struct {
	ID          string `json:"id"`
	Username    string `json:"username"`
	Certificate []byte `json:"certificate"`
}

const (
	ReachabilityUnknown = "unknown"
	ReachabilityPublic  = "public"
	ReachabilityPrivate = "private"
)

type Node struct {
	ID           string   `json:"id"`
	PeerID       string   `json:"peer_id"`
	Addresses    []string `json:"addresses"`
	Certificate  []byte   `json:"certificate"`
	MemberID     string   `json:"member_id"`
	Reachability string   `json:"reachability"`
	Relay        bool     `json:"relay"`
}

func (n *Node) DecodedPeerID() (peer.ID, error) {
	return peer.Decode(n.PeerID)
}

func (n *Node) AddrInfo() (*peer.AddrInfo, error) {
	peerID, err := n.DecodedPeerID()
	if err != nil {
		return nil, err
	}
	res := &peer.AddrInfo{
		ID:    peerID,
		Addrs: []ma.Multiaddr{},
	}

	for _, a := range n.Addresses {
		addr, err := ma.NewMultiaddr(a)
		if err != nil {
			return nil, err
		}
		res.Addrs = append(res.Addrs, addr)
	}

	return res, nil
}

type Process struct {
	ID string `json:"id"`

	AppName string   `json:"app_name"`
	Env     []string `json:"env"`
	Flags   []string `json:"flags"`
	Args    []string `json:"args"`

	Config interface{} `json:"config"`
}

type ProcessInstance struct {
	ProcessID string `json:"process_id"`
	NodeID    string `json:"node_id"`
}

func NewCommunityState() *CommunityState {
	return &CommunityState{
		Members:          []*Member{},
		Nodes:            []*Node{},
		Processes:        []*Process{},
		ProcessInstances: []*ProcessInstance{},
	}
}

func NewCommunityStateBytes() []byte {
	state := NewCommunityState()
	res, err := json.Marshal(state)
	if err != nil {
		panic("this panic should be unreachable")
	}
	return res
}
