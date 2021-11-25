package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/eagraf/habitat/apps/raft/structs"
	"github.com/eagraf/habitat/pkg/compass"
	"github.com/eagraf/habitat/pkg/rpc"
	"github.com/hashicorp/raft"
	raftboltdb "github.com/hashicorp/raft-boltdb"
	"github.com/rs/zerolog/log"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/eagraf/habitat/pkg/raft/fsm"
	"github.com/eagraf/habitat/pkg/raft/transport"
)

const (
	Localhost  = "localhost"
	DockerHost = "0.0.0.0"
	Protocol   = "http"

	MultiplexerPort = "6000"
	RPCPort         = "6001"

	RetainSnapshotCount = 1000

	RaftTimeout = 10 * time.Second
)

func main() {
	pflag.String("communitydir", "/habitat/data/communities", "directory where communities are stored")
	pflag.String("hostname", "", "hostname that this node can be reached at")
	pflag.BoolP("docker", "d", true, "use docker host rather than localhost")
	pflag.String("node", "", "the host node's id")

	pflag.Parse()
	viper.BindPFlags(pflag.CommandLine)

	rm := newRaftManager()

	handlers := map[string]rpc.RPCHandlerFunc{
		structs.RegisterRoute:             rm.registerCommunityHandler,
		structs.UnregisterRoute:           rm.unregisterCommunityHandler,
		structs.AddNodeRoute:              rm.addNodeHandler,
		structs.ApplyStateTransitionRoute: rm.applyStateTransitionHandler,
	}

	rpcServer := rpc.NewServer(handlers)
	rpcServer.Start(getRPCAddress())
}

func getMultiplexerAddress() string {
	host := Localhost
	usingDocker := viper.GetBool("docker")
	if usingDocker {
		host = compass.Hostname()
	}

	return fmt.Sprintf("%s:%s", host, MultiplexerPort)
}

func getRPCAddress() string {
	host := Localhost
	usingDocker := viper.GetBool("docker")
	if usingDocker {
		host = compass.Hostname()
	}

	return fmt.Sprintf("%s:%s", host, RPCPort)
}

func getServerID(communityID string) string {
	nodeID := compass.NodeID()
	return fmt.Sprintf("%s#%s", nodeID, communityID)
}

func getCommunityAddress(communityID string) string {
	return fmt.Sprintf("%s/%s", getMultiplexerAddress(), communityID)
}

func getCommunityExternalAddress(communityID string) string {
	return fmt.Sprintf("%s://%s", Protocol, getCommunityAddress(communityID))
}

func getCommunityRaftDirectory(communityID string) string {
	return filepath.Join(viper.GetString("communitydir"), communityID, "raft")
}

type raftManager struct {
	mux       transport.RaftMultiplexer
	instances map[string]*raftCommunityInstance

	nodeID   string
	hostname string
}

type raftCommunityInstance struct {
	communityID  string
	serverID     string
	address      string
	instance     *raft.Raft
	stateMachine *fsm.CommunityStateMachine
}

func newRaftManager() *raftManager {
	rm := &raftManager{
		mux:       *transport.NewRaftMultiplexer(),
		instances: make(map[string]*raftCommunityInstance),

		nodeID:   compass.NodeID(),
		hostname: compass.HabitatPath(),
	}
	go rm.mux.Listen(getMultiplexerAddress())
	return rm
}

func rpcError(msg string) (int, []byte) {
	return structs.RPC_STATUS_ERROR, []byte(msg)
}

func (m *raftManager) registerCommunityHandler(data []byte) (int, []byte) {
	var req rpc.RPCRequest
	err := json.Unmarshal(data, &req)
	if err != nil {
		return rpcError(fmt.Sprintf("failed to unmarshal RPC request: %s", err.Error()))
	}

	var registerReq structs.RegisterCommunityRequest
	err = json.Unmarshal(req.Data, &registerReq)
	if err != nil {
		return rpcError(fmt.Sprintf("failed to unmarshal RPC data: %s", err.Error()))
	}

	if _, ok := m.instances[registerReq.CommunityID]; ok {
		return rpcError(fmt.Sprintf("raft instance for community %s already initialized", registerReq.CommunityID))
	}

	// TODO do some stuff here
	ra, ht, err := m.setupRaftInstance(registerReq.CommunityID, registerReq.NewCommunity)
	if err != nil {
		return rpcError(fmt.Sprintf("failed to setup raft instance: %s", err.Error()))
	}

	err = m.mux.RegisterListener(registerReq.CommunityID, ht)
	if err != nil {
		return rpcError(fmt.Sprintf("failed to register listener: %s", err.Error()))
	}

	raftInstance := &raftCommunityInstance{
		communityID:  registerReq.CommunityID,
		serverID:     getServerID(registerReq.CommunityID),
		address:      getCommunityExternalAddress(registerReq.CommunityID),
		instance:     ra,
		stateMachine: &fsm.CommunityStateMachine{},
	}
	m.instances[registerReq.CommunityID] = raftInstance

	if registerReq.JoinAddress != "" {
		err := sendAddNodeRPC(raftInstance.communityID, registerReq.JoinAddress, raftInstance.address, raftInstance.serverID)
		if err != nil {
			return rpcError(fmt.Sprintf("error joining cluster: %s", err))
		}
	}

	return structs.RPC_STATUS_OK, []byte{}
}

func (m *raftManager) unregisterCommunityHandler(data []byte) (int, []byte) {
	return 0, []byte("unregister community handler not implemented")
}

func (m *raftManager) addNodeHandler(data []byte) (int, []byte) {
	var req rpc.RPCRequest
	err := json.Unmarshal(data, &req)
	if err != nil {
		return rpcError(fmt.Sprintf("failed to unmarshal RPC request: %s", err.Error()))
	}

	var addNodeReq structs.AddNodeRequest
	err = json.Unmarshal(req.Data, &addNodeReq)
	if err != nil {
		return rpcError(fmt.Sprintf("failed to unmarshal RPC data: %s", err.Error()))
	}

	err = m.addNode(addNodeReq.CommunityID, addNodeReq.ServerID, addNodeReq.Address)
	if err != nil {
		return rpcError(fmt.Sprintf("failed to add node to Raft cluster: %s", err))
	}
	return structs.RPC_STATUS_OK, []byte{}
}

func (m *raftManager) applyStateTransitionHandler(data []byte) (int, []byte) {
	var req rpc.RPCRequest
	err := json.Unmarshal(data, &req)
	if err != nil {
		return rpcError(fmt.Sprintf("failed to unmarshal RPC request: %s", err.Error()))
	}

	var applyReq structs.ApplyStateTransitionRequest
	err = json.Unmarshal(req.Data, &applyReq)
	if err != nil {
		return rpcError(fmt.Sprintf("failed to unmarshal RPC data: %s", err.Error()))
	}

	err = m.applyStateTransition(applyReq.CommunityID, applyReq.Transition)
	if err != nil {
		return rpcError(fmt.Sprintf("failed to apply state transition: %s", err))
	}
	return structs.RPC_STATUS_OK, nil
}

func (m *raftManager) setupRaftInstance(communityID string, newCommunity bool) (*raft.Raft, *transport.HTTPTransport, error) {
	log.Info().Msgf("setting up raft instance for node %s at %s", getServerID(communityID), getCommunityAddress(communityID))

	// Setup Raft configuration.
	config := raft.DefaultConfig()
	config.LocalID = raft.ServerID(getServerID(communityID))

	// Setup Raft communication.
	httpTransport, err := transport.NewHTTPTransport(getCommunityExternalAddress(communityID))
	if err != nil {
		return nil, nil, err
	}

	// Create the snapshot store. This allows the Raft to truncate the log.
	snapshots, err := raft.NewFileSnapshotStore(getCommunityRaftDirectory(communityID), RetainSnapshotCount, os.Stderr)
	if err != nil {
		return nil, nil, fmt.Errorf("file snapshot store: %s", err)
	}

	// Create the log store and stable store.
	var logStore raft.LogStore
	var stableStore raft.StableStore
	boltDB, err := raftboltdb.NewBoltStore(filepath.Join(getCommunityRaftDirectory(communityID), "raft.db"))
	if err != nil {
		return nil, nil, fmt.Errorf("new bolt store: %s", err)
	}
	logStore = boltDB
	stableStore = boltDB

	// Initialize state machine
	sm := fsm.NewCommunityStateMachine()

	// Instantiate the Raft systems.
	ra, err := raft.NewRaft(config, sm, logStore, stableStore, snapshots, httpTransport)
	if err != nil {
		return nil, nil, fmt.Errorf("new raft: %s", err)
	}

	// If this node is creating the community, bootstrap the raft cluster as well
	if newCommunity {
		configuration := raft.Configuration{
			Servers: []raft.Server{
				{
					ID:      config.LocalID,
					Address: httpTransport.LocalAddr(),
				},
			},
		}
		ra.BootstrapCluster(configuration)
	}

	return ra, httpTransport, nil
}

func (m *raftManager) addNode(communityID string, serverID string, address string) error {
	log.Info().Msgf("received request for %s at %s to join %s", serverID, address, communityID)

	raftInstance, ok := m.instances[communityID]
	if !ok {
		return fmt.Errorf("community %s raft instance does not exist", communityID)
	}

	configFuture := raftInstance.instance.GetConfiguration()

	for _, srv := range configFuture.Configuration().Servers {
		// If a node already exists with either the joining node's ID or address,
		// that node may need to be removed from the config first.
		if srv.ID == raft.ServerID(serverID) || srv.Address == raft.ServerAddress(address) {
			// However if *both* the ID and the address are the same, then nothing -- not even
			// a join operation -- is needed.
			if srv.Address == raft.ServerAddress(address) && srv.ID == raft.ServerID(serverID) {
				return fmt.Errorf("node %s at %s already member of cluster, ignoring join request", serverID, address)
			}

			future := raftInstance.instance.RemoveServer(srv.ID, 0, 0)
			if err := future.Error(); err != nil {
				return fmt.Errorf("error removing existing node %s at %s: %s", serverID, address, err)
			}
		}
	}

	f := raftInstance.instance.AddVoter(raft.ServerID(serverID), raft.ServerAddress(address), 0, 0)
	if f.Error() != nil {
		return f.Error()
	}

	return nil
}

func sendAddNodeRPC(communityID string, acceptingNodeAddr string, joiningNodeAddr string, joiningNodeID string) error {
	req := &structs.AddNodeRequest{
		CommunityID: communityID,
		ServerID:    joiningNodeID,
		Address:     joiningNodeAddr,
	}

	buf, err := json.Marshal(req)
	if err != nil {
		return err
	}

	resp, err := rpc.SendRPC(acceptingNodeAddr, buf)
	if err != nil {
		return err
	}

	if resp.StatusCode != structs.RPC_STATUS_OK {
		return fmt.Errorf("add node RPC returned error: %s", string(resp.Response))
	}

	return nil
}

func (m *raftManager) applyStateTransition(communityID string, transition []byte) error {
	log.Info().Msgf("applying transition to %s", communityID)

	raftInstance, ok := m.instances[communityID]
	if !ok {
		return fmt.Errorf("community %s raft instance does not exist", communityID)
	}

	future := raftInstance.instance.Apply(transition, RaftTimeout)
	err := future.Error()
	if err != nil {
		return fmt.Errorf("error applying state transition to community %s: %s", communityID, err)
	}

	return nil
}
