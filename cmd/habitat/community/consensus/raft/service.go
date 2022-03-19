package raft

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/eagraf/habitat/pkg/compass"
	"github.com/eagraf/habitat/pkg/raft/fsm"
	"github.com/eagraf/habitat/pkg/raft/transport"
	"github.com/hashicorp/raft"
	raftboltdb "github.com/hashicorp/raft-boltdb"
	"github.com/libp2p/go-libp2p-core/host"
	peer "github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/peerstore"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/rs/zerolog/log"
)

const (
	RetainSnapshotCount = 1000
	RaftTimeout         = 10 * time.Second
)

// ClusterService is an implementation of cluster.ClusterManager
type ClusterService struct {
	instances map[string]*raftClusterInstance

	host host.Host

	nodeID string
}

type raftClusterInstance struct {
	communityID  string
	serverID     string
	address      string
	instance     *raft.Raft
	stateMachine *fsm.CommunityStateMachine
}

func NewClusterService(host host.Host) *ClusterService {
	cs := &ClusterService{
		instances: make(map[string]*raftClusterInstance),
		host:      host,

		nodeID: compass.NodeID(),
	}

	return cs
}

func (cs *ClusterService) Start() error {
	return nil
}

// CreateCluster initializes a new Raft cluster, and bootstraps it with this nodes address
func (cs *ClusterService) CreateCluster(communityID string) error {
	if _, ok := cs.instances[communityID]; ok {
		return fmt.Errorf("raft instance for community %s already initialized", communityID)
	}

	stateMachine := fsm.NewCommunityStateMachine()

	ra, err := setupRaftInstance(communityID, stateMachine, true, cs.host)
	if err != nil {
		return fmt.Errorf("failed to setup raft instance: %s", err.Error())
	}

	raftInstance := &raftClusterInstance{
		communityID:  communityID,
		serverID:     getServerID(communityID),
		address:      getCommunityAddress(communityID),
		instance:     ra,
		stateMachine: stateMachine,
	}
	cs.instances[communityID] = raftInstance

	return nil
}

func (cs *ClusterService) RemoveCluster(communityID string) error {
	// TODO
	return nil
}

// JoinCluster requests for this node to join a cluster.
// In this implementation, the address is unused because the leader will begin sending
// heartbeets to this node once its AddNode routine has been called.
func (cs *ClusterService) JoinCluster(communityID string, address string) error {
	if _, ok := cs.instances[communityID]; ok {
		return fmt.Errorf("raft instance for community %s already initialized", communityID)
	}

	stateMachine := fsm.NewCommunityStateMachine()

	ra, err := setupRaftInstance(communityID, stateMachine, false, cs.host)
	if err != nil {
		return fmt.Errorf("failed to setup raft instance: %s", err.Error())
	}

	raftInstance := &raftClusterInstance{
		communityID:  communityID,
		serverID:     getServerID(communityID),
		address:      getCommunityAddress(communityID),
		instance:     ra,
		stateMachine: stateMachine,
	}
	cs.instances[communityID] = raftInstance

	return nil
}

// RestoreNode restarts this nodes raft instance if it has been stopped. All data is
// restored from snapshots and the write ahead log in stable storage.
func (cs *ClusterService) RestoreNode(communityID string) error {
	if _, ok := cs.instances[communityID]; ok {
		log.Error().Msgf("raft instance for community %s already initialized", communityID)
	}

	stateMachine := fsm.NewCommunityStateMachine()

	ra, err := setupRaftInstance(communityID, stateMachine, false, cs.host)
	if err != nil {
		log.Error().Msgf("failed to setup raft instance: %s", err.Error())
	}

	raftInstance := &raftClusterInstance{
		communityID:  communityID,
		serverID:     getServerID(communityID),
		address:      getCommunityAddress(communityID),
		instance:     ra,
		stateMachine: stateMachine,
	}
	cs.instances[communityID] = raftInstance

	return nil
}

// ProposeTransition takes a proposed update to community state in the form of a JSON patch,
// and attempts to get other nodes to agree to apply the transition to the state machine.
// If succesfully commited, the updated state should be available via the GetState() call.
// TODO if this node is a follower, forward transition to leader
func (cs *ClusterService) ProposeTransition(communityID string, transition []byte) error {
	log.Info().Msgf("applying transition to %s", communityID)

	raftInstance, ok := cs.instances[communityID]
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

// GetState returns the state tracked by the Raft instance's state machine. It returns
// a byte array with a marshaled JSON object inside.
func (cs *ClusterService) GetState(communityID string) ([]byte, error) {
	log.Info().Msgf("getting state for %s", communityID)

	raftInstance, ok := cs.instances[communityID]
	if !ok {
		return nil, fmt.Errorf("community %s raft instance does not exist", communityID)
	}

	state, err := raftInstance.stateMachine.State()
	if err != nil {
		return nil, fmt.Errorf("error getting raft instance's community state: %s", err)
	}

	return state, nil
}

// AddNode adds a new node to a cluster. After a node has been added, it must execute
// the JoinCluster routine to begin listening for state updates.
// nodeID represents a libp2p peer id base58 encoded in this instance
func (cs *ClusterService) AddNode(communityID string, nodeID string, address string) error {
	log.Info().Msgf("received request for %s at %s to join %s", nodeID, address, communityID)

	// add remote node address to this host's peer store
	addr, err := ma.NewMultiaddr(address)
	if err != nil {
		return err
	}

	// decode base58 encoded peer id for setting addresses
	peerID, err := peer.Decode(nodeID)
	if err != nil {
		return err
	}

	cs.host.Peerstore().AddAddrs(peerID, []ma.Multiaddr{addr}, peerstore.PermanentAddrTTL)

	// TODO the joining server should be authenticated at this point

	raftInstance, ok := cs.instances[communityID]
	if !ok {
		return fmt.Errorf("community %s raft instance does not exist", communityID)
	}

	configFuture := raftInstance.instance.GetConfiguration()

	for _, srv := range configFuture.Configuration().Servers {
		// If a node already exists with either the joining node's ID or address,
		// that node may need to be removed from the config first.
		if srv.ID == raft.ServerID(nodeID) || srv.Address == raft.ServerAddress(address) {
			// However if *both* the ID and the address are the same, then nothing -- not even
			// a join operation -- is needed.
			if srv.Address == raft.ServerAddress(address) && srv.ID == raft.ServerID(nodeID) {
				return fmt.Errorf("node %s at %s already member of cluster, ignoring join request", nodeID, address)
			}

			future := raftInstance.instance.RemoveServer(srv.ID, 0, 0)
			if err := future.Error(); err != nil {
				return fmt.Errorf("error removing existing node %s at %s: %s", nodeID, address, err)
			}
		}
	}

	f := raftInstance.instance.AddVoter(raft.ServerID(nodeID), raft.ServerAddress(address), 0, 0)
	if f.Error() != nil {
		return f.Error()
	}

	return nil
}

func (cs *ClusterService) RemoveNode(communityID string, nodeID string) error {
	// TODO
	return nil
}

// Internal wrapper of Hashicorp raft stuff
func setupRaftInstance(communityID string, stateMachine *fsm.CommunityStateMachine, newCommunity bool, host host.Host) (*raft.Raft, error) {
	log.Info().Msgf("setting up raft instance for node %s at %s", getServerID(communityID), getCommunityAddress(communityID))

	// setup raft folder
	raftDirPath := filepath.Join(compass.CommunitiesPath(), communityID, "raft")
	raftDBPath := filepath.Join(raftDirPath, "raft.db")

	_, err := os.Stat(raftDBPath)
	if errors.Is(err, os.ErrNotExist) {
		err := os.Mkdir(raftDirPath, 0700)
		if err != nil {
			return nil, fmt.Errorf("error creating raft directory for new community: %s", err)
		}

		raftDBFile, err := os.OpenFile(raftDBPath, os.O_CREATE|os.O_RDONLY, 0600)
		if err != nil {
			return nil, fmt.Errorf("error creating raft bolt db file: %s", err)
		}
		defer raftDBFile.Close()
	} else if err != nil {
		return nil, err
	}

	// Setup Raft configuration.
	config := raft.DefaultConfig()
	config.LocalID = raft.ServerID(host.ID().Pretty())

	// Setup Raft communication.
	protocol := getClusterProtocol(communityID)
	libP2PTransport := transport.NewLibP2PTransport(host, protocol)

	// Create the snapshot store. This allows the Raft to truncate the log.
	snapshots, err := raft.NewFileSnapshotStore(getCommunityRaftDirectory(communityID), RetainSnapshotCount, os.Stderr)
	if err != nil {
		return nil, fmt.Errorf("file snapshot store: %s", err)
	}

	// Create the log store and stable store.
	var logStore raft.LogStore
	var stableStore raft.StableStore
	boltDB, err := raftboltdb.NewBoltStore(filepath.Join(getCommunityRaftDirectory(communityID), "raft.db"))
	if err != nil {
		return nil, fmt.Errorf("new bolt store: %s", err)
	}
	logStore = boltDB
	stableStore = boltDB

	// Instantiate the Raft systems.
	ra, err := raft.NewRaft(config, stateMachine, logStore, stableStore, snapshots, libP2PTransport)
	if err != nil {
		return nil, fmt.Errorf("new raft: %s", err)
	}

	// If this node is creating the community, bootstrap the raft cluster as well
	if newCommunity {
		configuration := raft.Configuration{
			Servers: []raft.Server{
				{
					ID:      config.LocalID,
					Address: libP2PTransport.LocalAddr(),
				},
			},
		}
		ra.BootstrapCluster(configuration)
	}

	return ra, nil
}
