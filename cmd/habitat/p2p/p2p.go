package p2p

import (
	"fmt"

	"github.com/eagraf/habitat/pkg/compass"
	libp2p "github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/rs/zerolog/log"
)

type Node struct {
	listenAddr ma.Multiaddr
	host       host.Host
}

func NewNode(port string, priv crypto.PrivKey) *Node {
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

	log.Info().Msgf("starting libp2p node listening at %v, with peer identity %s", h.Addrs(), h.ID().Pretty())

	return node
}

func (n *Node) Host() host.Host {
	return n.host
}
