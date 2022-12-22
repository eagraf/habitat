package p2p

import (
	"fmt"
	"io"
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

func (n *Node) ConstructMultiAddr() string {
	return n.Addr().String() + "/p2p/" + n.Host().ID().Pretty()
}

func (n *Node) Addr() ma.Multiaddr {
	return n.listenAddr
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

func PostLibP2PRequestToAddress(node *Node, proxyAddr string, route string, req *http.Request) ([]byte, error) {

	peerID, addr, err := compass.LibP2PHabitatAPIAddr(proxyAddr)
	if err != nil {
		return nil, fmt.Errorf("error decomposing multiaddr: %s", err)
	}

	var p2pRes *http.Response
	if node == nil {
		randRes, err := libP2PHTTPRequestWithRandomClient(addr, route, peerID, req)
		if err != nil {
			return nil, err
		}
		p2pRes = randRes
	} else {
		nodeRes, err := node.PostHTTPRequest(addr, route, peerID, req)
		if err != nil {
			return nil, err
		}
		p2pRes = nodeRes
	}

	resBody, err := io.ReadAll(p2pRes.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %s", err)
	}

	// The request was fine, but we got an error back from server
	if p2pRes.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s: %s", p2pRes.Status, string(resBody))
	}

	return resBody, nil
}

func libP2PHTTPRequestWithRandomClient(addr ma.Multiaddr, route string, peerID peer.ID, req *http.Request) (*http.Response, error) {
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

	hostOpts := standardHostOpts(privKey, []ma.Multiaddr{listen})
	h, _ := libp2p.New(hostOpts...)

	return libP2PHTTPRequest(addr, h, route, peerID, req)
}

func libP2PHTTPRequest(addr ma.Multiaddr, host host.Host, route string, peerID peer.ID, req *http.Request) (*http.Response, error) {
	fmt.Println("libp2phttprequest called on ", addr, host, route, peerID, req)
	host.Peerstore().AddAddr(peerID, addr, peerstore.PermanentAddrTTL)

	tr := &http.Transport{}
	tr.RegisterProtocol("libp2p", p2phttp.NewTransport(host))

	client := &http.Client{
		Transport: tr,
	}
	url, err := url.Parse(fmt.Sprintf("libp2p://%s%s", peerID.String(), route))
	if err != nil {
		fmt.Println("error on url parsing")
		return nil, err
	}
	req.URL = url

	return client.Do(req)
}
