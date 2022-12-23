package dataproxy

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"github.com/eagraf/habitat/cmd/habitat/api"
	"github.com/eagraf/habitat/cmd/sources"
	"github.com/eagraf/habitat/pkg/compass"
	"github.com/eagraf/habitat/pkg/p2p"
	"github.com/eagraf/habitat/pkg/permissions"
	"github.com/eagraf/habitat/structs/ctl"
	"github.com/qri-io/jsonschema"
	"github.com/rs/zerolog/log"

	"github.com/gorilla/mux"
)

type nodeType string

// TODO: we should really pull these out to the top level node
const (
	LocalNode     nodeType = "local"
	CommunityNode nodeType = "community"
)

type DataProxy struct {
	// for sources
	schemaStore         *sources.LocalSchemaStore
	localSourcesHandler *sources.JSONReaderWriter
	sourcesPermissions  permissions.SourcesPermissionsManager

	// for app permissions
	appPermissions permissions.AppPermissionsManager
	nodeType       nodeType
	communityId    string
	nodeId         string
	// TODO: need localDb / filesystem if this node serves it
	// map of community id to data node
	dataNodes map[string]*httputil.ReverseProxy
	peers     map[string]string
	p2pNode   *p2p.Node
}

func NewDataProxy(ctx context.Context, p2pNode *p2p.Node, dataNodes map[string]*DataServerNode) *DataProxy {
	proxyNodes := make(map[string]*httputil.ReverseProxy)
	for community, dataNode := range dataNodes {
		url, err := dataNode.GetUrl()
		if err != nil {
			log.Error().Msgf("error parsing url %s for data node for community %s: %s", "http://"+dataNode.Host+":"+dataNode.Port, community, err.Error())
		}
		proxyNodes[community] = httputil.NewSingleHostReverseProxy(url)
	}
	p := permissions.NewBasicPermissionsManager()

	// TODO: for tests we should really pass in a mock p2pNode instead of doing this hacky stuff
	nodeId := ""
	if p2pNode != nil {
		nodeId = p2pNode.Host().ID().String()
	}

	return &DataProxy{
		schemaStore:         sources.NewLocalSchemaStore(compass.LocalSchemaPath()),
		localSourcesHandler: sources.NewJSONReaderWriter(ctx, compass.LocalSourcesPath()),
		dataNodes:           proxyNodes,
		nodeId:              nodeId,
		p2pNode:             p2pNode,
		peers:               make(map[string]string),
		appPermissions:      p,
		sourcesPermissions:  p,
	}
}

func (s *DataProxy) ReadHandler(w http.ResponseWriter, r *http.Request) {

	var req ctl.DataReadRequest
	slurp, err := ioutil.ReadAll(r.Body)
	r.Body.Close()

	fmt.Printf("got read request with body %s\n", string(slurp))

	if err != nil {
		api.WriteError(w, http.StatusBadRequest, fmt.Errorf("unable read body: %s", err.Error()))
		return
	}
	err = json.Unmarshal(slurp, &req)
	if err != nil {
		api.WriteError(w, http.StatusBadRequest, fmt.Errorf("unable read unmarshal json: %s", err.Error()))
		return
	}

	if s.nodeType == CommunityNode && req.CommunityID != s.communityId {
		if proxy, ok := s.dataNodes[req.CommunityID]; ok {
			proxy.ServeHTTP(w, r)
		} else {
			api.WriteError(w, http.StatusBadRequest, fmt.Errorf("error: could not locate data server for this community %s", req.CommunityID))
		}
		return
	}

	// leave node id field empty to request local data
	if req.NodeID != "" && req.NodeID != s.nodeId {
		// need to use libp2p to talk to other data server
		naddr, ok := s.peers[req.NodeID]
		if !ok {
			api.WriteError(w, http.StatusInternalServerError, fmt.Errorf("cannot find p2p node for peer with node id %s", req.NodeID))
			return
		}
		if err != nil {
			api.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error getting habitat libp2p addr: %s", err.Error()))
			return
		}

		p2pReq, err := http.NewRequest("POST", "", bytes.NewReader(slurp))
		if err != nil {
			api.WriteError(w, http.StatusInternalServerError, fmt.Errorf("unable to create POST request to forward: %s", err.Error()))
		}

		bytes, err := p2p.PostLibP2PRequestToAddress(s.p2pNode, naddr, "/data_read", p2pReq)
		if err != nil {
			api.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error forwarding read request to other dataproxy: %s", err.Error()))
			return
		}

		res := ctl.DataReadResponse{
			Data: bytes,
		}

		api.WriteResponse(w, res)
		return
	}

	switch req.Type {
	case ctl.SourcesRequest:
		// TODO: handle sources that are not stored locally
		var sreq sources.SourceRequest
		if err = json.Unmarshal(req.Body, &sreq); err != nil {
			api.WriteError(w, http.StatusBadRequest, fmt.Errorf("unable to unmarshal json of sources request: %s", err.Error()))
			return
		}

		if !s.sourcesPermissions.CheckCanRead(req.Token, sreq.ID) {
			api.WriteError(w, http.StatusMethodNotAllowed, fmt.Errorf("requester not allowed permission to read source %s", sreq.ID))
			return
		}

		bytes, err := s.localSourcesHandler.Read(sreq.ID)

		if err != nil {
			api.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error reading source: %s", err.Error()))
			return
		}

		res := ctl.DataReadResponse{
			Data: bytes,
		}

		api.WriteResponse(w, res)
		return
	}

	http.Error(w, fmt.Sprintf("unrecognized data type: %s", req.Type), http.StatusBadRequest)
}

func (s *DataProxy) WriteHandler(w http.ResponseWriter, r *http.Request) {

	var req ctl.DataWriteRequest
	slurp, err := ioutil.ReadAll(r.Body)
	if err != nil {
		api.WriteError(w, http.StatusBadRequest, fmt.Errorf("unable read body: %s", err.Error()))
		return
	}

	err = json.Unmarshal(slurp, &req)
	if err != nil {
		api.WriteError(w, http.StatusBadRequest, fmt.Errorf("unable read unmarshal json: %s", err.Error()))
		return
	}

	if req.CommunityID != s.communityId {
		if proxy, ok := s.dataNodes[req.CommunityID]; ok {
			proxy.ServeHTTP(w, r)
		} else {
			api.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error: could not locate data server for this community %s", req.CommunityID))
		}
		return
	}

	switch req.Type {
	case ctl.SourcesRequest:
		// TODO: handle sources that are not stored locally
		var sreq sources.SourceRequest
		if err = json.Unmarshal(req.Body, &sreq); err != nil {
			api.WriteError(w, http.StatusBadRequest, fmt.Errorf("unable to unmarshal json of sources request: %s", err.Error()))
			return
		}

		if !s.sourcesPermissions.CheckCanWrite(req.Token, sreq.ID) {
			api.WriteError(w, http.StatusMethodNotAllowed, fmt.Errorf("requester not allowed permission to write source %s: %s", sreq.ID, err.Error()))
			return
		}

		sch, err := s.schemaStore.Get(sreq.ID)

		if err != nil {
			api.WriteError(w, http.StatusInternalServerError, fmt.Errorf("error finding schema with id %s: %s", sreq.ID, err.Error()))
			return
		} else if sch == nil {
			// TODO: schema must be explicitly added through schema store: add support in CLI
			// if schema doesn't exist - right now just write it and continue
			sch = s.schemaStore.Resolve(r.Context(), sreq.ID)
			sch.Schema = jsonschema.GetSchemaRegistry().Get(context.Background(), sreq.ID)
			if sch.Schema == nil {
				api.WriteError(w, http.StatusInternalServerError, fmt.Errorf("unable to resolve schema for $id: %s", sreq.ID))
			}
			s.schemaStore.Add(sch)
			// TODO: here need to resolve the schema and add it to the local allow list if the user grants permission
			// for right now just always resolve and add the schema
		}

		err = s.localSourcesHandler.Write(sreq.ID, sch, req.Data)
		if err != nil {
			api.WriteError(w, http.StatusInternalServerError, fmt.Errorf("unable to write sources data: %s", err.Error()))
			return
		}

		api.WriteResponse(w, &ctl.DataWriteResponse{})
		return
	}
}

func (s *DataProxy) AddSchemaHandler(w http.ResponseWriter, r *http.Request) {
	return
}

func (s *DataProxy) LookupSchemaHandler(w http.ResponseWriter, r *http.Request) {
	return
}

func (s *DataProxy) DeleteSchemaHandler(w http.ResponseWriter, r *http.Request) {
	return
}

func (s *DataProxy) AddDataNode(communityID string, dataNode DataServerNode) error {
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

func (s *DataProxy) AddPeerNode(nodeId string, addr string) {
	s.peers[nodeId] = addr
}

func (s *DataProxy) Serve(ctx context.Context, addr string) {

	r := mux.NewRouter()
	r.HandleFunc("/read_source", s.ReadHandler)
	r.HandleFunc("/write_source", s.WriteHandler)
	r.HandleFunc("/add_schema", s.AddSchemaHandler)
	r.HandleFunc("/lookup_schema", s.LookupSchemaHandler)
	r.HandleFunc("/delete_schema", s.DeleteSchemaHandler)

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
