package ctl

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRequestSerialization(t *testing.T) {
	startReq := &StartRequest{
		App:  "myapp",
		Args: []string{"a", "b", "c"},
		Env:  []string{"a", "b", "c"},
	}

	wrapped, err := NewRequestWrapper(startReq)
	assert.Nil(t, err)
	var req StartRequest
	err = wrapped.Deserialize(&req)
	assert.Nil(t, err)
	assert.Equal(t, req.App, "myapp")

	psRes := &PSResponse{
		ProcIDs: []string{
			"ABC",
			"123",
		},
	}

	resWrapped, err := NewResponseWrapper(psRes, 0, "err")
	assert.Nil(t, err)
	var res PSResponse
	err = resWrapped.Deserialize(&res)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(res.ProcIDs))
}
