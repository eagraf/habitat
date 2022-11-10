package api

import (
	"encoding/json"
	"net/http"

	"github.com/eagraf/habitat/structs/ctl"
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"
)

func WriteError(w http.ResponseWriter, code int, err error) {
	w.WriteHeader(code)
	w.Write([]byte(err.Error() + "\n"))
}

func WriteResponse(w http.ResponseWriter, res interface{}) {
	body, err := json.Marshal(res)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, err)
	}
	w.Write(body)
}

func WriteWebsocketError(conn *websocket.Conn, err error, res ctl.WebsocketMessage) error {
	res.SetError(err)

	err = conn.WriteJSON(res)
	if err != nil {
		return err
	}

	err = conn.WriteMessage(websocket.CloseMessage, []byte{})
	if err != nil {
		return err
	}

	return nil
}

func WriteWebsocketClose(conn *websocket.Conn) {
	err := conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	if err != nil {
		log.Error().Err(err).Msg("error sending websocket close message")
	}

	err = conn.Close()
	if err != nil {
		log.Error().Err(err).Msg("error closing websocket connection")
	}
}
