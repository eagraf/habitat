package rpc

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
)

func SendRPC(address string, input []byte) (*RPCResponse, error) {

	req := &RPCRequest{
		Data: input,
	}

	buf, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	body := bytes.NewReader(buf)
	resp, err := http.Post(address, "application/json", body)
	if err != nil {
		return nil, err
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(string(respBody))
	}

	var rpcResp RPCResponse
	err = json.Unmarshal(respBody, &rpcResp)
	if err != nil {
		return nil, err
	}

	return &rpcResp, nil
}

func SendRPCDeserializeResponse(address string, input []byte, response interface{}) error {
	rpcResp, err := SendRPC(address, input)
	if err != nil {
		return err
	}

	// Try to unmarshal into user supplied interface
	err = json.Unmarshal(rpcResp.Response, response)
	if err != nil {
		return err
	}

	return nil
}
