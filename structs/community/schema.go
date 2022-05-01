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
		}
	},
	"required": ["community_id"]
}`)

// CommunityState is a Go struct that correspons to the community state JSON schema
// TODO look at ways to generate this from the schema or vice versa so there is a single
// source of truth
type CommunityState struct {
	CommunityID string `json:"community_id"`
	Counter     int    `json:"counter,omitempty"`
}
