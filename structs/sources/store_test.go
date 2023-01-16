package sources

import (
	"os"
	"testing"

	"github.com/qri-io/jsonschema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var tempSchPath string = "schema"

const idRaw = "test-geo"

var geoSch = `
{
	"$id": "test-geo",
	"$schema": "https://json-schema.org/draft/2020-12/schema",
	"title": "Longitude and Latitude Values",
	"description": "A geographical coordinate.",
	"required": [ "latitude", "longitude" ],
	"type": "object",
	"properties": {
	  "latitude": {
		"type": "number",
		"minimum": -90,
		"maximum": 90
	  },
	  "longitude": {
		"type": "number",
		"minimum": -180,
		"maximum": 180
	  }
	}
  }`

var geoSchema = &Schema{
	Schema:      jsonschema.Must(geoSch),
	ID:          idRaw,
	Name:        "geography",
	Description: "test json schema",
}

func TestSchemaId(t *testing.T) {
	assert.Equal(t, idRaw, GetSchemaIdRaw(geoSchema.Schema))
}

func TestSchemaLookupEmpty(t *testing.T) {
	defer os.RemoveAll(tempSchPath)
	sr := NewLocalSchemaStore(tempSchPath)
	sch, err := sr.Get(idRaw)
	// both schema and error are nil
	assert.Nil(t, sch)
	assert.Nil(t, err)
}

func TestSchemaAdd(t *testing.T) {
	defer os.RemoveAll(tempSchPath)
	sr := NewLocalSchemaStore(tempSchPath)
	err := sr.Add(geoSchema)
	require.Nil(t, err)
	sch, err := sr.Get(idRaw)
	require.Nil(t, err)
	require.NotNil(t, sch)
	assert.Equal(t, *geoSchema, *sch)
}

func TestSchemaDelete(t *testing.T) {
	defer os.RemoveAll(tempSchPath)
	sr := NewLocalSchemaStore(tempSchPath)
	err := sr.Add(geoSchema)
	assert.Nil(t, err)
	sch, err := sr.Get(geoSchema.ID)
	assert.Equal(t, geoSchema, sch)
	assert.Nil(t, err)
	err = sr.Delete(geoSchema.ID)
	assert.Nil(t, err)
	sch, err = sr.Get(geoSchema.ID)
	assert.Nil(t, err)
	assert.Nil(t, sch)
}
