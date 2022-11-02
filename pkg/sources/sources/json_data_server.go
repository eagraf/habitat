package sources

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
)

type JSONServer struct {
	CommunityID string
	Node        *DataServerNode
	Writer      *Writer
	Reader      *Reader
}

func NewJSONServer(communityID string, host string, port string, w *Writer, r *Reader) *JSONServer {
	return &JSONServer{
		CommunityID: communityID,
		Node: &DataServerNode{
			Host: host,
			Port: port,
		},
		Writer: w,
		Reader: r,
	}
}

func (s *JSONServer) ReadHandler(w http.ResponseWriter, r *http.Request) {
	readreq := &ReadRequest{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(readreq)
	if err != nil {
		w.WriteHeader(400)
		w.Write([]byte("error: bad JSON, can't decode"))
	}

	if readreq.Community != s.CommunityID {
		w.WriteHeader(400)
		w.Write([]byte("error: request community is not the same as this data server's community"))
	}

	allowed, err, data := s.Reader.Read(readreq)
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(fmt.Sprintf("error: error reading from source %s", err.Error())))
		return
	}

	if allowed {
		w.Write([]byte(data))
	} else {
		w.WriteHeader(403)
		w.Write([]byte("error: requester does not have permissions to write to this source"))
	}
}

func (s *JSONServer) WriteHandler(w http.ResponseWriter, r *http.Request) {
	writereq := &WriteRequest{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(writereq)
	if err != nil {
		w.WriteHeader(400)
		w.Write([]byte("error: bad JSON, can't decode"))
	}

	if writereq.CommunityID != s.CommunityID {
		w.WriteHeader(400)
		w.Write([]byte("error: request community is not the same as this data server's community"))
	}

	allowed, err := s.Writer.Write(writereq)
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(fmt.Sprintf("error: error writing to source %s", err.Error())))
		return
	}

	if allowed {
		w.Write([]byte("success!"))
	} else {
		w.WriteHeader(403)
		w.Write([]byte("error: requester does not have permissions to write to this source"))
	}
}

func (s *JSONServer) Start(ctx context.Context) {
	r := mux.NewRouter()
	r.HandleFunc("/read", s.ReadHandler)
	r.HandleFunc("/write", s.WriteHandler)

	addr := s.Node.Host + ":" + s.Node.Port
	log.Info().Msgf("Starting data source server on %s", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatal().Err(err)
	}
}
