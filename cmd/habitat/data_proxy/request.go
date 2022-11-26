package dataproxy

type SourcesRequest struct {
	sourceID   string `json:"sourceID,omitempty"`
	schemaHash string `json:"schemaHash,omitempty"`
}
