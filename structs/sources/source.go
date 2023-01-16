package sources

import (
	"encoding/json"
)

type Source struct {
	Data json.RawMessage `json:"data"`
}

type SourceRequest struct {
	ID string `json:"sourceID,omitempty"`
}

type SchemaStore interface {
	Add(*Schema) error
	Get(string) (*Schema, error)
	Delete(string) error
}
