package dex

import (
	"encoding/json"
	"errors"
)

type Driver interface {
	Get(addr string) (*GetResult, error)
	Schema(hash string) (*SchemaResult, error)
	Interface(hash string) (*InterfaceResult, error)
	Implementations(interfaceHash string) (*ImplementationsResult, error)
}

type GetResult struct {
	Body json.RawMessage `json:"body"`
}

type SchemaResult struct {
	Hash   string `json:"hash"`
	Schema Schema `json:"schema"`
}

type InterfaceResult struct {
	Hash      string     `json:"hash"`
	Interface *Interface `json:"interface"`
}

type ImplementationsResult struct {
	Implementations *Implementations `json:"implementations"`
}

type Schema json.RawMessage

func (s Schema) MarshalJSON() ([]byte, error) {
	raw := json.RawMessage(s)
	return raw.MarshalJSON()
}

func (s *Schema) UnmarshalJSON(data []byte) error {
	if s == nil {
		return errors.New("dex.Schema: UnmarshalJSON on nil pointer")
	}
	*s = append((*s)[0:0], data...)
	return nil
}

func (s Schema) Hash() (string, error) {
	return hash(s), nil
}

type Interface struct {
	SchemaHash  string `json:"schema_hash"`
	Description string `json:"description"`
}

func (i *Interface) Hash() (string, error) {
	marshaled, err := json.Marshal(i)
	if err != nil {
		return "", err
	}
	return hash(marshaled), nil
}

type Implementations struct {
	InterfaceHash string            `json:"interface_hash"`
	Map           map[string]string `json:"map"`
}

// No-op implementation of the Driver interface for testing
type NoopDriver struct {
	MockGetRes             *GetResult
	MockSchemaRes          *SchemaResult
	MockInterfaceRes       *InterfaceResult
	MockImplementationsRes *ImplementationsResult
}

func (n *NoopDriver) Get(addr string) (*GetResult, error) {
	return n.MockGetRes, nil
}

func (n *NoopDriver) Schema(hash string) (*SchemaResult, error) {
	return n.MockSchemaRes, nil
}

func (n *NoopDriver) Interface(hash string) (*InterfaceResult, error) {
	return n.MockInterfaceRes, nil
}

func (n *NoopDriver) Implementations(interfaceHash string) (*ImplementationsResult, error) {
	return n.MockImplementationsRes, nil
}
