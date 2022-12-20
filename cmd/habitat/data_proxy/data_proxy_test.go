package dataproxy

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/eagraf/habitat/cmd/sources"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const rawid = "https://json-schema.org/learn/examples/geographical-location.schema.json"

var geoSch = `
{
	"$id": "https://json-schema.org/learn/examples/geographical-location.schema.json",
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

var id = sources.EncodeId(rawid)
var geoSchema = &sources.Schema{
	Schema:      []byte(geoSch),
	B64id:       id,
	Name:        "geography",
	Description: "test json schema",
}

func TestSourcesWriteRead(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	p := NewDataProxy(ctx, map[string]*DataServerNode{})

	path := "tmp"
	os.MkdirAll("tmp", os.ModePerm)
	defer os.RemoveAll("tmp")

	p.localSourcesHandler = sources.NewJSONReaderWriter(ctx, filepath.Join(path, "sources"))
	p.schemaStore = sources.NewLocalSchemaStore(filepath.Join(path, "schema"))
	p.schemaStore.Add(geoSchema)

	addr := "0.0.0.0:8765"
	go p.Serve(ctx, addr)
	time.Sleep(1 * time.Second)

	data := `{"latitude":48,"longitude":90}`

	sourcereq := sources.SourceRequest{
		Id: id,
	}
	b, err := json.Marshal(sourcereq)
	require.Nil(t, err)

	req := WriteRequest{
		T:    SourcesRequest,
		Body: json.RawMessage(b),
		Data: []byte(data),
	}

	b2, err := json.Marshal(req)
	require.Nil(t, err)

	rsp, err := http.Post("http://"+addr+"/write", "application/json", bytes.NewReader(b2))
	require.Nil(t, err)

	slurp, err := ioutil.ReadAll(rsp.Body)
	require.Nil(t, err)
	assert.Equal(t, "success!", string(slurp))

	sourcereq = sources.SourceRequest{
		Id: id,
	}
	b, err = json.Marshal(sourcereq)
	require.Nil(t, err)

	rreq := ReadRequest{
		T:    SourcesRequest,
		Body: json.RawMessage(b),
	}

	b2, err = json.Marshal(rreq)
	require.Nil(t, err)

	rsp, err = http.Post("http://"+addr+"/read", "application/json", bytes.NewReader(b2))
	require.Nil(t, err)

	slurp, err = ioutil.ReadAll(rsp.Body)
	require.Nil(t, err)
	assert.Equal(t, data, string(slurp))
}
