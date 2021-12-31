package community

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/eagraf/habitat/apps/raft/structs"
	"github.com/eagraf/habitat/pkg/rpc"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

type Manager struct {
	Path string
}

func NewManager(path string) *Manager {
	return &Manager{
		Path: path,
	}
}

func (m *Manager) setupCommunity(communityID string) error {
	path := path.Join(m.Path, communityID)

	// check if community dir already exists
	_, err := os.Stat(path)
	if err == nil {
		return fmt.Errorf("data dir for community %s already exists", path)
	}

	// create community dir
	err = os.MkdirAll(path, 0766)
	if err != nil {
		return err
	}

	return nil
}

func (m *Manager) checkCommunityExists(communityID string) bool {
	path := path.Join(m.Path, communityID)

	_, err := os.Stat(path)
	return err == nil

}

func (m *Manager) CreateCommunity() (string, error) {
	// Generate UUID for now
	id := uuid.New()
	communityID := id.String()

	err := m.setupCommunity(communityID)
	if err != nil {
		return "", err
	}

	// setup raft folder
	raftDirPath := filepath.Join(m.Path, communityID, "raft")
	err = os.Mkdir(raftDirPath, 0700)
	if err != nil {
		return "", fmt.Errorf("error creating raft directory for new community: %s", err)
	}

	raftDBPath := filepath.Join(raftDirPath, "raft.db")
	raftDBFile, err := os.OpenFile(raftDBPath, os.O_CREATE|os.O_RDONLY, 0600)
	if err != nil {
		return "", fmt.Errorf("error creating raft bolt db file")
	}
	defer raftDBFile.Close()

	err = sendRegisterRPC(communityID, true, "")
	if err != nil {
		return "", fmt.Errorf("error sending register RPC: %s", err)
	}

	return communityID, nil
}

func (m *Manager) JoinCommunity(acceptingNodeAddr string, communityID string) error {
	err := m.setupCommunity(communityID)
	if err != nil {
		return fmt.Errorf("error setting up community: %s", err)
	}

	err = sendRegisterRPC(communityID, false, acceptingNodeAddr)
	if err != nil {
		return fmt.Errorf("error sending register RPC: %s", err)
	}

	return nil
}

func (m *Manager) ProposeTransition(communityID string, transition []byte) error {
	if !m.checkCommunityExists(communityID) {
		return fmt.Errorf("community %s does not exist in communities directory", communityID)
	}

	err := sendProposeTransitionRPC(communityID, transition)
	if err != nil {
		return fmt.Errorf("error sending RPC: %s", err)
	}

	return nil
}

func sendRegisterRPC(communityID string, createCommunity bool, joinAddr string) error {
	// send RPC to raft service to start syncing with the cluster
	req := &structs.RegisterCommunityRequest{
		CommunityID:  communityID,
		NewCommunity: createCommunity,
		JoinAddress:  joinAddr,
	}
	buf, err := json.Marshal(req)
	if err != nil {
		return err
	}

	// TODO don't hardcode the address
	resp, err := rpc.SendRPC("http://0.0.0.0:2041/raft/rpc/register", buf)
	if err != nil {
		log.Error().Err(err)
		return err
	}

	if resp.StatusCode != structs.RPC_STATUS_OK {
		return fmt.Errorf("register RPC returned error: %s", string(resp.Response))
	}

	return nil
}

func sendProposeTransitionRPC(communityID string, transition []byte) error {
	req := &structs.ApplyStateTransitionRequest{
		CommunityID: communityID,
		Transition:  transition,
	}
	buf, err := json.Marshal(req)
	if err != nil {
		return err
	}

	resp, err := rpc.SendRPC("http://0.0.0.0:2041/raft/rpc/apply", buf)
	if err != nil {
		log.Error().Err(err)
		return err
	}

	if resp.StatusCode != structs.RPC_STATUS_OK {
		return fmt.Errorf("register RPC returned error: %s", string(resp.Response))
	}

	return nil
}
