package transport

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/hashicorp/raft"
	"github.com/rs/zerolog/log"
)

type RaftMultiplexer struct {
	listeners map[string]*HTTPTransport
}

func NewRaftMultiplexer() *RaftMultiplexer {
	return &RaftMultiplexer{
		listeners: make(map[string]*HTTPTransport),
	}
}

func (rm *RaftMultiplexer) RegisterListener(communityID string, transport *HTTPTransport) error {
	if _, ok := rm.listeners[communityID]; ok {
		return fmt.Errorf("community %s is already a registered listener", communityID)
	}

	rm.listeners[communityID] = transport
	return nil
}

func (rm *RaftMultiplexer) Listen(addr string) {
	r := mux.NewRouter()
	r.HandleFunc("/{community_id}", rm.handler)

	srv := &http.Server{
		Handler:      r,
		Addr:         addr,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Info().Msgf("raft multiplexer listening on %s", addr)
	log.Fatal().Err(srv.ListenAndServe())
}

func writeError(w http.ResponseWriter, status int, msg string) {
	log.Error().Msg(msg)
	w.WriteHeader(status)
	w.Write([]byte(msg))
}

func (rm *RaftMultiplexer) handler(w http.ResponseWriter, r *http.Request) {
	communityID, ok := mux.Vars(r)["community_id"]
	if !ok {
		writeError(w, http.StatusBadRequest, "community_id missing from raft request parameters")
		return
	}

	if _, ok := rm.listeners[communityID]; !ok {
		writeError(w, http.StatusNotFound, fmt.Sprintf("community_id %s is not a listener", communityID))
		return
	}
	transport := rm.listeners[communityID]
	if transport == nil {
		writeError(w, http.StatusInternalServerError, (fmt.Sprintf("transport is nil for community %s", communityID)))
		return
	}

	buf, err := ioutil.ReadAll(r.Body)
	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("unable to read req body: %s", err))
		return
	}

	var reqBody RaftRequest
	err = json.Unmarshal(buf, &reqBody)
	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("unable to unmarshal req body: %s", err))
		return
	}

	command, err := unmarshalRPCRequest(reqBody.RPCType, reqBody.Args)
	if err != nil {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("error unmarshaling command: %s", err))
		return
	}

	// If we got an AppendEntriesRequest, check if it is a heartbeat
	isHeartbeat := false
	if reqBody.RPCType == rpcAppendEntries {
		aeReq := command.(*raft.AppendEntriesRequest)
		if aeReq.Term != 0 && aeReq.Leader != nil &&
			aeReq.PrevLogEntry == 0 && aeReq.PrevLogTerm == 0 &&
			len(aeReq.Entries) == 0 && aeReq.LeaderCommitIndex == 0 {
			isHeartbeat = true

		}
	}

	respChan := make(chan raft.RPCResponse, 1)
	consumeCh := transport.consumeCh

	rpc := &raft.RPC{
		Command: command,
		// TODO figure out the reader part
		RespChan: respChan,
	}

	if isHeartbeat {
		transport.heartbeatFnLock.Lock()
		fn := transport.heartbeatFn
		transport.heartbeatFnLock.Unlock()
		if fn != nil {
			fn(*rpc)
			goto RESP
		}
	}

	consumeCh <- *rpc

RESP:
	// Wait on response channel to respond to request
	resp := <-respChan

	if resp.Error != nil {
		// TODO find a more appropriate error code
		writeError(w, http.StatusBadRequest, resp.Error.Error())
		return
	}

	// Serialize raft response and serialize into HTTP response
	buf, err = json.Marshal(resp.Response)
	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("error marshaling raft response: %s", err))
		return
	}

	httpRespBody := &RaftResponse{
		RPCType: reqBody.RPCType,
		Resp:    buf,
	}

	marshaledBody, err := json.Marshal(httpRespBody)
	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("error marshaling http response body: %s", err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(marshaledBody))
}
