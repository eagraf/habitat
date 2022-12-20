package sources

import (
	"context"
	"encoding/json"
	"os"
	"testing"

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

var geoSchema = &Schema{
	Schema:      []byte(geoSch),
	B64id:       EncodeId("test-geo"),
	Name:        "geography",
	Description: "test json schema",
}

// Schema Tests

func TestSchemaId(t *testing.T) {
	assert.Equal(t, "test-geo", GetSchemaIdRaw(geoSchema))
}

var tempSchPath string = "schema"

func init() {
	os.RemoveAll(tempSchPath)
}

func TestSchemaLookupEmpty(t *testing.T) {
	defer os.RemoveAll(tempSchPath)
	sr := NewLocalSchemaStore(tempSchPath)
	sch, err := sr.Get(geoSchema.B64id)
	assert.Nil(t, sch)
	assert.Equal(t, "stat schema/dGVzdC1nZW8=.json: no such file or directory", err.Error())
}

func TestSchemaAdd(t *testing.T) {
	defer os.RemoveAll(tempSchPath)
	sr := NewLocalSchemaStore(tempSchPath)
	err := sr.Add(geoSchema)
	require.Nil(t, err)
	sch, err := sr.Get(geoSchema.B64id)
	require.Nil(t, err)
	assert.Equal(t, *geoSchema, *sch)
	assert.Nil(t, err)
}

func TestSchemaDelete(t *testing.T) {
	defer os.RemoveAll(tempSchPath)
	sr := NewLocalSchemaStore(tempSchPath)
	err := sr.Add(geoSchema)
	assert.Nil(t, err)
	sch, err := sr.Get(geoSchema.B64id)
	assert.Equal(t, geoSchema, sch)
	assert.Nil(t, err)
	err = sr.Delete(geoSchema.B64id)
	assert.Nil(t, err)
	sch, err = sr.Get(geoSchema.B64id)
	assert.Nil(t, sch)
	assert.Equal(t, "stat schema/dGVzdC1nZW8=.json: no such file or directory", err.Error())
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
	id := EncodeId("test-geo")
	sourcePath := getPath(".", id)
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
