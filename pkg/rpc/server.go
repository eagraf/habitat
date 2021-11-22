package rpc

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/rs/zerolog/log"
)

// For now make things simple with only one layer of routes
type Server struct {
	handlers map[string]RPCHandlerFunc
}

type RPCHandlerFunc func([]byte) (int, []byte)

type RPCRequest struct {
	Data []byte `json:"data"`
}

type RPCResponse struct {
	StatusCode int    `json:"status_code"`
	Response   []byte `json:"response"`
}

func NewServer(handlers map[string]RPCHandlerFunc) *Server {
	return &Server{
		handlers: handlers,
	}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	for route, handler := range s.handlers {
		if route == r.URL.Path {
			HandlerWrapper(handler)(w, r)
			return
		}
	}
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte(fmt.Sprintf("route %s not found", r.URL.Path)))
}

func (s *Server) Start(host string) {
	log.Info().Msgf("rpc server listening on %s", host)
	log.Fatal().Err(http.ListenAndServe(host, s)).Msg("RPC server failed")
}

func HandlerWrapper(fn RPCHandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Error().Err(err).Msg("error reading RPC HTTP body")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("error reading RPC HTTP body"))
		}

		code, resp := fn(body)

		rpcResp := &RPCResponse{
			StatusCode: code,
			Response:   resp,
		}

		buf, err := json.Marshal(rpcResp)
		if err != nil {
			log.Error().Err(err).Msg("error marshaling RPC response")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("error marshaling RPC response"))
		}

		w.WriteHeader(http.StatusOK)
		w.Write(buf)
	}
}
