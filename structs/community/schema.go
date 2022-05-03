package community

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
				"peer_id": {
					"type": "string"
				},
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
	PeerID             string   `json:"peer_id"`
	BootstrapAddresses []string `json:"bootstrap_addresses"`
}
