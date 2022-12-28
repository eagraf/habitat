package main

import (
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
	pflag.StringP("p2p-port", "p", "6000", "port used by the libp2p host")

	pflag.Parse()
	viper.BindPFlags(pflag.CommandLine)

	initHabitatDirectory()

	p2pPort := viper.GetString("p2p-port")

	node, err := node.NewNode(p2pPort)
	if err != nil {
		log.Fatal().Err(err).Msgf("error starting Habitat node")
	}

	go node.Start()

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
