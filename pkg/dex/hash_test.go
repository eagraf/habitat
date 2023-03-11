package dex

import (
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
