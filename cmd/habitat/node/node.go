package node

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"

	"github.com/eagraf/habitat/cmd/habitat/api"
	dataproxy "github.com/eagraf/habitat/cmd/habitat/data_proxy"
	"github.com/eagraf/habitat/cmd/habitat/node/dex"
	"github.com/eagraf/habitat/cmd/habitat/node/fs"
	"github.com/eagraf/habitat/cmd/habitat/procs"
	"github.com/eagraf/habitat/cmd/habitat/proxy"
	"github.com/eagraf/habitat/pkg/compass"
	"github.com/eagraf/habitat/pkg/ipfs"
	"github.com/eagraf/habitat/pkg/p2p"
	"github.com/eagraf/habitat/structs/ctl"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/rs/zerolog/log"
)

const (
	ReverseProxyHost = "0.0.0.0"
	ReverseProxyPort = "2041"
	SourcesPort      = "8765"

	P2PPort = "6000"

	IPFSAPIURL = "http://localhost:5001/api/v0"
)

type Node struct {
	ID string

	ProcessManager *procs.Manager
	P2PNode        *p2p.Node
	ReverseProxy   *proxy.Server
	DataProxy      *dataproxy.DataProxy
	FS             *fs.FS
	DexNodeAPI     *dex.DexNodeAPI

	// Temporary IPFS
	IPFSClient *ipfs.Client
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
	dataproxy := dataproxy.NewDataProxy(context.Background(), p2pNode, map[string]*dataproxy.DataServerNode{})

	ipfsClient, err := ipfs.NewClient(IPFSAPIURL)
	if err != nil {
		return nil, err
	}

	fs := fs.NewFS(ipfsClient)
	dexNodeAPI := dex.NewDexNodeAPI(ipfsClient)

	return &Node{
		ID:             compass.NodeID(),
		P2PNode:        p2pNode,
		ReverseProxy:   reverseProxy,
		DataProxy:      dataproxy,
		ProcessManager: procs.NewManager(procsDir, reverseProxy.Rules),
		FS:             fs,
		IPFSClient:     ipfsClient,
		DexNodeAPI:     dexNodeAPI,
	}, nil
}

func (n *Node) Start() error {
	// Start reverse proxy
	proxyAddr := fmt.Sprintf("%s:%s", ReverseProxyHost, ReverseProxyPort)
	go n.ReverseProxy.Start(proxyAddr)

	redirectURL, err := url.Parse("http://" + proxyAddr)
	if err != nil {
		log.Fatal().Err(err)
	}
	go proxy.LibP2PHTTPProxy(n.P2PNode.Host(), redirectURL)

	// Forward data reads
	ustr := compass.DefaultHabitatAPIAddr()
	u, err := url.Parse(ustr)
	if err != nil {
		log.Fatal().Err(err)
	}

	n.ReverseProxy.Rules.Add("data-proxy", &proxy.RedirectRule{
		Matcher:         "/data_read",
		ForwardLocation: u,
	})

	n.ReverseProxy.Rules.Add("data-proxy", &proxy.RedirectRule{
		Matcher:         "/data_write",
		ForwardLocation: u,
	})

	// Start process manager
	go n.ProcessManager.ListenForErrors()
	go handleInterupt(n.ProcessManager)

	// start IPFS grand-child process
	// Note that this is temporary, and we will likely eventually move away from IPFS
	ipfsPath := filepath.Join(compass.HabitatPath(), "ipfs")
	_, err = n.ProcessManager.StartProcessInstance("", "ipfs", "ipfs-driver", []string{ipfsPath}, []string{}, []string{})
	if err != nil {
		log.Fatal().Err(err).Msg("error start IPFS grandchild process")
	}

	// Keep this thread running
	ctx := context.Background()
	<-ctx.Done()

	return nil
}

func (n *Node) LibP2PProxyMultiaddr() (ma.Multiaddr, error) {
	return ma.NewMultiaddr(n.P2PNode.Host().Addrs()[0].String() + "/p2p/" + n.P2PNode.Host().ID().String())
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
