package dataproxy

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/eagraf/habitat/cmd/sources"
	"github.com/rs/zerolog/log"

	"github.com/gorilla/mux"
)

type DataProxy struct {
	// map of community id to data node
	dataNodes map[string]*httputil.ReverseProxy
}

func NewDataProxy(dataNodes map[string]*sources.DataServerNode) *DataProxy {
	proxyNodes := make(map[string]*httputil.ReverseProxy)
	for community, dataNode := range dataNodes {
		url, err := dataNode.GetUrl()
		if err != nil {
			log.Error().Msgf("error parsing url %s for data node for community %s: %s", "http://"+dataNode.Host+":"+dataNode.Port, community, err.Error())
		}
		proxyNodes[community] = httputil.NewSingleHostReverseProxy(url)
	}
	return &DataProxy{
		dataNodes: proxyNodes,
	}
}

func (s *DataProxy) ReadHandler(w http.ResponseWriter, r *http.Request) {

	qry := r.URL.Query()
	community := qry.Get("community")
	if community != "" {
		if proxy, ok := s.dataNodes[community]; ok {
			proxy.ServeHTTP(w, r)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("error: could not locate data server for this community %s", community)))
		}
		return
	}
}

func (s *DataProxy) WriteHandler(w http.ResponseWriter, r *http.Request) {
	qry := r.URL.Query()
	community := qry.Get("community")
	if community != "" {
		if proxy, ok := s.dataNodes[community]; ok {
			proxy.ServeHTTP(w, r)
		} else {
			w.Write([]byte(fmt.Sprintf("error: could not locate data server for this community %s", community)))
		}
		return
	}
}

func (s *DataProxy) AddDataNode(communityID string, dataNode sources.DataServerNode) error {
	if _, found := s.dataNodes[communityID]; found {
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
