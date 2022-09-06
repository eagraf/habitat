package sources

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/rs/zerolog/log"

	"github.com/gorilla/mux"
)

type SourcesServer struct {
	// serves data stored locally
	localWriter *Writer
	localReader *Reader
	// map of community to data node
	dataNodes map[string]*httputil.ReverseProxy
}

func NewSourcesServer(localReader *Reader, localWriter *Writer, dataNodes map[string]DataServerNode) *SourcesServer {
	proxyNodes := make(map[string]*httputil.ReverseProxy)
	for community, dataNode := range dataNodes {
		url, err := url.Parse("http://" + dataNode.Host + ":" + dataNode.Port)
		if err != nil {
			log.Error().Msgf("error parsing url %s for data node for community %s: %s", "http://"+dataNode.Host+":"+dataNode.Port, community, err.Error())
		}
		proxyNodes[community] = httputil.NewSingleHostReverseProxy(url)
	}
	return &SourcesServer{
		localWriter: localWriter,
		localReader: localReader,
		dataNodes:   proxyNodes,
	}
}

func (s *SourcesServer) ReadHandler(w http.ResponseWriter, r *http.Request) {

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

	readreq := &ReadRequest{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(readreq)
	if err != nil {
		w.Write([]byte("error: bad JSON, can't decode"))
		return
	}

	allowed, err, data := s.localReader.Read(readreq)
	if err != nil {
		w.Write([]byte(fmt.Sprintf("error reading from local sources: %s", err.Error())))
	} else if !allowed {
		w.Write([]byte("error: requester does not have permissions to read from this source"))
	} else {
		w.Write([]byte(data))
	}

}

func (s *SourcesServer) WriteHandler(w http.ResponseWriter, r *http.Request) {
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

	writereq := &WriteRequest{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(writereq)
	if err != nil {
		w.Write([]byte("error: bad JSON, can't decode"))
		return
	}

	allowed, err := s.localWriter.Write(writereq)
	if err != nil {
		w.Write([]byte(fmt.Sprintf("error reading from local sources: %s", err.Error())))
	} else if !allowed {
		w.Write([]byte("error: requester does not have permissions to write to this source"))
	} else {
		w.Write([]byte("success!"))
	}

}

func (s *SourcesServer) Start(ctx context.Context, port string) {
	r := mux.NewRouter()
	r.HandleFunc("/read", s.ReadHandler)
	r.HandleFunc("/write", s.WriteHandler)
	log.Info().Msgf("Starting source server on %s", port)
	if err := http.ListenAndServe(port, r); err != nil {
		log.Fatal().Err(err)
	}
}
