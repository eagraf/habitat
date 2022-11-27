package sources

import (
	"os"
	"testing"

	assert "github.com/stretchr/testify/assert"
)

func TestSchemaId(t *testing.T) {
	assert.Equal(t, "test-geo", GetSchemaId(geoSchema))
}

var tempSchPath string = "schema"

func TestSchemaLookupEmpty(t *testing.T) {
	defer os.RemoveAll(tempSchPath)
	sr := NewSchemaRegistry(tempSchPath)
	sch, err := sr.Lookup("test-geo")
	assert.Nil(t, sch)
	assert.Nil(t, err)
}

func TestSchemaAdd(t *testing.T) {
	defer os.RemoveAll(tempSchPath)
	sr := NewSchemaRegistry(tempSchPath)
	err := sr.Add("test-geo", geoSchema)
	assert.Nil(t, err)
	sch, err := sr.Lookup("test-geo")
	assert.Equal(t, geoSchema, sch)
	assert.Nil(t, err)
}

func TestSchemaDelete(t *testing.T) {
	defer os.RemoveAll(tempSchPath)
	sr := NewSchemaRegistry(tempSchPath)
	err := sr.Add("test-geo", geoSchema)
	assert.Nil(t, err)
	sch, err := sr.Lookup("test-geo")
	assert.Equal(t, geoSchema, sch)
	assert.Nil(t, err)
	err = sr.Delete("test-geo")
	assert.Nil(t, err)
	sch, err = sr.Lookup("test-geo")
	assert.Nil(t, sch)
	assert.Nil(t, err)
}
