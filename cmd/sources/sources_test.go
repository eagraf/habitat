package sources

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/qri-io/jsonschema"
	assert "github.com/stretchr/testify/assert"
)

var geoData = json.RawMessage(`{"latitude":45,"longitude":45}`)
var geoSource = &SourceFile{
	ID:          "test-geo",
	Name:        "geography",
	Description: "test json schema",
	Schema: *jsonschema.Must(`
	{
		"$id": "https://example.com/geographical-location.schema.json",
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
	  }`),
	Data: json.RawMessage(geoData),
}

var reader *JSONReader
var writer *JSONWriter

func setupReaderWriter() {
	reader = NewJSONReader(context.Background(), ".")
	writer = NewJSONWriter(context.Background(), ".")
}

func setupSource(json string, path string) {
	os.WriteFile(path, []byte(json), 0644)
}

func teardownSource(path string) {
	// os.Remove(path)
}

func getSourceRaw(path string) string {
	res, _ := os.ReadFile(path)
	return string(res)
}
func TestBasicReadWrite(t *testing.T) {
	setupReaderWriter()
	sourcePath := "./test-geo.json"
	bytes, err := json.Marshal(geoSource)
	assert.Nil(t, err)
	setupSource(string(bytes), sourcePath)
	defer teardownSource(sourcePath)

	assert.Equal(t, getSourceRaw(sourcePath), `{"id":"test-geo","name":"geography","description":"test json schema","schema":{"$id":"https://example.com/geographical-location.schema.json","$schema":"https://json-schema.org/draft/2020-12/schema","description":"A geographical coordinate.","properties":{"latitude":{"maximum":90,"minimum":-90,"type":"number"},"longitude":{"maximum":180,"minimum":-180,"type":"number"}},"required":["latitude","longitude"],"title":"Longitude and Latitude Values","type":"object"},"data":{"latitude":45,"longitude":45}}`)

	data, err := reader.Read("test-geo")
	assert.Nil(t, err)
	geoDataBytes, err := json.Marshal(geoData)
	assert.Nil(t, err)
	assert.Equal(t, string(data), string(geoDataBytes))

	err = writer.Write("test-geo", []byte(`{"latitude":9,"longitude":90}`))
	assert.Nil(t, err)

	assert.Equal(t, getSourceRaw(sourcePath), `{"id":"test-geo","name":"geography","description":"test json schema","schema":{"$id":"https://example.com/geographical-location.schema.json","$schema":"https://json-schema.org/draft/2020-12/schema","description":"A geographical coordinate.","properties":{"latitude":{"maximum":90,"minimum":-90,"type":"number"},"longitude":{"maximum":180,"minimum":-180,"type":"number"}},"required":["latitude","longitude"],"title":"Longitude and Latitude Values","type":"object"},"data":{"latitude":9,"longitude":90}}`)

	err = writer.Write("test-geo", []byte(`{"latitude":-100,"longitude":90}`))
	assert.NotNil(t, err)

	// same as before
	assert.Equal(t, getSourceRaw(sourcePath), `{"id":"test-geo","name":"geography","description":"test json schema","schema":{"$id":"https://example.com/geographical-location.schema.json","$schema":"https://json-schema.org/draft/2020-12/schema","description":"A geographical coordinate.","properties":{"latitude":{"maximum":90,"minimum":-90,"type":"number"},"longitude":{"maximum":180,"minimum":-180,"type":"number"}},"required":["latitude","longitude"],"title":"Longitude and Latitude Values","type":"object"},"data":{"latitude":9,"longitude":90}}`)

}
