package community

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/eagraf/habitat/cmd/habitat/api"
	client "github.com/eagraf/habitat/pkg/habitat_client"
	"github.com/eagraf/habitat/pkg/identity"
	"github.com/eagraf/habitat/structs/community"
	"github.com/eagraf/habitat/structs/ctl"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
)

type testWebsocketResponse struct {
	Member *community.Member `json:"member"`
	ctl.WebsocketControl
}

func TestKeySigningExchangeOverWebsocket(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upgrader := &websocket.Upgrader{}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			api.WriteError(w, http.StatusInternalServerError, err)
		}
		defer api.WriteWebsocketClose(conn)

		var res testWebsocketResponse
		member, _, err := signKeyExchange(conn, &res)
		if err != nil {
			// signKeyExchange should already have sent back response
			return
		}

		res.Member = member

		err = conn.WriteJSON(res)
		if err != nil {
			api.WriteWebsocketError(conn, err, &res)
		}
	}))

	defer testServer.Close()

	wsURL, err := url.Parse(testServer.URL)
	assert.Nil(t, err)
	wsURL.Scheme = "ws"

	userIdentity, err := identity.GenerateNewUserCert("bob_ross", "abc")
	assert.Nil(t, err)

	conn, _, err := websocket.DefaultDialer.Dial(wsURL.String(), nil)
	assert.Nil(t, err)

	client.WebsocketKeySigningExchange(conn, userIdentity)

	var res testWebsocketResponse
	err = conn.ReadJSON(&res)
	assert.Nil(t, err)
	assert.Nil(t, res.GetError())
	assert.Equal(t, "bob_ross", string(res.Member.Username))
}

func TestWebsocketError(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upgrader := &websocket.Upgrader{}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			api.WriteError(w, http.StatusInternalServerError, err)
		}
		defer api.WriteWebsocketClose(conn)

		res := &testWebsocketResponse{}
		api.WriteWebsocketError(conn, errors.New("abc"), res)
	}))

	defer testServer.Close()

	wsURL, err := url.Parse(testServer.URL)
	assert.Nil(t, err)
	wsURL.Scheme = "ws"

	conn, _, err := websocket.DefaultDialer.Dial(wsURL.String(), nil)
	assert.Nil(t, err)

	var res testWebsocketResponse
	err = conn.ReadJSON(&res)
	assert.Nil(t, err)

	werr := res.GetError()
	assert.NotNil(t, werr)
	assert.Equal(t, "abc", werr.Error())
}
