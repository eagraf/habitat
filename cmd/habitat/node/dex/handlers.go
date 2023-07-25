package dex

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/eagraf/habitat/cmd/habitat/api"
	"github.com/eagraf/habitat/pkg/ipfs"
	"github.com/eagraf/habitat/structs/ctl"
)

type DexNodeAPI struct {
	ipfsClient *ipfs.Client
}

func NewDexNodeAPI(ipfsClient *ipfs.Client) *DexNodeAPI {
	return &DexNodeAPI{
		ipfsClient: ipfsClient,
	}
}

func (d *DexNodeAPI) GetSchemaHandler(w http.ResponseWriter, r *http.Request) {
	// Unmarshal the request body into the `SchemaRequest` struct
	var req ctl.SchemaRequest
	err := api.BindPostRequest(r, &req)
	if err != nil {
		api.WriteError(w, http.StatusBadRequest, err)
		return
	}

	// Fetch the schema file from IPFS based on the provided hash
	schemaPath := schemaPath(req.Hash)
	schemaData, exists, err := d.ipfsClient.ReadFile(schemaPath)
	if err != nil {
		api.WriteError(w, http.StatusInternalServerError, err)
		return
	}
	if !exists {
		// TODO cache miss, should hit global service when that exists
		api.WriteError(w, http.StatusNotFound, fmt.Errorf("hash %s not found", req.Hash))
		return
	}

	buf, err := io.ReadAll(schemaData)
	if err != nil {
		api.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	// Create a `SchemaResponse` struct with the hash and fetched schema
	result := ctl.SchemaResponse{
		Hash:   req.Hash,
		Schema: ctl.Schema(buf),
	}

	// Write the response as JSON
	api.WriteResponse(w, result)
}

func (d *DexNodeAPI) WriteSchemaHandler(w http.ResponseWriter, r *http.Request) {
	var req ctl.SchemaWriteRequest
	err := api.BindPostRequest(r, &req)
	if err != nil {
		api.WriteError(w, http.StatusBadRequest, err)
		return
	}

	hash, err := req.Schema.Hash()
	if err != nil {
		api.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	// Write the schema to IPFS
	reader := bytes.NewReader(req.Schema)
	schemaPath := schemaPath(hash)
	_, err = d.ipfsClient.WriteFile(schemaPath, "schema", reader)
	if err != nil {
		api.WriteError(w, http.StatusInternalServerError, err)
		return
	}
	res := ctl.SchemaWriteResponse{
		Hash: hash,
	}
	api.WriteResponse(w, res)
}

func (d *DexNodeAPI) GetInterfaceHandler(w http.ResponseWriter, r *http.Request) {
	// Unmarshal the request body into the `InterfaceRequest` struct
	var req ctl.InterfaceRequest
	err := api.BindPostRequest(r, &req)
	if err != nil {
		api.WriteError(w, http.StatusBadRequest, err)
		return
	}

	// Fetch the interface file from IPFS based on the provided hash
	interfacePath := interfacePath(req.Hash)
	reader, exists, err := d.ipfsClient.ReadFile(interfacePath)
	if err != nil {
		api.WriteError(w, http.StatusInternalServerError, err)
		return
	}
	if !exists {
		// TODO cache miss, should hit global service when that exists
		api.WriteError(w, http.StatusNotFound, fmt.Errorf("hash %s not found", req.Hash))
		return
	}

	// Read the interface data from the reader
	interfaceData, err := ioutil.ReadAll(reader)
	if err != nil {
		api.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	// Create a `ctl.Interface` struct to hold the unmarshaled data
	var iface ctl.Interface

	// Unmarshal the interfaceData into the `ctl.Interface` struct
	err = json.Unmarshal(interfaceData, &iface)
	if err != nil {
		api.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	// Create a `InterfaceResponse` struct with the hash and unmarshaled interface
	result := ctl.InterfaceResponse{
		Hash:      req.Hash,
		Interface: &iface,
	}

	// Write the response as JSON
	api.WriteResponse(w, result)
}

func (d *DexNodeAPI) WriteInterfaceHandler(w http.ResponseWriter, r *http.Request) {
	var req ctl.InterfaceWriteRequest
	err := api.BindPostRequest(r, &req)
	if err != nil {
		api.WriteError(w, http.StatusBadRequest, err)
		return
	}

	hash, err := req.Interface.Hash()
	if err != nil {
		api.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	// Write the schema to IPFS
	marshaled, err := json.Marshal(req.Interface)
	if err != nil {
		api.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	reader := bytes.NewReader(marshaled)
	interfacePath := interfacePath(hash)
	_, err = d.ipfsClient.WriteFile(interfacePath, "interface", reader)
	if err != nil {
		api.WriteError(w, http.StatusInternalServerError, err)
		return
	}
	res := ctl.InterfaceWriteResponse{
		Hash: hash,
	}
	api.WriteResponse(w, res)
}

func schemaPath(hash string) string {
	return fmt.Sprintf("/schema/%s", hash)
}

func interfacePath(hash string) string {
	return fmt.Sprintf("/interface/%s", hash)
}
