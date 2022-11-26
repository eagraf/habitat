package main

import (
	"context"
	"net/url"
	"os"

	"github.com/eagraf/habitat/cmd/habitat/community"
	"github.com/eagraf/habitat/cmd/habitat/node"
	"github.com/eagraf/habitat/cmd/habitat/proxy"
	"github.com/eagraf/habitat/pkg/compass"
	"github.com/rs/zerolog/log"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func main() {
	pflag.String("hostname", "", "hostname that this node can be reached at")
	pflag.BoolP("docker", "d", false, "use docker host rather than localhost")
	pflag.String("data-proxy-port", ":8675", "port used by the data proxy server")

	pflag.Parse()
	viper.BindPFlags(pflag.CommandLine)

	initHabitatDirectory()

	node, err := node.NewNode()
	if err != nil {
		log.Fatal().Err(err).Msgf("error starting Habitat node")
	}

	go node.Start()
	// Start data proxy
	viper.SetDefault("SOURCES_PORT", ":8765")
	sourcesPort := viper.Get("SOURCES_PORT").(string)
	dataProxy := dataproxy.NewDataProxy(map[string]*dataproxy.DataServerNode{})
	go dataProxy.Start(context.Background(), sourcesPort)

	// Create community manager
	communityDir := compass.CommunitiesPath()
	CommunityManager, err := community.NewManager(communityDir, node)
	if err != nil {
		log.Fatal().Msgf("unable to start community manager: %s", err)
	}

	apiURL, err := url.Parse("http://0.0.0.0:2040")
	if err != nil {
		log.Fatal().Msgf("unable to get url for Habitat API: %s", err)
	}
	node.ReverseProxy.Rules.Add("habitat-api", &proxy.RedirectRule{
		Matcher:         "/habitat",
		ForwardLocation: apiURL,
	})

	router := getRouter(node, CommunityManager)

	serveHabitatAPI(router)
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
