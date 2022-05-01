package cluster

import (
	"errors"
	"io/ioutil"
	"net/url"
	"os"

	"github.com/eagraf/habitat/cmd/habitat/community/consensus/raft"
	"github.com/eagraf/habitat/cmd/habitat/proxy"
	"github.com/eagraf/habitat/pkg/compass"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/rs/zerolog/log"
)

type ClusterService interface {
	Start() error

	CreateCluster(communityID string, initState []byte) error
	RemoveCluster(communityID string) error
	JoinCluster(communityID string, address string) error
	RestoreNode(communityID string) error

	ProposeTransition(communityID string, transition []byte) error
	GetState(communityID string) ([]byte, error)

	AddNode(communityID string, nodeID string, address string) error
	RemoveNode(communityID string, nodeID string) error
}

// ClusterManager is a layer of abstraction that allows multiple Cluster Services to share
// a common interface.
// TODO Right now, the only cluster service implementation is Raft, so the switching implementation
// needs to be implemented in the future.
type ClusterManager struct {
	raftClusterService *raft.ClusterService
}

func NewClusterManager(host host.Host) *ClusterManager {
	raft := raft.NewClusterService(host)
	return &ClusterManager{
		raftClusterService: raft,
	}
}

func (cm *ClusterManager) Start(proxyRules *proxy.RuleSet) error {
	// TODO fix this when centralized port management is implemented
	url, err := url.Parse("http://0.0.0.0:6000/raft/msg")
	if err != nil {
		return err
	}

	// TODO switch between implementations of cluster services
	proxyRules.Add("raft-service", &proxy.RedirectRule{
		Matcher:         "/raft/msg",
		ForwardLocation: url,
	})

	err = cm.raftClusterService.Start()
	if err != nil {
		return err
	}

	// Restart any existing communities
	comDirs, err := ioutil.ReadDir(compass.CommunitiesPath())
	if err == nil {
		for _, dir := range comDirs {
			err := cm.RestoreNode(dir.Name())
			if err != nil {
				log.Error().Err(err).Msgf("error restoring cluster for community %s", dir.Name())
			}
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}

	return nil
}

func (cm *ClusterManager) CreateCluster(communityID string, initState []byte) error {
	return cm.raftClusterService.CreateCluster(communityID, initState)
}

func (cm *ClusterManager) RemoveCluster(communityID string) error {
	return cm.raftClusterService.RemoveCluster(communityID)

}

func (cm *ClusterManager) JoinCluster(communityID string, address string) error {
	return cm.raftClusterService.JoinCluster(communityID, address)
}

func (cm *ClusterManager) RestoreNode(communityID string) error {
	return cm.raftClusterService.RestoreNode(communityID)
}

func (cm *ClusterManager) ProposeTransition(communityID string, transition []byte) error {
	return cm.raftClusterService.ProposeTransition(communityID, transition)
}

func (cm *ClusterManager) GetState(communityID string) ([]byte, error) {
	return cm.raftClusterService.GetState(communityID)
}

func (cm *ClusterManager) AddNode(communityID string, nodeID string, address string) error {
	return cm.raftClusterService.AddNode(communityID, nodeID, address)
}

func (cm *ClusterManager) RemoveNode(communityID string, nodeID string) error {
	return cm.raftClusterService.RemoveNode(communityID, nodeID)
}
