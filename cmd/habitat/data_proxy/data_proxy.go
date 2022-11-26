package dataproxy

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/eagraf/habitat/cmd/sources"
	"github.com/eagraf/habitat/pkg/compass"
	"github.com/eagraf/habitat/pkg/permissions"
	"github.com/rs/zerolog/log"

	"github.com/gorilla/mux"
)

type dataType string

const (
	sourcesRequest dataType = "sources"
)

type ReadRequest struct {
	t           dataType        `json:"dataType"`
	communityID string          `json:"communityID"`
	token       string          `json:"token"`
	body        json.RawMessage `json:"body"`
}

type WriteRequest struct {
	t           dataType        `json:"dataType"`
	communityID string          `json:"communityID"`
	token       string          `json:"token"`
	body        json.RawMessage `json:"body"`
	data        []byte          `json:"data"`
}

type DataProxy struct {
	localSchemaRegistryAddr string
	sourcesPermissions      permissions.SourcesPermissionsManager

	appPermissions permissions.AppPermissionsManager
	communityId    string
	// TODO: need localDb / filesystem if this node serves it
	// map of community id to data node
	dataNodes map[string]*httputil.ReverseProxy
}

func writeHeaderAndBytes(w http.ResponseWriter, status int, resp string) {
	w.WriteHeader(status)
	w.Write([]byte(resp))
}

func NewDataProxy(dataNodes map[string]*DataServerNode) *DataProxy {
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
		dataNodes:          proxyNodes,
		appPermissions:     p,
		sourcesPermissions: p,
	}
}

func (s *DataProxy) ReadHandler(w http.ResponseWriter, r *http.Request) {

	var req ReadRequest
	slurp, err := ioutil.ReadAll(r.Body)
	if err != nil {
		writeHeaderAndBytes(w, http.StatusBadRequest, fmt.Sprintf("unable read body: %s", err.Error()))
		return
	}
	err = json.Unmarshal(slurp, &req)
	if err != nil {
		writeHeaderAndBytes(w, http.StatusBadRequest, fmt.Sprintf("unable read unmarshal json: %s", err.Error()))
		return
	}

	if req.communityID != s.communityId {
		if proxy, ok := s.dataNodes[req.communityID]; ok {
			proxy.ServeHTTP(w, r)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("error: could not locate data server for this community %s", req.communityID)))
		}
		return
	}

	switch req.t {
	case sourcesRequest:
		var sreq SourcesRequest
		if err = json.Unmarshal(req.body, &sreq); err != nil {
			writeHeaderAndBytes(w, http.StatusBadRequest, fmt.Sprintf("unable to unmarshal json of sources request: %s", err.Error()))
			return
		}

		s.sourcesPermissions.CheckCanRead(req.token, sreq.sourceID)
		// TODO: pull from a pool of json readers; also need to manage concurrent requests
		reader := sources.NewJSONReader(r.Context(), compass.LocalSourcesPath())

		reader.Read(sources.SourceID(sreq.sourceID))
	}

	writeHeaderAndBytes(w, http.StatusBadRequest, fmt.Sprintf("unrecognized data type: %s", req.t))
}

func (s *DataProxy) WriteHandler(w http.ResponseWriter, r *http.Request) {

	var req WriteRequest
	slurp, err := ioutil.ReadAll(r.Body)
	if err != nil {
		writeHeaderAndBytes(w, http.StatusBadRequest, fmt.Sprintf("unable read body: %s", err.Error()))
		return
	}
	err = json.Unmarshal(slurp, &req)
	if err != nil {
		writeHeaderAndBytes(w, http.StatusBadRequest, fmt.Sprintf("unable read unmarshal json: %s", err.Error()))
		return
	}

	if req.communityID != s.communityId {
		if proxy, ok := s.dataNodes[req.communityID]; ok {
			proxy.ServeHTTP(w, r)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("error: could not locate data server for this community %s", req.communityID)))
		}
		return
	}

	switch req.t {
	case sourcesRequest:
		var sreq SourcesRequest
		if err = json.Unmarshal(req.body, &sreq); err != nil {
			writeHeaderAndBytes(w, http.StatusBadRequest, fmt.Sprintf("unable to unmarshal json of sources request: %s", err.Error()))
			return
		}

		s.sourcesPermissions.CheckCanWrite(req.token, sreq.sourceID)
		// TODO: pull from a pool of json readers; also need to manage concurrent requests
		writer := sources.NewJSONWriter(r.Context(), compass.LocalSourcesPath())

		writer.Write(sources.SourceID(sreq.sourceID), req.data)
	}
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

func (s *DataProxy) Start(ctx context.Context, port string) {
	r := mux.NewRouter()
	r.HandleFunc("/read", s.ReadHandler)
	r.HandleFunc("/write", s.WriteHandler)
	log.Info().Msgf("Starting source server on %s", port)
	if err := http.ListenAndServe(port, r); err != nil {
		log.Fatal().Err(err)
	}
}
