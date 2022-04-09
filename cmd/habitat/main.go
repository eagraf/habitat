package main

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/signal"
	"path/filepath"

	"github.com/eagraf/habitat/cmd/habitat/community"
	"github.com/eagraf/habitat/cmd/habitat/p2p"
	"github.com/eagraf/habitat/cmd/habitat/procs"
	"github.com/eagraf/habitat/cmd/habitat/proxy"
	"github.com/eagraf/habitat/pkg/compass"
	"github.com/eagraf/habitat/structs/configuration"
	"github.com/eagraf/habitat/structs/ctl"
	"github.com/rs/zerolog/log"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	HabitatCTLHost   = "0.0.0.0"
	HabitatCTLPort   = "2040"
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

func listen() {
	// TODO make port number configurable
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%s", HabitatCTLHost, HabitatCTLPort))
	if err != nil {
		log.Fatal().Msgf("habitat service listener failed to start: %s", err)
	}
	defer listener.Close()

	log.Info().Msgf("habitat service listening on %s:%s", HabitatCTLHost, HabitatCTLPort)
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Error().Msgf("error accepting message: %s", err)
		}

		go handleRequest(conn)
	}
}

func handleRequest(conn net.Conn) error {
	buf, err := bufio.NewReader(conn).ReadBytes('\n')
	if err != nil {
		return err
	}

	req, err := decodeRequest(buf)
	if err != nil {
		return writeResponse(conn, &ctl.ResponseWrapper{
			Type:   req.Type,
			Status: ctl.StatusBadRequest,
			Error:  err.Error(),
		})
	}

	res, err := requestRouter(req)
	if err != nil {
		return writeResponse(conn, &ctl.ResponseWrapper{
			Type:   req.Type,
			Status: ctl.StatusInternalServerError,
			Error:  err.Error(),
		})
	}

	return writeResponse(conn, res)
}

func writeResponse(conn net.Conn, res *ctl.ResponseWrapper) error {
	msg, err := res.Encode()
	if err != nil {
		return err
	}

	_, err = conn.Write(msg)
	if err != nil {
		return err
	}

	conn.Close()
	return nil
}

func decodeRequest(buf []byte) (*ctl.RequestWrapper, error) {
	decoded, err := base64.StdEncoding.DecodeString(string(buf))
	if err != nil {
		return nil, err
	}

	var req ctl.RequestWrapper
	err = json.Unmarshal(decoded, &req)
	if err != nil {
		return nil, err
	}

	return &req, nil
}

func requestRouter(req *ctl.RequestWrapper) (*ctl.ResponseWrapper, error) {

	switch req.Type {
	case ctl.CommandStart:
		return ProcessManager.StartProcessHandler(req)
	case ctl.CommandStop:
		return ProcessManager.StopProcessHandler(req)
	case ctl.CommandListProcesses:
		return ProcessManager.ListProcessesHandler(req)
	case ctl.CommandCommunityCreate:
		return CommunityManager.CommunityCreateHandler(req)
	case ctl.CommandCommunityJoin:
		return CommunityManager.CommunityJoinHandler(req)
	case ctl.CommandCommunityAddMember:
		return CommunityManager.CommunityAddMemberHandler(req)
	case ctl.CommandCommunityPropose:
		return CommunityManager.CommunityProposeHandler(req)
	case ctl.CommandCommunityState:
		return CommunityManager.CommunityStateHandler(req)
	case ctl.CommandCommunityList:
		return CommunityManager.CommunityListHandler(req)
	default:
		return &ctl.ResponseWrapper{
			Type:   req.Type,
			Status: ctl.StatusBadRequest,
			Error:  fmt.Sprintf("command %s does not exist", req.Type),
		}, nil
	}
}

func handleInterupt(manager *procs.Manager) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
	manager.StopAllProcesses()
	os.Exit(1)
}
