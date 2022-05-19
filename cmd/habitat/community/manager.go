package community

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"

	"github.com/eagraf/habitat/cmd/habitat/community/consensus/cluster"
	"github.com/eagraf/habitat/cmd/habitat/community/state"
	"github.com/eagraf/habitat/cmd/habitat/proxy"
	"github.com/eagraf/habitat/pkg/compass"
	"github.com/eagraf/habitat/pkg/ipfs"
	"github.com/eagraf/habitat/structs/community"
	"github.com/google/uuid"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/rs/zerolog/log"
)

type Manager struct {
	Path    string
	config  *ipfs.IPFSConfig
	p2pHost host.Host

	clusterManager *cluster.ClusterManager
	communities    []*community.Community
}

func NewManager(path string, proxyRules *proxy.RuleSet, host host.Host) (*Manager, error) {
	clusterManager := cluster.NewClusterManager(host)

	err := clusterManager.Start(proxyRules)
	if err != nil {
		return nil, fmt.Errorf("error starting cluster manager: %s", err)
	}

	// Restart any existing communities
	comDirs, err := ioutil.ReadDir(compass.CommunitiesPath())
	if err == nil {
		for _, dir := range comDirs {
			_, err := clusterManager.RestoreNode(dir.Name())
			if err != nil {
				log.Error().Err(err).Msgf("error restoring cluster for community %s", dir.Name())
			}
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}

	return &Manager{
		Path: path,
		config: &ipfs.IPFSConfig{
			CommunitiesPath: path,
			// TODO: @arushibandi remove this usage of compass
			StartCmd: filepath.Join(compass.ProcsPath(), "bin", "amd64-darwin", "start-ipfs"),
		},
		p2pHost:        host,
		clusterManager: clusterManager,
	}, nil
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

func (m *Manager) CreateCommunity(name string, createIpfs bool) (*community.CommunityState, error) {
	// Generate UUID for now
	communityID := uuid.New().String()

	commExists, err := m.setupCommunity(communityID)
	if commExists {
		return nil, fmt.Errorf("can't create community that already exists %s", communityID)
	} else if err != nil {
		return nil, err
	}

	initState := &community.CommunityState{
		CommunityID: communityID,
	}
	stateBuf, err := json.Marshal(initState)
	if err != nil {
		return nil, err
	}

	stateMachine, err := m.clusterManager.CreateCluster(communityID, stateBuf)
	if err != nil {
		return nil, err
	}

	if createIpfs {
		// After cluster is created, immediately add transition to initialize IPFS
		ipfsConfig, err := newIPFSSwarm(communityID)
		if err != nil {
			return nil, err
		}

		transition := &state.InitializeIPFSSwarmTransition{
			IPFSConfig: ipfsConfig,
		}

		err = stateMachine.ProposeTransition(transition)
		if err != nil {
			return nil, err
		}
	}

	return initState, nil
}

// TODO don't return community state since that is retrieved asynchronously. Or we block
func (m *Manager) JoinCommunity(name string, swarmkey string, btstps []string, acceptingNodeAddr string, communityID string) (*community.CommunityState, error) {
	commExists, err := m.setupCommunity(communityID)
	if err != nil && !commExists {
		return nil, fmt.Errorf("error setting up community: %s", err)
	}

	_, err = m.clusterManager.JoinCluster(communityID, acceptingNodeAddr)
	if err != nil {
		return nil, err
	}

	// TODO @eagraf have this be downstream of a Raft update
	return &community.CommunityState{
		/*	Name:      name,
			Id:        communityID,
			Peers:     btstps,
			SwarmKey:  swarmkey,
			Addresses: []string{},*/
	}, nil
}

func (m *Manager) ProposeTransition(communityID string, transition []byte) error {
	if !m.checkCommunityExists(communityID) {
		return fmt.Errorf("community %s does not exist in communities directory", communityID)
	}

	err := m.clusterManager.ProposeTransition(communityID, transition)
	if err != nil {
		return err
	}

	return nil
}

func (m *Manager) GetState(communityID string) ([]byte, error) {
	if !m.checkCommunityExists(communityID) {
		return nil, fmt.Errorf("community %s does not exist in communities directory", communityID)
	}

	state, err := m.clusterManager.GetState(communityID)
	if err != nil {
		return state, err
	}

	return state, nil
}
