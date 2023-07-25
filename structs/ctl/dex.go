package ctl

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
)

type SchemaRequest struct {
	Hash string `json:"hash"`
}

type SchemaResponse struct {
	Hash   string `json:"hash"`
	Schema Schema `json:"schema"`
}

type SchemaWriteRequest struct {
	Schema Schema `json:"schema"`
}

type SchemaWriteResponse struct {
	Hash string `json:"hash"`
}

type InterfaceRequest struct {
	Hash string `json:"hash"`
}

type InterfaceResponse struct {
	Hash      string     `json:"hash"`
	Interface *Interface `json:"interface"`
}

type InterfaceWriteRequest struct {
	Interface *Interface `json:"interface"`
}

type InterfaceWriteResponse struct {
	Hash string `json:"hash"`
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

func hash(buf []byte) string {
	h := sha1.Sum(buf)
	return hex.EncodeToString(h[:])
}
