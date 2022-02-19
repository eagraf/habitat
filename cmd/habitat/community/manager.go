package community

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/eagraf/habitat/cmd/habitat/community/consensus/cluster"
	"github.com/eagraf/habitat/cmd/habitat/proxy"
	"github.com/eagraf/habitat/pkg/compass"
	"github.com/eagraf/habitat/pkg/ipfs"
	"github.com/eagraf/habitat/structs/community"
	"github.com/google/uuid"
)

type Manager struct {
	Path           string
	config         *ipfs.IPFSConfig
	clusterManager *cluster.ClusterManager
	communities    []*community.Community
}

func NewManager(path string, proxyRules *proxy.RuleSet) (*Manager, error) {
	clusterManager := cluster.NewClusterManager()

	err := clusterManager.Start(proxyRules)
	if err != nil {
		return nil, fmt.Errorf("error starting cluster manager: %s", err)
	}

	return &Manager{
		Path: path,
		config: &ipfs.IPFSConfig{
			CommunitiesPath: path,
			// TODO: @arushibandi remove this usage of compass
			StartCmd: filepath.Join(compass.ProcsPath(), "bin", "amd64-darwin", "start-ipfs"),
		},
		clusterManager: clusterManager,
	}, nil
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

func (m *Manager) CreateCommunity(name string) (*community.Community, error) {
	// Generate UUID for now
	id := uuid.New()
	communityID := id.String()

	err := m.setupCommunity(communityID)
	if err != nil {
		return nil, err
	}

	err = m.clusterManager.CreateCluster(communityID)
	if err != nil {
		return nil, err
	}

	err, swarmkey, peerid, addrs := m.config.NewCommunityIPFSNode(name, filepath.Join(m.Path, communityID))
	if err != nil {
		return nil, err
	}
	return &community.Community{
		Name:      name,
		Id:        communityID,
		PeerId:    peerid,
		Peers:     []string{},
		SwarmKey:  swarmkey,
		Addresses: addrs,
	}, nil
}

func (m *Manager) JoinCommunity(acceptingNodeAddr string, communityID string) error {
	err := m.setupCommunity(communityID)
	if err != nil {
		return fmt.Errorf("error setting up community: %s", err)
	}

	err = m.clusterManager.JoinCluster(communityID, acceptingNodeAddr)
	if err != nil {
		return err
	}

	return nil
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
