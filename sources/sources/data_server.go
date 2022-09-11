package sources

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

type DataServer struct {
	// TODO: fill this in
	Community string
	Node      *DataServerNode
	Writer    *Writer
	Reader    *Reader
}

type DataServerNode struct {
	Host string
	Port string
}

func NewDataServer(community string, host string, port string, w *Writer, r *Reader) *DataServer {
	return &DataServer{
		Community: community,
		Node: &DataServerNode{
			Host: host,
			Port: port,
		},
		Writer: w,
		Reader: r,
	}
}

func (s *DataServer) ReadHandler(w http.ResponseWriter, r *http.Request) {
	readreq := &ReadRequest{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(readreq)
	if err != nil {
		w.Write([]byte("error: bad JSON, can't decode"))
	}

	if readreq.Community != s.Community {
		w.Write([]byte("error: request community is not the same as this data server's community"))
	}

	allowed, err, data := s.Reader.Read(readreq)
	if err != nil {
		w.Write([]byte(fmt.Sprintf("error: error reading from source %s", err.Error())))
	}

	if allowed {
		w.Write([]byte(data))
	} else {
		w.Write([]byte("error: requester does not have permissions to write to this source"))
	}
}

func (s *DataServer) WriteHandler(w http.ResponseWriter, r *http.Request) {
	writereq := &WriteRequest{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(writereq)
	if err != nil {
		w.Write([]byte("error: bad JSON, can't decode"))
	}

	if writereq.Community != s.Community {
		w.Write([]byte("error: request community is not the same as this data server's community"))
	}

	allowed, err := s.Writer.Write(writereq)
	if err != nil {
		w.Write([]byte(fmt.Sprintf("error: error writing to source %s", err.Error())))

	}

	if allowed {
		w.Write([]byte("success!"))
	} else {
		w.Write([]byte("error: requester does not have permissions to write to this source"))
	}
}

func (s *DataServer) Start() {
	r := mux.NewRouter()
	r.HandleFunc("/read", s.ReadHandler)
	r.HandleFunc("/write", s.WriteHandler)
}
