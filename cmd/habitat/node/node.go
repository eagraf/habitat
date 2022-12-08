package node

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/signal"

	"github.com/eagraf/habitat/cmd/habitat/api"
	dataproxy "github.com/eagraf/habitat/cmd/habitat/data_proxy"
	"github.com/eagraf/habitat/cmd/habitat/procs"
	"github.com/eagraf/habitat/cmd/habitat/proxy"
	"github.com/eagraf/habitat/cmd/sources"
	"github.com/eagraf/habitat/pkg/compass"
	"github.com/eagraf/habitat/pkg/p2p"
	"github.com/eagraf/habitat/structs/ctl"
	"github.com/libp2p/go-libp2p/core/peer"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

const (
	ReverseProxyHost = "0.0.0.0"
	ReverseProxyPort = "2041"

	P2PPort = "6000"
)

type Node struct {
	ID     string
	PeerID peer.ID

	ProcessManager *procs.Manager
	P2PNode        *p2p.Node
	ReverseProxy   *proxy.Server
	DataProxy      *dataproxy.DataProxy
}

func NewNode() (*Node, error) {
	priv, _ := compass.GetPeerIDKeyPair()

	p2pNode, err := p2p.NewNode(P2PPort, priv)
	if err != nil {
		return nil, err
	}
	log.Info().Msgf("starting LibP2P node with peer ID %s listening at port %s", p2pNode.Host().ID().Pretty(), P2PPort)

	procsDir := compass.ProcsPath()
	reverseProxy := proxy.NewServer()

	return &Node{
		ID:             compass.NodeID(),
		PeerID:         p2pNode.Host().ID(),
		P2PNode:        p2pNode,
		ReverseProxy:   reverseProxy,
		DataProxy:      dataproxy.NewDataProxy(map[string]*sources.DataServerNode{}),
		ProcessManager: procs.NewManager(procsDir, reverseProxy.Rules),
	}, nil
}

func (n *Node) Start() error {
	// Start reverse proxy
	proxyAddr := fmt.Sprintf("%s:%s", ReverseProxyHost, ReverseProxyPort)
	go n.ReverseProxy.Start(proxyAddr)

	redirectURL, err := url.Parse("http://" + proxyAddr + "/habitat")
	if err != nil {
		log.Fatal().Err(err)
	}
	go proxy.LibP2PHTTPProxy(n.P2PNode.Host(), redirectURL)

	// Start data proxy
	sourcesPort := viper.GetString("data-proxy-port")
	go n.DataProxy.Start(context.Background(), sourcesPort)

	// Start process manager
	go n.ProcessManager.ListenForErrors()
	go handleInterupt(n.ProcessManager)

	// Keep this thread running
	ctx := context.Background()
	<-ctx.Done()

	return nil
}

func (n *Node) LibP2PProxyMultiaddr() (ma.Multiaddr, error) {
	return ma.NewMultiaddr(n.P2PNode.Host().Addrs()[0].String() + "/p2p/" + n.P2PNode.Host().ID().String())
}

func (n *Node) Addrs() []string {
	multiAddrs := n.P2PNode.Host().Addrs()
	res := []string{}
	for _, m := range multiAddrs {
		res = append(res, m.String())
	}
	return res
}

func (n *Node) InspectHandler(w http.ResponseWriter, r *http.Request) {
	libp2pProxyMultiaddr, err := n.LibP2PProxyMultiaddr()
	if err != nil {
		api.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	res := &ctl.InspectResponse{
		LibP2PProxyMultiaddr: libp2pProxyMultiaddr.String(),
		LibP2PPeerID:         n.P2PNode.Host().ID().String(),
	}

	api.WriteResponse(w, res)
}

func handleInterupt(manager *procs.Manager) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
	manager.StopAllProcesses()
	os.Exit(1)
}
