package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/eagraf/habitat/pkg/dex"
	"github.com/eagraf/habitat/pkg/ipfs"
)

type IPFSDexDriver struct {
	ipfsClient *ipfs.Client
}

func NewIPFSDexDriver(ipfsPort string) (*IPFSDexDriver, error) {
	ipfsClient, err := ipfs.NewClient(fmt.Sprintf("http://localhost:%s/api/v0", ipfsPort))
	if err != nil {
		return nil, err
	}
	return &IPFSDexDriver{
		ipfsClient: ipfsClient,
	}, nil
}

func (d *IPFSDexDriver) Get(addr string) (*dex.GetResult, error) {
	return nil, nil
}

func (d *IPFSDexDriver) Schema(hash string) (*dex.SchemaResult, error) {
	r, _, err := d.ipfsClient.ReadFile(schemaPath(hash))
	if err != nil {
		return nil, err
	}

	buf, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	return &dex.SchemaResult{
		Hash:   hash,
		Schema: dex.Schema(buf),
	}, nil
}

func (d *IPFSDexDriver) Interface(hash string) (*dex.InterfaceResult, error) {
	r, _, err := d.ipfsClient.ReadFile(interfacePath(hash))
	if err != nil {
		return nil, err
	}

	buf, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	var iface dex.Interface
	err = json.Unmarshal(buf, &iface)
	if err != nil {
		return nil, err
	}

	return &dex.InterfaceResult{
		Hash:      hash,
		Interface: &iface,
	}, nil
}

func (d *IPFSDexDriver) Implementations(interfaceHash string) (*dex.ImplementationsResult, error) {
	r, exists, err := d.ipfsClient.ReadFile(implementationsPath(interfaceHash))
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.New("no implementations found")
	}

	buf, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	var impls dex.Implementations
	err = json.Unmarshal(buf, &impls)
	if err != nil {
		return nil, err
	}

	return &dex.ImplementationsResult{
		Implementations: &impls,
	}, nil
}
