package sources

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/qri-io/jsonschema"
	assert "github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
var geoData = json.RawMessage(`{"latitude":45,"longitude":45}`)
var geoSource = &Source{
	Data: json.RawMessage(geoData),
}

const idRaw = "test-geo"

var geoSchema = &Schema{
	Schema:      jsonschema.Must(geoSch),
	ID:          idRaw,
	Name:        "geography",
	Description: "test json schema",
}

// Schema Tests

func TestSchemaId(t *testing.T) {
	assert.Equal(t, idRaw, GetSchemaIdRaw(geoSchema))
}

var tempSchPath string = "schema"

func init() {
	os.RemoveAll(tempSchPath)
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

var readerwriter *JSONReaderWriter

func setupReaderWriter() {
	readerwriter = NewJSONReaderWriter(context.Background(), ".")
}

func setupSource(json string, path string) {
	os.WriteFile(path, []byte(json), 0644)
}

func teardownSource(path string) {
	os.RemoveAll(path)
}

func getSourceRaw(path string) string {
	res, _ := os.ReadFile(path)
	return string(res)
}
func TestBasicReadWrite(t *testing.T) {
	setupReaderWriter()
	id := "test-geo"
	sourcePath := getSourcePath(".", id)
	setupSource(string(geoSource.Data), sourcePath)
	defer teardownSource(sourcePath)

	assert.Equal(t, getSourceRaw(sourcePath), `{"latitude":45,"longitude":45}`)

	data, err := readerwriter.Read(id)
	require.Nil(t, err)
	assert.Equal(t, string(data), string(geoSource.Data))

	err = readerwriter.Write(id, geoSchema, []byte(`{"latitude":9,"longitude":90}`))
	assert.Nil(t, err)

	assert.Equal(t, getSourceRaw(sourcePath), `{"latitude":9,"longitude":90}`)

	err = readerwriter.Write(id, geoSchema, []byte(`{"latitude":-100,"longitude":90}`))
	assert.NotNil(t, err)

	// same as before
	assert.Equal(t, getSourceRaw(sourcePath), `{"latitude":9,"longitude":90}`)

}
