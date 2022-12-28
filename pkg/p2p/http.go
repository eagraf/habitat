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
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"
	ma "github.com/multiformats/go-multiaddr"
)

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

	hostOpts := standardHostOpts(privKey, []ma.Multiaddr{listen})
	h, _ := libp2p.New(hostOpts...)

	return libP2PHTTPRequest(addr, h, route, peerID, req)
}

func libP2PHTTPRequest(addr ma.Multiaddr, host host.Host, route string, peerID peer.ID, req *http.Request) (*http.Response, error) {
	host.Peerstore().AddAddr(peerID, addr, peerstore.PermanentAddrTTL)

	tr := &http.Transport{}
	tr.RegisterProtocol("libp2p", p2phttp.NewTransport(host))

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
