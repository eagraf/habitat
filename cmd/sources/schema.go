package sources

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/qri-io/jsonschema"
)

/*
Basic types about schema and sources, shared types.
*/

type SourceName string
type SourceFile struct {
	ID          SourceName        `json:"name"`
	Description string            `json:"description"`
	Schema      jsonschema.Schema `json:"schema"`
	Data        json.RawMessage   `json:"data"`
}

// unused right now, use later for token
type RequestToken struct {
	Token string `json:"token"`
}

func (s *SourceFile) ValidateDataAgainstSchema(ctx context.Context, data []byte) error {
	kerr, err := s.Schema.ValidateBytes(ctx, data)
	es := make([]string, len(kerr))
	for i, e := range kerr {
		es[i] = e.Error()
	}
	if len(kerr) > 0 {
		return fmt.Errorf("key errors: %s", strings.Join(es, ","))
	}
	return err
}

// TODO: NewSchema()
// functions for evolving schema
/*
func NewSchema(schema []byte) (error, jsonschema.Schema) {
	rs := &jsonschema.Schema{}
	if err := json.Unmarshal(schema, rs); err != nil {
		panic("unmarshal schema: " + err.Error())
	}
}
*/
