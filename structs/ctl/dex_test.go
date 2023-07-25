package ctl

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHash(t *testing.T) {
	h := hash([]byte("abc"))
	assert.Equal(t, 40, len(h))

	schema := Schema([]byte("{}"))
	h, err := schema.Hash()
	assert.Nil(t, err)
	assert.Equal(t, 40, len(h))

	inter := &Interface{
		SchemaHash:  "abc",
		Description: "desc",
	}
	h, err = inter.Hash()
	assert.Nil(t, err)
	assert.Equal(t, 40, len(h))
}

// TestSchemaJSONRawMessage ensures that marshaling/unmarshaling the json.RawMessage
// in the Schema type works. The test is extremely simple, but getting it to work was a
// bit tricky and required a custom unmarshal for schemas.
func TestSchemaJSONRawMessage(t *testing.T) {
	s := Schema(`{"abc":"xyz"}`)
	m, err := json.Marshal(s)
	assert.Nil(t, err)
	assert.Equal(t, `{"abc":"xyz"}`, string(m))

	var res Schema
	err = json.Unmarshal(m, &res)
	assert.Nil(t, err)
	assert.Equal(t, `{"abc":"xyz"}`, string(res))
}
