package rpc

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRPCServer(t *testing.T) {
	successFunc := func(in []byte) (int, []byte) {
		return 0, []byte("Hey you guys")
	}

	failureFunc := func([]byte) (int, []byte) {
		return 1, []byte("woops something broke")
	}

	routes := map[string]RPCHandlerFunc{
		"/good": successFunc,
		"/bad":  failureFunc,
	}

	server := NewServer(routes)
	go server.Start("localhost:1234")

	resp, err := SendRPC("http://localhost:1234/good", []byte{})
	assert.Nil(t, err)
	assert.Equal(t, "Hey you guys", string(resp.Response))

	resp, err = SendRPC("http://localhost:1234/bad", []byte{})
	assert.Nil(t, err)
	assert.Equal(t, "woops something broke", string(resp.Response))

	resp, err = SendRPC("http://localhost:1234/ugly", []byte{})
	assert.NotNil(t, err)
}

func TestRPCDeserializationHelper(t *testing.T) {
	type myType struct {
		Abc         string
		Onetwothree int
	}

	a := myType{"abc", 123}
	data, err := json.Marshal(a)
	if err != nil {
		t.Error(err)
	}

	foo := func([]byte) (int, []byte) {
		return 0, data
	}

	routes := map[string]RPCHandlerFunc{
		"/foo": foo,
	}

	server := NewServer(routes)
	go server.Start("localhost:1234")

	var res myType
	err = SendRPCDeserializeResponse("http://localhost:1234/foo", []byte{}, &res)
	assert.Nil(t, err)
	assert.Equal(t, "abc", res.Abc)
}
