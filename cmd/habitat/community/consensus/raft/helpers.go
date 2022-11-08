package raft

import (
	"fmt"
	"path/filepath"

	"github.com/eagraf/habitat/pkg/compass"
	"github.com/libp2p/go-libp2p/core/protocol"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/rs/zerolog/log"
)

const (
	Localhost  = "localhost"
	DockerHost = "0.0.0.0"
	Protocol   = "http"

	MultiplexerPort = "6000"
	RPCPort         = "6001"
	P2PPort         = "6000"
)

func getMultiplexerAddress() string {
	ip, err := compass.LocalIPv4()
	if err != nil {
		log.Fatal().Msgf("error getting local IP address: %s", err)
	}

	return fmt.Sprintf("%s:%s", ip.String(), MultiplexerPort)
}

// Used by Raft to identify this node in the cluster.
func getServerID(communityID string) string {
	nodeID := compass.NodeID()
	return fmt.Sprintf("%s#%s", nodeID, communityID)
}

// Get the address for a specific Raft instance inside a cluster.
func getCommunityAddress(communityID string) string {
	return fmt.Sprintf("%s://%s/%s", Protocol, getMultiplexerAddress(), communityID)
}

func getCommunityRaftDirectory(communityID string) string {
	return filepath.Join(compass.CommunitiesPath(), communityID, "raft")
}

func getClusterProtocol(communityID string) protocol.ID {
	return protocol.ID(filepath.Join("/habitat-raft", "0.0.1", communityID))
}

func getPublicMultiaddr() (ma.Multiaddr, error) {
	ip, err := compass.PublicIP()
	if err != nil {
		return nil, err
	}
	ipVersion := "ip4"
	if ip.To4() == nil {
		ipVersion = "ip6"
	}
	addr, err := ma.NewMultiaddr(fmt.Sprintf("/%s/%s/tcp/%s", ipVersion, ip.String(), P2PPort))
	if err != nil {
		return nil, err
	}
	return addr, nil
}
