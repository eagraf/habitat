package sources

import (
	"encoding/json"
	"os"
)

type SourceID string
type SourceFile struct {
	ID          SourceID        `json:"id"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Data        json.RawMessage `json:"data"`
}

type SourceRequest struct {
	SourceID   string `json:"sourceID,omitempty"`
	SchemaHash string `json:"schemaHash,omitempty"`
}

func ReadSource(path string) (*SourceFile, error) {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var source SourceFile
	if err = json.Unmarshal(bytes, &source); err != nil {
		return nil, err
	}

	return &source, err
}
