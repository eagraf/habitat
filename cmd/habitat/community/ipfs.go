package community

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"

	"github.com/eagraf/habitat/pkg/compass"
	"github.com/eagraf/habitat/pkg/ipfs"
	"github.com/eagraf/habitat/structs/community"
)

// TODO refactor this into the ipfs-driver proc
//nolint
func newIPFSSwarm(communityID string) (*community.IPFSConfig, error) {
	ipfsPath := filepath.Join(compass.CommunitiesPath(), communityID, "ipfs")
	err := os.MkdirAll(ipfsPath, 0700)
	if err != nil {
		return nil, err
	}
	ipfsInstance := &ipfs.IPFSInstance{
		IPFSPath: ipfsPath,
	}
	swarmKey, err := ipfsInstance.GenerateSwarmKey()
	if err != nil {
		return nil, err
	}

	// first write a config
	err = ipfsInstance.Init()
	if err != nil {
		return nil, err
	}

	config, err := ipfsInstance.Config()
	if err != nil {
		return nil, err
	}

	// Set the bootstrap peers to this nodes swarm addresses
	// This node will become the only bootstrap peer available
	config.Bootstrap = config.Addresses.Swarm

	// json struct of config : here we can modify it and write back
	// ignore the peers for now (connect after bootstrapping?)
	// choose a random port to be used by ipfs config (so that having)
	// multiple ipfs instances doesn't cause a conflict
	addr := config.Addresses.API
	addrstring := addr[0]
	parts := strings.Split(addrstring, "/")
	parts[len(parts)-1] = fmt.Sprint(rand.Intn(65536-9999) + 9999)
	config.Addresses.API = []string{strings.Join(parts, "/")}

	err = ipfsInstance.Configure(config)
	if err != nil {
		return nil, err
	}

	return &community.IPFSConfig{
		SwarmKey:           swarmKey,
		BootstrapAddresses: config.Bootstrap,
	}, nil
}
