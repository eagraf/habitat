package community

import "encoding/json"

var CommunityStateSchema = []byte(`{
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
		}
	},
	"required": ["community_id"]
}`)

// CommunityState is a Go struct that correspons to the community state JSON schema
// TODO look at ways to generate this from the schema or vice versa so there is a single
// source of truth
type CommunityState struct {
	CommunityID string      `json:"community_id"`
	Counter     int         `json:"counter,omitempty"`
	IPFSConfig  *IPFSConfig `json:"ipfs_config"`
}

type IPFSConfig struct {
	SwarmKey           string   `json:"swarm_key"`
	BootstrapAddresses []string `json:"bootstrap_addresses"`
}

func NewCommunityState() *CommunityState {
	return &CommunityState{}
}

func NewCommunityStateBytes() []byte {
	state := NewCommunityState()
	res, err := json.Marshal(state)
	if err != nil {
		panic("this panic should be unreachable")
	}
	return res
}
