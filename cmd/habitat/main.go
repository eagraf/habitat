package main

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"

	"github.com/eagraf/habitat/cmd/habitat/community"
	"github.com/eagraf/habitat/cmd/habitat/p2p"
	"github.com/eagraf/habitat/cmd/habitat/procs"
	"github.com/eagraf/habitat/cmd/habitat/proxy"
	"github.com/eagraf/habitat/pkg/compass"
	"github.com/eagraf/habitat/structs/configuration"
	"github.com/rs/zerolog/log"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	ReverseProxyHost = "0.0.0.0"
	ReverseProxyPort = "2041"

	P2PPort = "6000"
)

// TODO dependency inject this state so we don't use globals
var (
	ProcessManager   *procs.Manager
	CommunityManager *community.Manager
)

func main() {
	pflag.String("hostname", "", "hostname that this node can be reached at")
	pflag.BoolP("docker", "d", false, "use docker host rather than localhost")

	pflag.Parse()
	viper.BindPFlags(pflag.CommandLine)

	procsDir := compass.ProcsPath()
	communityDir := compass.CommunitiesPath()

	_, err := os.Stat(procsDir)
	if err != nil {
		log.Fatal().Msgf("invalid procs directory: %s", err)
	}

	// Get node id
	nodeID := compass.NodeID()
	if err != nil {
		log.Fatal().Msgf("unable to read node ID", err)
	}
	viper.Set("node_id", nodeID)

	// Read apps configuration in proc dir
	appConfigs, err := configuration.ReadAppConfigs(filepath.Join(procsDir, "apps.yml"))
	if err != nil {
		log.Fatal().Msgf("unable to read apps.yml: %s", err)
	}

	p2pNode := p2p.NewNode(P2PPort)

	// Start reverse proxy
	reverseProxy := proxy.NewServer()
	go reverseProxy.Start(fmt.Sprintf("%s:%s", ReverseProxyHost, ReverseProxyPort))

	// Create community manager
	CommunityManager, err = community.NewManager(communityDir, &reverseProxy.Rules, p2pNode.Host())
	if err != nil {
		log.Fatal().Msgf("unable to start community manager: %s", err)
	}

	// Start process manager
	ProcessManager = procs.NewManager(procsDir, reverseProxy.Rules, appConfigs)
	go ProcessManager.ListenForErrors()
	go handleInterupt(ProcessManager)

	listen()
}

func handleInterupt(manager *procs.Manager) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
	manager.StopAllProcesses()
	os.Exit(1)
}
