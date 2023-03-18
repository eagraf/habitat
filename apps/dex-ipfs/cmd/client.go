package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	"github.com/eagraf/habitat/pkg/dex"
	"github.com/eagraf/habitat/pkg/ipfs"
)

// Note that this write client writes to IPFS directly. It does not go through the dex driver
type DexIpfsWriteClient struct {
	ipfsClient *ipfs.Client
}

func (c *DexIpfsWriteClient) WriteSchema(hash string, schema dex.Schema) (*dex.SchemaResult, error) {
	r := bytes.NewBuffer(schema)
	_, err := c.ipfsClient.WriteFile(schemaPath(hash), "schema", r)
	if err != nil {
		return nil, err
	}
	return &dex.SchemaResult{Hash: hash, Schema: schema}, nil
}

func (c *DexIpfsWriteClient) WriteInterface(hash string, iface *dex.Interface) (*dex.InterfaceResult, error) {

	// stat the path first

	marshaled, err := json.Marshal(iface)
	if err != nil {
		return nil, err
	}
	r := bytes.NewBuffer(marshaled)
	_, err = c.ipfsClient.WriteFile(interfacePath(hash), "interface", r)
	if err != nil {
		return nil, err
	}
	return &dex.InterfaceResult{Hash: hash, Interface: iface}, nil
}

func (c *DexIpfsWriteClient) implementations(ifaceHash string) (*dex.Implementations, bool, error) {
	res, exists, err := c.ipfsClient.ReadFile(implementationsPath(ifaceHash))
	if err != nil {
		return nil, false, err
	}
	if !exists {
		return nil, false, nil
	}

	// TODO if file does not exist yet

	buf, err := io.ReadAll(res)
	if err != nil {
		return nil, false, err
	}

	var impls dex.Implementations
	err = json.Unmarshal(buf, &impls)
	if err != nil {
		return nil, false, err
	}

	return &impls, true, nil
}

func (c *DexIpfsWriteClient) AddImplementation(ifaceHash string, datastoreID string, query string) error {
	impls, exists, err := c.implementations(ifaceHash)
	if err != nil {
		return err
	}

	if !exists {
		fmt.Println("NOT EXIST")
		impls = &dex.Implementations{
			InterfaceHash: ifaceHash,
			Map:           map[string]string{},
		}
	}

	// check if impl for datastore already exists
	if _, ok := impls.Map[datastoreID]; ok {
		fmt.Printf("Implementation query already exists for datastore %s", datastoreID)
		return nil
	}

	impls.Map[datastoreID] = query
	marshaled, err := json.Marshal(&impls)
	if err != nil {
		return err
	}

	r := bytes.NewBuffer(marshaled)
	_, err = c.ipfsClient.WriteFile(implementationsPath(ifaceHash), "implementations", r)
	if err != nil {
		return err
	}

	return nil
}

func (c *DexIpfsWriteClient) RemoveImplementation(ifaceHash string, datastoreID string) error {
	impls, exists, err := c.implementations(ifaceHash)
	if err != nil {
		return err
	}

	if !exists {
		fmt.Printf("No implementation record exists for interface %s", ifaceHash)
	}

	if _, ok := impls.Map[datastoreID]; !ok {
		fmt.Printf("No implementation for datastore %s", datastoreID)
		return nil
	}

	delete(impls.Map, datastoreID)
	marshaled, err := json.Marshal(&impls)
	if err != nil {
		return err
	}

	r := bytes.NewBuffer(marshaled)
	_, err = c.ipfsClient.WriteFile(implementationsPath(ifaceHash), "implementations", r)
	if err != nil {
		return err
	}

	return nil
}
