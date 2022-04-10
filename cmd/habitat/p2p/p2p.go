package p2p

import (
	"fmt"

	"github.com/eagraf/habitat/pkg/compass"
	libp2p "github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/host"
	ma "github.com/multiformats/go-multiaddr"
)

const clientVersion = "cmd/habitat/p2p/0.0.1"

type Node struct {
	listenAddr ma.Multiaddr
	host       host.Host
}

func NewNode(port string) *Node {
	priv, _ := compass.GetPeerIDKeyPair()

	ip, err := compass.LocalIPv4()
	if err != nil {
		panic(err)
	}

	listen, _ := ma.NewMultiaddr(fmt.Sprintf("/ip4/%s/tcp/%s", ip, port))
	h, _ := libp2p.New(
		libp2p.ListenAddrs(listen),
		libp2p.Identity(priv),
		libp2p.NATPortMap(),
	)
	node := &Node{
		listenAddr: listen,
		host:       h,
	}

	return node
}

func (n *Node) Host() host.Host {
	return n.host
}
