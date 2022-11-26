package sources

import (
	"encoding/json"

	"github.com/qri-io/jsonschema"
)

type SourceID string
type SourceFile struct {
	ID          SourceID          `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Schema      jsonschema.Schema `json:"schema"`
	Data        json.RawMessage   `json:"data"`
}
