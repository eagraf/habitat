package ctl

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
)

const (
	CommandStart              = "start"
	CommandStop               = "stop"
	CommandListProcesses      = "ps"
	CommandCommunityCreate    = "community_create"
	CommandCommunityJoin      = "community_join"
	CommandCommunityAddMember = "community_add_member"
	CommandCommunityPropose   = "community_propose"
	CommandCommunityState     = "community_state"
	CommandCommunityList      = "community_list"

	StatusOK                  = 0
	StatusBadRequest          = 1
	StatusInternalServerError = 2
)

type Request struct {
	Command string   `json:"command"`
	Args    []string `json:"args"`
	Env     []string `json:"env"`
	Flags   []string `json:"flags"`
}

type Response struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
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

func (r *Response) String() string {
	if r.Status != 0 {
		return fmt.Sprintf("Error: %s", r.Message)
	}
	return string(r.Message)
}
