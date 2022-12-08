package community

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"sync"

	"github.com/eagraf/habitat/cmd/habitat/community/consensus/cluster"
	"github.com/eagraf/habitat/cmd/habitat/community/state"
	"github.com/eagraf/habitat/cmd/habitat/node"
	"github.com/eagraf/habitat/cmd/habitat/procs"
	"github.com/eagraf/habitat/pkg/compass"
	"github.com/eagraf/habitat/pkg/ipfs"
	"github.com/eagraf/habitat/structs/community"
	"github.com/google/uuid"
	"github.com/libp2p/go-libp2p/core/event"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/rs/zerolog/log"
)

type Manager struct {
	Path   string
	config *ipfs.IPFSConfig
	node   *node.Node

	clusterManager  *cluster.ClusterManager
	communities     map[string]*state.CommunityStateMachine
	communitiesLock *sync.Mutex

	nodeID       string
	reachability network.Reachability
}

func NewManager(path string, habitatNode *node.Node) (*Manager, error) {
	clusterManager := cluster.NewClusterManager(habitatNode.P2PNode.Host())

	err := clusterManager.Start(&habitatNode.ReverseProxy.Rules)
	if err != nil {
		return nil, fmt.Errorf("error starting cluster manager: %s", err)
	}

	manager := &Manager{
		Path: path,
		config: &ipfs.IPFSConfig{
			CommunitiesPath: path,
			// TODO: @arushibandi remove this usage of compass
			StartCmd: filepath.Join(compass.ProcsPath(), "bin", "amd64-darwin", "start-ipfs"),
		},
		node:            habitatNode,
		clusterManager:  clusterManager,
		communities:     make(map[string]*state.CommunityStateMachine),
		communitiesLock: &sync.Mutex{},

		nodeID: compass.NodeID(),
	}

	// Restart any existing communities
	comDirs, err := ioutil.ReadDir(compass.CommunitiesPath())
	if err == nil {
		for _, dir := range comDirs {
			updateChan, err := clusterManager.RestoreNode(dir.Name())
			if err != nil {
				log.Error().Err(err).Msgf("error restoring cluster for community %s", dir.Name())
			}

			initState := community.NewCommunityState()
			stateMachine, err := state.NewCommunityStateMachine(initState,
				updateChan, &ClusterDispatcher{
					communityID:    dir.Name(),
					clusterManager: manager.clusterManager,
				},
				NewCommunityExecutor(manager.node),
			)
			if err != nil {
				return nil, err
			}

			err = manager.addCommunity(dir.Name(), stateMachine)
			if err != nil {
				log.Error().Err(err).Msgf("error restoring cluster for community %s", dir.Name())
			}
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}

	// Subscribe to reachability updates from LibP2P. This lets us know when the node becomes dialable
	// over the public internet. When this occurs, this status is announced via the community state machine.
	sub, err := habitatNode.P2PNode.ReachabilitySubscription()
	if err != nil {
		return nil, err
	}
	go func() {
		for e := range sub.Out() {
			ev := e.(event.EvtLocalReachabilityChanged)

			for communityID, c := range manager.communities {

				_, err := c.ProposeTransitions([]state.CommunityStateTransition{
					&state.ReachabilityTransition{
						NodeID:       manager.nodeID,
						Reachability: ev.Reachability,
					},
				})
				if err != nil {
					log.Error().Err(err).Msgf("error updating node reachability in community %s", communityID)
				}
			}
		}
	}()

	return manager, nil
}

func (m *Manager) setupCommunity(communityID string) (bool, error) {
	path := path.Join(m.Path, communityID)

	// check if community dir already exists
	_, err := os.Stat(path)
	if err == nil {
		return true, fmt.Errorf("data dir for community %s already exists", path)
	}

	// create community dir
	err = os.MkdirAll(path, 0766)
	if err != nil {
		return false, err
	}

	return false, nil
}

func (m *Manager) checkCommunityExists(communityID string) bool {
	path := path.Join(m.Path, communityID)

	_, err := os.Stat(path)
	return err == nil

}

func (m *Manager) CreateCommunity(name string, createIpfs bool, member *community.Member, node *community.Node) (*community.CommunityState, error) {
	// Generate UUID for now
	communityID := uuid.New().String()

	commExists, err := m.setupCommunity(communityID)
	if commExists {
		return nil, fmt.Errorf("can't create community that already exists %s", communityID)
	} else if err != nil {
		return nil, err
	}

	updateChan, err := m.clusterManager.CreateCluster(communityID)
	if err != nil {
		return nil, err
	}

	stateMachine, err := state.NewCommunityStateMachine(community.NewCommunityState(), updateChan, &ClusterDispatcher{
		communityID:    communityID,
		clusterManager: m.clusterManager,
	}, NewCommunityExecutor(m.node))
	if err != nil {
		return nil, err
	}

	err = m.addCommunity(communityID, stateMachine)
	if err != nil {
		return nil, err
	}

	// The first state transition in a new community is alway initialize_community, which sets the community_id
	node.PeerID = m.node.PeerID.String()
	node.Addresses = m.node.Addrs()
	node.Reachability = m.reachability.String()
	transitions := []state.CommunityStateTransition{
		&state.InitializeCommunityTransition{
			CommunityID: communityID,
		},
		&state.AddMemberTransition{
			Member: member,
		},
		&state.AddNodeTransition{
			Node: node,
		},
	}

	if createIpfs {
		// After cluster is created, immediately add transition to initialize IPFS
		// TODO the details of starting this process should be handled by a higher level
		// controller
		ipfsConfig, err := newIPFSSwarm(communityID)
		if err != nil {
			return nil, err
		}

		ipfsPath := filepath.Join(compass.CommunitiesPath(), communityID, "ipfs")

		communityIPFSConfig, err := json.Marshal(ipfsConfig)
		if err != nil {
			return nil, fmt.Errorf("error marshaling IPFS config: %s", err)
		}

		communityIPFSConfigB64 := base64.StdEncoding.EncodeToString(communityIPFSConfig)
		if err != nil {
			return nil, fmt.Errorf("error base64 encoding IPFS config: %s", err)
		}

		procID := procs.RandomProcessID()

		transitions = append(transitions,
			&state.StartProcessTransition{
				Process: &community.Process{
					ID:      procID,
					AppName: "ipfs-driver",
					Args:    []string{ipfsPath},
					Flags:   []string{"-c", communityIPFSConfigB64},
					Env:     []string{},

					Config: ipfsConfig,
				},
			},
			&state.StartProcessInstanceTransition{
				ProcessInstance: &community.ProcessInstance{
					ProcessID: procID,
					NodeID:    node.ID,
				},
			},
		)
	}

	state, err := stateMachine.ProposeTransitions(transitions)
	if err != nil {
		return nil, err
	}

	return state, nil
}

// TODO don't return community state since that is retrieved asynchronously. Or we block
func (m *Manager) JoinCommunity(name string, swarmkey string, btstps []string, acceptingNodeAddr string, communityID string) (*community.CommunityState, error) {
	commExists, err := m.setupCommunity(communityID)
	if err != nil && !commExists {
		return nil, fmt.Errorf("error setting up community: %s", err)
	}

	updateChan, err := m.clusterManager.JoinCluster(communityID, acceptingNodeAddr)
	if err != nil {
		return nil, err
	}

	stateMachine, err := state.NewCommunityStateMachine(community.NewCommunityState(), updateChan, &ClusterDispatcher{
		communityID:    communityID,
		clusterManager: m.clusterManager,
	}, NewCommunityExecutor(m.node))
	if err != nil {
		return nil, err
	}

	err = m.addCommunity(communityID, stateMachine)
	if err != nil {
		return nil, err
	}

	return stateMachine.State()
}

func (m *Manager) AddMemberNode(communityID string, member *community.Member, node *community.Node) (*community.CommunityState, error) {
	stateMachine, ok := m.communities[communityID]
	if !ok {
		return nil, fmt.Errorf("community %s is not on this instance", communityID)
	}

	transitions := []state.CommunityStateTransition{
		&state.AddMemberTransition{
			Member: member,
		},
		&state.AddNodeTransition{
			Node: node,
		},
	}
	state, err := stateMachine.ProposeTransitions(transitions)
	if err != nil {
		return nil, err
	}

	return state, nil
}

func (m *Manager) ProposeTransitions(communityID string, transitions []byte) error {
	if !m.checkCommunityExists(communityID) {
		return fmt.Errorf("community %s does not exist in communities directory", communityID)
	}

	_, err := m.clusterManager.ProposeTransitions(communityID, transitions)
	if err != nil {
		return err
	}

	return nil
}

func (m *Manager) GetState(communityID string) (*community.CommunityState, error) {
	if !m.checkCommunityExists(communityID) {
		return nil, fmt.Errorf("community %s does not exist in communities directory", communityID)
	}

	stateMachine := m.communities[communityID]
	if stateMachine != nil {
		state, err := stateMachine.State()
		if err != nil {
			return nil, err
		}

		return state, nil
	}

	return nil, fmt.Errorf("community state machine for community id %s not found", communityID)
}

func (m *Manager) addCommunity(communityID string, communityState *state.CommunityStateMachine) error {
	m.communitiesLock.Lock()
	if _, ok := m.communities[communityID]; !ok {
		m.communities[communityID] = communityState
		communityState.StartListening()
	} else {
		return fmt.Errorf("community %s is already running", communityID)
	}
	m.communitiesLock.Unlock()

	state, err := communityState.State()
	if err != nil {
		return err
	}
	for _, n := range state.Nodes {
		if n.Reachability == network.ReachabilityPublic.String() {
			err := m.node.P2PNode.AnnounceReachableNode(n)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

//nolint
func (m *Manager) removeCommunity(communityID string) error {
	m.communitiesLock.Lock()
	if _, ok := m.communities[communityID]; ok {
		delete(m.communities, communityID)
	} else {
		return fmt.Errorf("community %s is not running", communityID)
	}
	m.communitiesLock.Unlock()
	return nil
}

type ClusterDispatcher struct {
	communityID    string
	clusterManager *cluster.ClusterManager
}

func (d *ClusterDispatcher) Dispatch(json []byte) (*community.CommunityState, error) {
	encoded := base64.StdEncoding.EncodeToString(json)
	return d.clusterManager.ProposeTransitions(d.communityID, []byte(encoded))
}
