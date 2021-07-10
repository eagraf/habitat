package ctl

import (
	"encoding/base64"
	"encoding/json"
)

const (
	CommandStart = "start"
	CommandStop  = "stop"

	StatusOK                  = 0
	StatusBadRequest          = 1
	StatusInternalServerError = 2
)

type Request struct {
	Command string `json:"command"`
}

type Response struct {
	Status  int
	Message string
}

func (r *Response) Encode() ([]byte, error) {
	marshalled, err := json.Marshal(r)
	if err != nil {
		return nil, err
	}

	// base64 encode to make sure newlines are not present in bytes sent
	encoded := base64.StdEncoding.EncodeToString(marshalled)

	msg := append([]byte(encoded), '\n')

	return msg, nil
}
