package sources

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

var geographyJSON = `{"latitude":45,"longitude":45}`
var reader *Reader
var writer *Writer

func setupReaderWriter() {
	sreader := NewJSONReader(".")
	swriter := NewJSONWriter(".")
	pmanager := NewBasicPermissionsManager()
	reader = NewReader(sreader, pmanager)
	writer = NewWriter(swriter, pmanager)
}

func setupSource(json string, path string) {
	os.WriteFile(path, []byte(json), 0600)
}

func teardownSource(path string) {
	os.Remove(path)
}

func getSourceRaw(path string) string {
	res, _ := os.ReadFile(path)
	return string(res)
}
func TestBasicReadWrite(t *testing.T) {
	setupReaderWriter()
	sourcePath := "./test-geo.json"
	setupSource(geographyJSON, sourcePath)
	defer teardownSource(sourcePath)

	readReq := []byte(`{
		"requester": "arushi",
		"source": {
		  "name": "test-geo",
		  "description": "hi",
		  "schema": {
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
		  }
		}
	  }
	  `)
	rreq := &ReadRequest{}
	json.Unmarshal(readReq, rreq)

	allowed, err, data := reader.Read(rreq)
	assert.True(t, allowed)
	assert.Nil(t, err)
	assert.Equal(t, string(data), geographyJSON)

	writeReq := []byte(`{
		"requester": "arushi",
		"source": {
		  "name": "test-geo",
		  "description": "hi",
		  "schema": {
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
		  }
		},
		"data": "{\"latitude\":90,\"longitude\":90}"
	  }`)

	wreq := &WriteRequest{}
	json.Unmarshal(writeReq, wreq)

	allowed, err = writer.Write(wreq)
	assert.True(t, allowed)
	assert.Nil(t, err)

	assert.Equal(t, getSourceRaw(sourcePath), `{"latitude":90,"longitude":90}`)

	// invalid write
	writeReq2 := []byte(`{
		"requester": "arushi",
		"source": {
		  "name": "test-geo",
		  "description": "hi",
		  "schema": {
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
		  }
		},
		"data": "{\"latitude\":-100,\"longitude\":90}"
	  }`)

	wreq2 := &WriteRequest{}
	json.Unmarshal(writeReq2, wreq2)

	allowed, err = writer.Write(wreq2)
	assert.True(t, allowed)
	assert.NotNil(t, err)

	// same as before
	assert.Equal(t, getSourceRaw(sourcePath), `{"latitude":90,"longitude":90}`)

}
