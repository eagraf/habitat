package p2p

import (
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"

	"github.com/eagraf/habitat/pkg/compass"
	libp2p "github.com/libp2p/go-libp2p"
	p2phttp "github.com/libp2p/go-libp2p-http"
	"github.com/libp2p/go-libp2p/config"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"
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

func LibP2PHTTPRequestWithRandomClient(addr ma.Multiaddr, route string, peerID peer.ID, req *http.Request) (*http.Response, error) {
	// generate a temporary host to make the request
	ip, err := compass.LocalIPv4()
	if err != nil {
		return nil, err
	}

	// use a random port number - yuck!
	port := strconv.Itoa(5000 + rand.Intn(5000))

	privKey, _, err := crypto.GenerateKeyPair(crypto.Ed25519, 256)
	if err != nil {
		return nil, err
	}

	// TODO @eagraf handle IPv6
	listen, _ := ma.NewMultiaddr(fmt.Sprintf("/ip4/%s/tcp/%s", ip, port))

	fmt.Println(addr.String())

	hostOpts := standardHostOpts(privKey, []ma.Multiaddr{listen})
	h, _ := libp2p.New(hostOpts...)

	h.Peerstore().AddAddr(peerID, addr, peerstore.PermanentAddrTTL)

	tr := &http.Transport{}
	tr.RegisterProtocol("libp2p", p2phttp.NewTransport(h))

	client := &http.Client{
		Transport: tr,
	}
	url, err := url.Parse(fmt.Sprintf("libp2p://%s%s", peerID.String(), route))
	if err != nil {
		return nil, err
	}
	req.URL = url

	return client.Do(req)
}
