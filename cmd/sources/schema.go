package sources

import (
	"context"
	"fmt"
	"strings"
)

/*
Basic types about schema, shared types.
*/

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
