package dataproxy

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"github.com/eagraf/habitat/cmd/sources"
	"github.com/eagraf/habitat/pkg/compass"
	"github.com/eagraf/habitat/pkg/permissions"
	"github.com/rs/zerolog/log"

	"github.com/gorilla/mux"
)

type dataType string

const (
	SourcesRequest dataType = "sources"
)

type ReadRequest struct {
	Type        dataType        `json:"data_type"`
	CommunityID string          `json:"community_id"`
	Token       string          `json:"token"`
	Body        json.RawMessage `json:"body"`
}

type ReadResponse struct {
	Error error  `json:"error"`
	Data  []byte `json:"data"`
}

type WriteRequest struct {
	Type        dataType        `json:"data_type"`
	CommunityID string          `json:"community_id"`
	Token       string          `json:"token"`
	Body        json.RawMessage `json:"body"`
	Data        []byte          `json:"data"`
}

type WriteResponse struct {
	Error error `json:"error"`
}

type DataProxy struct {
	// for sources
	schemaStore         *sources.LocalSchemaStore
	localSourcesHandler *sources.JSONReaderWriter
	sourcesPermissions  permissions.SourcesPermissionsManager

	// for app permissions
	appPermissions permissions.AppPermissionsManager
	communityId    string
	// TODO: need localDb / filesystem if this node serves it
	// map of community id to data node
	dataNodes map[string]*httputil.ReverseProxy
}

func NewDataProxy(ctx context.Context, dataNodes map[string]*DataServerNode) *DataProxy {
	proxyNodes := make(map[string]*httputil.ReverseProxy)
	for community, dataNode := range dataNodes {
		url, err := dataNode.GetUrl()
		if err != nil {
			log.Error().Msgf("error parsing url %s for data node for community %s: %s", "http://"+dataNode.Host+":"+dataNode.Port, community, err.Error())
		}
		proxyNodes[community] = httputil.NewSingleHostReverseProxy(url)
	}
	p := permissions.NewBasicPermissionsManager()
	return &DataProxy{
		schemaStore:         sources.NewLocalSchemaStore(compass.LocalSchemaPath()),
		localSourcesHandler: sources.NewJSONReaderWriter(ctx, compass.LocalSourcesPath()),
		dataNodes:           proxyNodes,
		appPermissions:      p,
		sourcesPermissions:  p,
	}
}

func (s *DataProxy) ReadHandler(w http.ResponseWriter, r *http.Request) {

	writeError := func(e error) {
		res := &ReadResponse{
			Error: e,
		}

		b, err := json.Marshal(res)
		if err != nil {
			http.Error(w, fmt.Sprintf("unable to marshal json of sources response: %s", err.Error()), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
		w.Write(b)
	}

	var req ReadRequest
	slurp, err := ioutil.ReadAll(r.Body)
	if err != nil {
		writeError(fmt.Errorf("unable read body: %s", err.Error()))
		return
	}
	err = json.Unmarshal(slurp, &req)
	if err != nil {
		writeError(fmt.Errorf("unable read unmarshal json: %s", err.Error()))
		return
	}

	if req.CommunityID != s.communityId {
		if proxy, ok := s.dataNodes[req.CommunityID]; ok {
			proxy.ServeHTTP(w, r)
		} else {
			writeError(fmt.Errorf("error: could not locate data server for this community %s", req.CommunityID))
		}
		return
	}

	switch req.Type {
	case SourcesRequest:
		// TODO: handle sources that are not stored locally
		var sreq sources.SourceRequest
		if err = json.Unmarshal(req.Body, &sreq); err != nil {
			writeError(fmt.Errorf("unable to unmarshal json of sources request: %s", err.Error()))
			return
		}

		if !s.sourcesPermissions.CheckCanRead(req.Token, sreq.ID) {
			writeError(fmt.Errorf("requester not allowed permission to read source %s", sreq.ID))
			return
		}

		bytes, err := s.localSourcesHandler.Read(sreq.ID)

		if err != nil {
			writeError(fmt.Errorf("error reading source: %s", err.Error()))
			return
		}

		res := ReadResponse{
			Error: nil,
			Data:  bytes,
		}

		b, err := json.Marshal(res)
		if err != nil {
			writeError(fmt.Errorf("unable to marshal json of sources response: %s", err.Error()))
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write(b)
		return
	}

	http.Error(w, fmt.Sprintf("unrecognized data type: %s", req.Type), http.StatusBadRequest)
}

func (s *DataProxy) WriteHandler(w http.ResponseWriter, r *http.Request) {

	writeError := func(e error) {
		res := &WriteResponse{
			Error: e,
		}

		b, err := json.Marshal(res)
		if err != nil {
			http.Error(w, fmt.Sprintf("unable to marshal json of sources response: %s", err.Error()), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
		w.Write(b)
	}

	var req WriteRequest
	slurp, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("unable read body: %s", err.Error()), http.StatusBadRequest)
		return
	}
	err = json.Unmarshal(slurp, &req)
	if err != nil {
		http.Error(w, fmt.Sprintf("unable read unmarshal json: %s", err.Error()), http.StatusBadRequest)
		return
	}

	if req.CommunityID != s.communityId {
		if proxy, ok := s.dataNodes[req.CommunityID]; ok {
			proxy.ServeHTTP(w, r)
		} else {
			writeError(fmt.Errorf("error: could not locate data server for this community %s", req.CommunityID))
		}
		return
	}

	switch req.Type {
	case SourcesRequest:
		// TODO: handle sources that are not stored locally
		var sreq sources.SourceRequest
		if err = json.Unmarshal(req.Body, &sreq); err != nil {
			http.Error(w, fmt.Sprintf("unable to unmarshal json of sources request: %s", err.Error()), http.StatusBadRequest)
			return
		}

		if !s.sourcesPermissions.CheckCanWrite(req.Token, sreq.ID) {
			writeError(fmt.Errorf("requester not allowed permission to write source %s: %s", sreq.ID, err.Error()))
			return
		}

		sch, err := s.schemaStore.Get(sreq.ID)

		if err != nil {
			writeError(fmt.Errorf("error finding schema with id %s: %s", sreq.ID, err.Error()))
			return
		} else if sch == nil {
			// TODO: schema must be explicitly added through schema store: add support in CLI
			// if schema doesn't exist - right now just write it and continue
			sch = s.schemaStore.Resolve(r.Context(), sreq.ID)
			s.schemaStore.Add(sch)
		}

		err = s.localSourcesHandler.Write(sreq.ID, sch, req.Data)
		if err != nil {
			writeError(fmt.Errorf("unable to write sources data: %s", err.Error()))
			return
		}

		res := &WriteResponse{}

		b, err := json.Marshal(res)
		if err != nil {
			writeError(fmt.Errorf("unable to marshal json of sources response: %s", err.Error()))
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write(b)
		return
	}
}

func (s *DataProxy) AddDataNode(communityID string, dataNode *DataServerNode) error {
	if _, found := s.dataNodes[communityID]; found {
		// TODO: allow multiple data nodes per community
		return fmt.Errorf("error: there already exists a data node for this community")
	}

	url, err := url.Parse("http://" + dataNode.Host + ":" + dataNode.Port)
	if err != nil {
		log.Error().Msgf("error parsing url %s for data node for community %s: %s", "http://"+dataNode.Host+":"+dataNode.Port, communityID, err.Error())
	}
	s.dataNodes[communityID] = httputil.NewSingleHostReverseProxy(url)
	return nil
}

func (s *DataProxy) Serve(ctx context.Context, addr string) {

	r := mux.NewRouter()
	r.HandleFunc("/read", s.ReadHandler)
	r.HandleFunc("/write", s.WriteHandler)

	srv := &http.Server{
		Handler:      r,
		Addr:         addr,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Info().Msgf("Starting source server on %s", addr)
	err := srv.ListenAndServe()
	log.Fatal().Err(err)
}
