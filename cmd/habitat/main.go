package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/eagraf/habitat/cmd/habitat/community"
	dataproxy "github.com/eagraf/habitat/cmd/habitat/data_proxy"
	"github.com/eagraf/habitat/cmd/habitat/p2p"
	"github.com/eagraf/habitat/cmd/habitat/procs"
	"github.com/eagraf/habitat/cmd/habitat/proxy"
	"github.com/eagraf/habitat/pkg/compass"
	"github.com/eagraf/habitat/pkg/sources/sources"
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

	initHabitatDirectory()

	p2pNode := p2p.NewNode(P2PPort)

	// Start reverse proxy
	reverseProxy := proxy.NewServer()
	go reverseProxy.Start(fmt.Sprintf("%s:%s", ReverseProxyHost, ReverseProxyPort))

	// Start data proxy
	viper.SetDefault("SOURCES_PORT", ":8765")
	sourcesPort := viper.Get("SOURCES_PORT").(string)
	dataProxy := dataproxy.NewDataProxy(map[string]sources.DataServerNode{})
	go dataProxy.Start(context.Background(), sourcesPort)

	// Start process manager
	ProcessManager = procs.NewManager(procsDir, reverseProxy.Rules)
	go ProcessManager.ListenForErrors()
	go handleInterupt(ProcessManager)

	// Create community manager
	var err error
	CommunityManager, err = community.NewManager(communityDir, ProcessManager, &reverseProxy.Rules, p2pNode.Host())
	if err != nil {
		log.Fatal().Msgf("unable to start community manager: %s", err)
	}

	listen()
}

func initHabitatDirectory() {
	err := os.MkdirAll(compass.CommunitiesPath(), 0700)
	if err != nil {
		log.Fatal().Msgf("unable to create communities directory: %s", err)
	}
	procsDir := compass.ProcsPath()

	_, err = os.Stat(procsDir)
	if err != nil {
		log.Fatal().Msgf("invalid procs directory: %s", err)
	}
	return
}

func handleInterupt(manager *procs.Manager) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
	manager.StopAllProcesses()
	os.Exit(1)
}
