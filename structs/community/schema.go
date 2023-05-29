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
				"p2p_id": { "type": "string" },
				"address": { "type": "string" },
				"ipfs_swarm_address": { "type": "string" },
				"certificate": { "type": "string" },
				"member_id": { "type": "string" }
			},
			"required": [ "id", "address", "ipfs_swarm_address", "certificate", "member_id" ]
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
				},
				"is_datastore": {
					"type": "boolean"
				}
			},
			"required": [ "id", "app_name", "env", "flags", "args", "is_datastore" ]
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

type Node struct {
	ID               string `json:"id"`
	P2PID            string `json:"p2p_id"`
	Address          string `json:"address"`
	IPFSSwarmAddress string `json:"ipfs_swarm_address"` // This should be temporary
	Certificate      []byte `json:"certificate"`
	MemberID         string `json:"member_id"`
}

type Process struct {
	ID string `json:"id"`

	AppName string   `json:"app_name"`
	Env     []string `json:"env"`
	Flags   []string `json:"flags"`
	Args    []string `json:"args"`

	Config interface{} `json:"config"`

	IsDatastore bool `json:"is_datastore"`
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
