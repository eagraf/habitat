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
