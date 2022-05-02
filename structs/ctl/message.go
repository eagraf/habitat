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

func requestType(req interface{}) (string, error) {
	switch req.(type) {
	case StartRequest, *StartRequest, StartResponse, *StartResponse:
		return CommandStart, nil
	case StopRequest, *StopRequest, StopResponse, *StopResponse:
		return CommandStop, nil
	case PSRequest, *PSRequest, PSResponse, *PSResponse:
		return CommandListProcesses, nil
	case CommunityCreateRequest, *CommunityCreateRequest, CommunityCreateResponse, *CommunityCreateResponse:
		return CommandCommunityCreate, nil
	case CommunityJoinRequest, *CommunityJoinRequest, CommunityJoinResponse, *CommunityJoinResponse:
		return CommandCommunityJoin, nil
	case CommunityAddMemberRequest, *CommunityAddMemberRequest, CommunityAddMemberResponse, *CommunityAddMemberResponse:
		return CommandCommunityAddMember, nil
	case CommunityProposeRequest, *CommunityProposeRequest, CommunityProposeResponse, *CommunityProposeResponse:
		return CommandCommunityPropose, nil
	case CommunityStateRequest, *CommunityStateRequest, CommunityStateResponse, *CommunityStateResponse:
		return CommandCommunityState, nil
	case CommunityListRequest, *CommunityListRequest, CommunityListResponse, *CommunityListResponse:
		return CommandCommunityList, nil
	default:
		return "", fmt.Errorf("type %T is not a valid request type", req)
	}
}

type RequestWrapper struct {
	/*	Command string   `json:"command"`
		Args    []string `json:"args"`
		Env     []string `json:"env"`
		Flags   []string `json:"flags"`*/
	Type       string `json:"request_type"`
	Serialized string `json:"serialized"`
}

func NewRequestWrapper(req interface{}) (*RequestWrapper, error) {
	reqType, err := requestType(req)
	if err != nil {
		return nil, err
	}

	serialized, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	return &RequestWrapper{
		Type:       reqType,
		Serialized: string(serialized),
	}, nil
}

func (r *RequestWrapper) Deserialize(dest interface{}) error {
	return json.Unmarshal([]byte(r.Serialized), dest)
}

type ResponseWrapper struct {
	Type       string `json:"response_type"`
	Status     int    `json:"status"`
	Serialized string `json:"serialized"`
	Error      string `json:"error"`
}

func NewResponseWrapper(res interface{}, status int, errorMessage string) (*ResponseWrapper, error) {
	resType, err := requestType(res)
	if err != nil {
		return nil, err
	}

	serialized, err := json.Marshal(res)
	if err != nil {
		return nil, err
	}

	return &ResponseWrapper{
		Type:       resType,
		Status:     status,
		Serialized: string(serialized),
		Error:      errorMessage,
	}, nil
}

func (r *ResponseWrapper) Deserialize(res interface{}) error {
	return json.Unmarshal([]byte(r.Serialized), res)
}

func (r *ResponseWrapper) Encode() ([]byte, error) {
	marshalled, err := json.Marshal(r)
	if err != nil {
		return nil, err
	}

	// base64 encode to make sure newlines are not present in bytes sent
	encoded := base64.StdEncoding.EncodeToString(marshalled)

	msg := append([]byte(encoded), '\n')

	return msg, nil
}

/*func (r *ResponseWrapper) String() string {
	if r.Status != 0 {
		return fmt.Sprintf("Error: %s", r.Message)
	}
	return string(r.Message)
}*/
