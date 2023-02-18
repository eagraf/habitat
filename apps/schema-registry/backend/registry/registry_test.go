package registry

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/eagraf/habitat/structs/ctl"
	"github.com/eagraf/habitat/structs/sources"
	"github.com/qri-io/jsonschema"
	"github.com/stretchr/testify/assert"
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

const idRaw = "test-geo"

var geoSchema = &sources.Schema{
	Schema:      jsonschema.Must(geoSch),
	ID:          idRaw,
	Name:        "geography",
	Description: "test json schema",
}

func TestBasicSchemaRegistry(t *testing.T) {
	reg := NewRegistry(
		sources.NewLocalSchemaStore(
			t.TempDir(),
		),
	)

	closeCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go reg.Serve(closeCtx, "1234")

	time.Sleep(1 * time.Second)

	addReq := ctl.AddSchemaRequest{
		Sch: geoSchema,
	}
	reqJson, err := json.Marshal(addReq)
	require.Nil(t, err)

	reader := bytes.NewReader(reqJson)

	resp, err := http.DefaultClient.Post("http://0.0.0.0:1234/add_schema", "application/json", reader)
	uerr, ok := err.(*url.Error)
	if ok {
		fmt.Println(uerr.Err.Error())
	}
	require.Nil(t, err)

	slurp, err := ioutil.ReadAll(resp.Body)

	var addRes ctl.AddSchemaResponse
	err = json.Unmarshal(slurp, &addRes)
	require.Nil(t, err)
	assert.Equal(t, addRes.ID, idRaw)

	getReq := ctl.GetSchemaRequest{
		Id: idRaw,
	}

	reqJson, err = json.Marshal(getReq)
	require.Nil(t, err)

	reader = bytes.NewReader(reqJson)

	resp, err = http.DefaultClient.Post("http://0.0.0.0:1234/get_schema", "application/json", reader)
	uerr, ok = err.(*url.Error)
	if ok {
		fmt.Println(uerr.Err.Error())
	}
	require.Nil(t, err)

	slurp, err = ioutil.ReadAll(resp.Body)

	var getRes ctl.GetSchemaResponse
	err = json.Unmarshal(slurp, &getRes)
	require.Nil(t, err)
	assert.Equal(t, getRes.Sch, geoSchema)

	delReq := ctl.DeleteSchemaRequest{
		ID: idRaw,
	}
	reqJson, err = json.Marshal(delReq)
	require.Nil(t, err)

	reader = bytes.NewReader(reqJson)

	resp, err = http.DefaultClient.Post("http://0.0.0.0:1234/delete_schema", "application/json", reader)
	uerr, ok = err.(*url.Error)
	if ok {
		fmt.Println(uerr.Err.Error())
	}
	require.Nil(t, err)

	slurp, err = ioutil.ReadAll(resp.Body)

	var delRes ctl.DeleteSchemaResponse
	err = json.Unmarshal(slurp, &delRes)
	require.Nil(t, err)

	reqJson, err = json.Marshal(getReq)
	require.Nil(t, err)

	reader = bytes.NewReader(reqJson)

	resp, err = http.DefaultClient.Post("http://0.0.0.0:1234/get_schema", "application/json", reader)
	uerr, ok = err.(*url.Error)
	if ok {
		fmt.Println(uerr.Err.Error())
	}
	require.Nil(t, err)

	slurp, err = ioutil.ReadAll(resp.Body)

	err = json.Unmarshal(slurp, &getRes)
	require.Nil(t, err)
	assert.Equal(t, getRes.Sch, (*sources.Schema)(nil))
}
