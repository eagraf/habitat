package proxy

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/eagraf/habitat/pkg/p2p"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/stretchr/testify/assert"
)

func TestLibP2PProxy(t *testing.T) {
	// Simulate a server sitting behind the reverse proxy
	redirectedServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Hello, World!")
	}))
	defer redirectedServer.Close()

	serverPrivKey, _, err := crypto.GenerateKeyPair(crypto.Ed25519, 256)
	assert.Nil(t, err)

	p2pNode, err := p2p.NewNode("6660", serverPrivKey)
	assert.Nil(t, err)

	url, err := url.Parse(redirectedServer.URL)
	assert.Nil(t, err)

	go LibP2PHTTPProxy(p2pNode.Host(), url)

	req, err := http.NewRequest("GET", "", nil)
	assert.Nil(t, err)

	res, err := p2p.LibP2PHTTPRequestWithRandomClient(p2pNode.Host().Addrs()[0], "/hello", p2pNode.Host().ID(), req)
	assert.Nil(t, err)

	body, err := io.ReadAll(res.Body)
	assert.Nil(t, err)
	assert.Equal(t, "Hello, World!", string(body))
}
