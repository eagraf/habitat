package p2p

import (
	"context"
	"fmt"
	"net/http"

	"github.com/eagraf/habitat/pkg/compass"
	"github.com/eagraf/habitat/structs/community"
	libp2p "github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/config"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/event"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/host/autorelay"
	ma "github.com/multiformats/go-multiaddr"
)

type Node struct {
	listenAddr ma.Multiaddr
	host       host.Host

	// When reachable peers join a community this node participates in
	// they are fed through this channel so that libp2p can attempt to
	// create a latent relay connection with them.
	peerChan chan<- peer.AddrInfo
}

func NewNode(port string, priv crypto.PrivKey) (*Node, error) {
	ip, err := compass.LocalIPv4()
	if err != nil {
		return nil, err
	}

	peerChan := make(chan peer.AddrInfo)

	listen, _ := ma.NewMultiaddr(fmt.Sprintf("/ip4/%s/tcp/%s", ip, port))
	hostOpts := append(standardHostOpts(priv, []ma.Multiaddr{listen}), relayHostOpts(peerChan)...)
	h, err := libp2p.New(hostOpts...)
	if err != nil {
		return nil, err
	}

	node := &Node{
		listenAddr: listen,
		host:       h,
		peerChan:   peerChan,
	}

	return node, nil
}

func (n *Node) Host() host.Host {
	return n.host
}

func (n *Node) ReachabilitySubscription() (event.Subscription, error) {
	sub, err := n.host.EventBus().Subscribe(new(event.EvtLocalReachabilityChanged))
	if err != nil {
		return nil, err
	}
	return sub, nil
}

func (n *Node) AnnounceReachableNode(node *community.Node) error {
	addrInfo, err := node.AddrInfo()
	if err != nil {
		return err
	}
	n.reachablePeerChan <- *addrInfo
	return nil
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
	}
}

func relayHostOpts(peerChan chan peer.AddrInfo) []config.Option {
	return []config.Option{
		//libp2p.Routing(),
		libp2p.EnableHolePunching(),
		libp2p.EnableRelay(),
		libp2p.EnableAutoRelay(autorelay.WithPeerSource(func(ctx context.Context, numPeers int) <-chan peer.AddrInfo {
			r := make(chan peer.AddrInfo)
			go func() {
				defer close(r)
				for ; numPeers != 0; numPeers-- {
					select {
					case v, ok := <-peerChan:
						if !ok {
							return
						}
						select {
						case r <- v:
						case <-ctx.Done():
							return
						}
					case <-ctx.Done():
						return
					}
				}
			}()
			return r
		}, 0)),
	}
}
