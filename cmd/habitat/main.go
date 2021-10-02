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
	"strings"

	"github.com/eagraf/habitat/cmd/habitat/community"
	"github.com/eagraf/habitat/cmd/habitat/procs"
	"github.com/eagraf/habitat/cmd/habitat/proxy"
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
)

// TODO dependency inject this state so we don't use globals
var (
	ProcessManager *procs.Manager
)

func main() {
	pflag.String("procdir", "", "directory where process configs are stored")

	pflag.Parse()
	viper.BindPFlags(pflag.CommandLine)

	procDir := viper.GetString("procdir")

	_, err := os.Stat(procDir)
	if err != nil {
		log.Fatal().Msgf("invalid proc directory: %s", err)
	}

	// Read apps configuration in proc dir
	appConfigs, err := configuration.ReadAppConfigs(filepath.Join(procDir, "apps.yml"))
	if err != nil {
		log.Fatal().Msgf("unable to read apps.yml: %s", err)
	}

	// Start reverse proxy
	reverseProxy := proxy.NewServer()
	go reverseProxy.Start(fmt.Sprintf("%s:%s", ReverseProxyHost, ReverseProxyPort))

	// Start process manager
	ProcessManager = procs.NewManager(procDir, reverseProxy.Rules, appConfigs)
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

	fmt.Println(string(buf))

	req, err := decodeRequest(buf)
	if err != nil {
		return writeResponse(conn, &ctl.Response{
			Status:  ctl.StatusBadRequest,
			Message: err.Error(),
		})
	}

	res, err := requestRouter(req)
	if err != nil {
		return writeResponse(conn, &ctl.Response{
			Status:  ctl.StatusInternalServerError,
			Message: err.Error(),
		})
	}

	return writeResponse(conn, res)
}

func writeResponse(conn net.Conn, res *ctl.Response) error {
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

func decodeRequest(buf []byte) (*ctl.Request, error) {
	decoded, err := base64.StdEncoding.DecodeString(string(buf))
	if err != nil {
		return nil, err
	}

	var req ctl.Request
	err = json.Unmarshal(decoded, &req)
	if err != nil {
		return nil, err
	}

	return &req, nil
}

func requestRouter(req *ctl.Request) (*ctl.Response, error) {
	switch req.Command {
	case ctl.CommandStart:
		if len(req.Args) != 1 {
			return nil, fmt.Errorf("start has %d arguments, expected 1", len(req.Args))
		}

		err := ProcessManager.StartProcess(req.Args[0])
		if err != nil {
			return &ctl.Response{
				Status:  ctl.StatusInternalServerError,
				Message: err.Error(),
			}, nil
		}
		fmt.Println(ProcessManager.Procs)

		return &ctl.Response{
			Status:  ctl.StatusOK,
			Message: fmt.Sprintf("started process %s", req.Args[0]),
		}, nil
	case ctl.CommandStop:
		if len(req.Args) != 1 {
			return nil, fmt.Errorf("stop has %d arguments, expected 1", len(req.Args))
		}

		err := ProcessManager.StopProcess(req.Args[0])
		if err != nil {
			return &ctl.Response{
				Status:  ctl.StatusInternalServerError,
				Message: err.Error(),
			}, nil
		}
		fmt.Println(ProcessManager.Procs)

		return &ctl.Response{
			Status:  ctl.StatusOK,
			Message: fmt.Sprintf("stopped process %s", req.Args[0]),
		}, nil

	case ctl.CommandListProcesses:

		procs, err := ProcessManager.ListProcesses()
		if err != nil {
			return &ctl.Response{
				Status:  ctl.StatusInternalServerError,
				Message: err.Error(),
			}, nil
		}

		var b strings.Builder
		for _, p := range procs {
			fmt.Fprintf(&b, "%s\n", p.Name)
		}

		return &ctl.Response{
			Status:  ctl.StatusOK,
			Message: b.String(),
		}, nil

	case ctl.CommandCommunityCreate:
		return community.CommunityCreateHandler(req)
	case ctl.CommandCommunityJoin:
		return community.CommunityJoinHandler(req)
	case ctl.CommandCommunityAddMember:
		return community.CommunityAddMemberHandler(req)
	case ctl.CommandCommunityPropose:
		return community.CommunityProposeHandler(req)
	default:
		return &ctl.Response{
			Status:  ctl.StatusBadRequest,
			Message: fmt.Sprintf("command %s does not exist", req.Command),
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
