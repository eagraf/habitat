package community

import (
	"errors"
	"fmt"

	"github.com/eagraf/habitat/structs/ctl"
)

type CTLHandler func(req *ctl.Request) (*ctl.Response, error)

func (m *Manager) CommunityCreateHandler(req *ctl.Request) (*ctl.Response, error) {
	id, err := m.CreateCommunity()
	if err != nil {
		return nil, err
	}
	return &ctl.Response{
		Status:  ctl.StatusOK,
		Message: fmt.Sprintf("created community with uuid: %s", id),
	}, nil
}

func (m *Manager) CommunityJoinHandler(req *ctl.Request) (*ctl.Response, error) {
	// validate args
	if len(req.Args) != 2 {
		return nil, errors.New("need 2 arguments to join community")
	}

	err := m.JoinCommunity(req.Args[0], req.Args[1])
	if err != nil {
		return nil, err
	}

	return &ctl.Response{
		Status:  ctl.StatusOK,
		Message: fmt.Sprintf("joined community %s", req.Args[0]),
	}, nil
}

func (m *Manager) CommunityAddMemberHandler(req *ctl.Request) (*ctl.Response, error) {
	return nil, nil
}

func (m *Manager) CommunityProposeHandler(req *ctl.Request) (*ctl.Response, error) {
	// validate args
	if len(req.Args) != 1 {
		return nil, errors.New("need 1 arguments to join community")
	}

	err := m.ProposeTransition(req.Args[0], []byte("THIS IS A SAMPLE PROPOSAL"))
	if err != nil {
		return nil, err
	}

	return &ctl.Response{
		Status:  ctl.StatusOK,
		Message: fmt.Sprintf("proposed transition"),
	}, nil
}
