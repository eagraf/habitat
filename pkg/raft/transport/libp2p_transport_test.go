package transport

import (
	"bufio"
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/eagraf/habitat/pkg/compass"
	"github.com/libp2p/go-libp2p"
	crypto "github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/peerstore"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/rs/zerolog/log"
)

// helper method - create a lib-p2p host to listen on a port
func makeRandomHost(ip string, port int) host.Host {
	// Ignoring most errors for brevity
	// See echo example for more details and better implementation
	priv, _, _ := crypto.GenerateKeyPair(crypto.Secp256k1, 256)
	listen, _ := ma.NewMultiaddr(fmt.Sprintf("/ip4/%s/tcp/%d", ip, port))
	host, _ := libp2p.New(
		libp2p.ListenAddrs(listen),
		libp2p.Identity(priv),
	)

	return host
}

type node struct {
	host.Host
	remoteID peer.ID
}

func (n *node) mockHandler(stream network.Stream) {
	// Create a buffer stream for non blocking read and write.
	rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))
	_, err := rw.ReadString('\n')
	if err != nil {
		log.Fatal().Err(err)
	}
	_, err = rw.Write([]byte("yooooo\n"))
	if err != nil {
		log.Fatal().Err(err)
	}
	rw.Flush()
}

func TestStuff(t *testing.T) {
	ip, err := compass.LocalIPv4()
	if err != nil {
		t.Error(err)
	}

	h1 := makeRandomHost(ip.String(), 6000)
	h2 := makeRandomHost(ip.String(), 6001)
	n2 := &node{
		Host:     h2,
		remoteID: h1.ID(),
	}

	h2.SetStreamHandler("/fake-protocol", n2.mockHandler)

	_, err = h1.NewStream(context.Background(), h2.ID(), "/fake-protocol")
	if err != nil {
		t.Error(err)
	}
}

func TestLibP2PConn(t *testing.T) {
	ip, err := compass.LocalIPv4()
	if err != nil {
		t.Error(err)
	}

	h1 := makeRandomHost(ip.String(), 6000)
	h2 := makeRandomHost(ip.String(), 6001)
	n2 := &node{
		Host:     h2,
		remoteID: h1.ID(),
	}

	h2.SetStreamHandler("/fake-protocol", n2.mockHandler)

	h1.Peerstore().AddAddr(h2.ID(), h2.Addrs()[0], peerstore.PermanentAddrTTL)
	h2.Peerstore().AddAddr(h1.ID(), h1.Addrs()[0], peerstore.PermanentAddrTTL)

	stream, err := h1.NewStream(context.Background(), h2.ID(), "/fake-protocol")
	if err != nil {
		t.Error(err)
	}

	rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))

	_, err = rw.Write([]byte("helloo\n"))
	if err != nil {
		t.Error(err)
	}
	rw.Flush()
	nn, err := rw.ReadString('\n')
	if err != nil {
		t.Error(err)
	}
	log.Info().Msgf("response btytes %s", nn)
	stream.Close()

	time.Sleep(5 * time.Second)
}
