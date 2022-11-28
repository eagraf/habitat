package community

import "encoding/json"

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
				"address": { "type": "string" },
				"certificate": { "type": "string" },
				"member_id": { "type": "string" }
			},
			"required": [ "id", "address", "certificate", "member_id" ]
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
		}
	},
	"required": [ "community_id", "members", "nodes" ]
}`)

// CommunityState is a Go struct that correspons to the community state JSON schema
// TODO look at ways to generate this from the schema or vice versa so there is a single
// source of truth
type CommunityState struct {
	CommunityID string      `json:"community_id"`
	Counter     int         `json:"counter,omitempty"`
	IPFSConfig  *IPFSConfig `json:"ipfs_config"`
	Members     []*Member   `json:"members"`
	Nodes       []*Node     `json:"nodes"`
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

type Node struct {
	ID          string `json:"id"`
	Address     string `json:"address"`
	Certificate []byte `json:"certificate"`
	MemberID    string `json:"member_id"`
}

func NewCommunityState() *CommunityState {
	return &CommunityState{
		Members: []*Member{},
		Nodes:   []*Node{},
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
