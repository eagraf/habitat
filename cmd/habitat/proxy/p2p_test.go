package proxy

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/eagraf/habitat/pkg/p2p"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

	u, err := url.Parse(redirectedServer.URL)
	assert.Nil(t, err)

	go LibP2PHTTPProxy(p2pNode.Host(), u)

	req, err := http.NewRequest("GET", "", nil)
	assert.Nil(t, err)

	res, err := p2p.LibP2PHTTPRequestWithRandomClient(p2pNode.Addr(), "/hello", p2pNode.Host().ID(), req)
	assert.Nil(t, err)

	slurp, err := ioutil.ReadAll(res.Body)
	require.Nil(t, err)

	assert.Equal(t, "Hello, World!", string(slurp))
}
