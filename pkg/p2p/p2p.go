package p2p

import (
	"fmt"
	"net/http"

	"github.com/eagraf/habitat/pkg/compass"
	libp2p "github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/config"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	ma "github.com/multiformats/go-multiaddr"
)

type Node struct {
	listenAddr ma.Multiaddr
	host       host.Host
}

func NewNode(port string, priv crypto.PrivKey) (*Node, error) {
	ip, err := compass.LocalIPv4()
	if err != nil {
		return nil, err
	}

	listen, _ := ma.NewMultiaddr(fmt.Sprintf("/ip4/%s/tcp/%s", ip, port))
	hostOpts := standardHostOpts(priv, []ma.Multiaddr{listen})
	h, err := libp2p.New(hostOpts...)
	if err != nil {
		return nil, err
	}

	node := &Node{
		listenAddr: listen,
		host:       h,
	}

	return node, nil
}

func (n *Node) Host() host.Host {
	return n.host
}

func (n *Node) PostHTTPRequest(addr ma.Multiaddr, route string, peerID peer.ID, req *http.Request) (*http.Response, error) {
	return libP2PHTTPRequest(addr, n.host, route, peerID, req)
}

func standardHostOpts(priv crypto.PrivKey, listenAddrs []ma.Multiaddr) []config.Option {
	return []config.Option{
		libp2p.ListenAddrs(listenAddrs...),
		libp2p.Identity(priv),
		libp2p.NATPortMap(),
		// TODO @eagraf enable these when experimenting with auto-holepunching
		//	libp2p.Routing(),
		//	libp2p.EnableAutoRelay(),
	}

}
