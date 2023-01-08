package dataproxy

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"testing"
	"time"

	"github.com/eagraf/habitat/cmd/sources"
	"github.com/eagraf/habitat/structs/ctl"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const id = "https://json-schema.org/learn/examples/geographical-location.schema.json"

func TestSourcesWriteRead(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	p := NewDataProxy(ctx, nil, map[string]*DataServerNode{})

	path := t.TempDir()

	p.localSourcesHandler = sources.NewJSONReaderWriter(ctx, filepath.Join(path, "sources"))
	p.schemaStore = sources.NewLocalSchemaStore(filepath.Join(path, "schema"))
	// require.Nil(t, p.schemaStore.Add(geoSchema))

	addr := "0.0.0.0:8765"
	go p.Serve(ctx, addr)

	// TODO: find a better way to wait for server to startup
	time.Sleep(1 * time.Second)

	data := `{"latitude":48,"longitude":91}`

	sourcereq := sources.SourceRequest{
		ID: id,
	}
	b, err := json.Marshal(sourcereq)
	require.Nil(t, err)

	req := ctl.DataWriteRequest{
		Type: ctl.SourcesRequest,
		Body: json.RawMessage(b),
		Data: []byte(data),
	}

	b2, err := json.Marshal(req)
	require.Nil(t, err)

	rsp, err := http.Post("http://"+addr+"/write_source", "application/json", bytes.NewReader(b2))
	require.Nil(t, err)

	slurp, err := ioutil.ReadAll(rsp.Body)
	require.Nil(t, err)

	var res ctl.DataWriteResponse
	err = json.Unmarshal(slurp, &res)
	require.Nil(t, err)

	sourcereq = sources.SourceRequest{
		ID: id,
	}
	b, err = json.Marshal(sourcereq)
	require.Nil(t, err)

	rreq := ctl.DataReadRequest{
		Type: ctl.SourcesRequest,
		Body: json.RawMessage(b),
	}

	b2, err = json.Marshal(rreq)
	require.Nil(t, err)

	rsp, err = http.Post("http://"+addr+"/read_source", "application/json", bytes.NewReader(b2))
	require.Nil(t, err)

	slurp, err = ioutil.ReadAll(rsp.Body)
	assert.Nil(t, err)

	var rres ctl.DataReadResponse
	err = json.Unmarshal(slurp, &rres)
	assert.Nil(t, err)

	assert.Equal(t, data, string(rres.Data))
}
