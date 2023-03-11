package dex

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSchemaType(t *testing.T) {
	s := Schema(`{"abc":"xyz"}`)
	m, err := json.Marshal(s)
	assert.Nil(t, err)
	assert.Equal(t, `{"abc":"xyz"}`, string(m))

	var res Schema
	err = json.Unmarshal(m, &res)
	assert.Nil(t, err)
	assert.Equal(t, `{"abc":"xyz"}`, string(res))
}
