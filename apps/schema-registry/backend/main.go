package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/eagraf/habitat/cmd/sources"
	"github.com/eagraf/habitat/pkg/compass"
	"github.com/gorilla/mux"
	"github.com/qri-io/jsonschema"
	"github.com/rs/zerolog/log"
)

type SchemaRegistryServer struct {
	Host string
	Port string
	sr   *sources.SchemaRegistry
	// TODO: add local cache
}

func NewSchemaRegistryServer(host string, port string, sr *sources.SchemaRegistry) *SchemaRegistryServer {
	return &SchemaRegistryServer{
		Host: host,
		Port: port,
		sr:   sr,
	}
}

func writeHeaderAndBytes(w http.ResponseWriter, status int, resp string) {
	w.WriteHeader(status)
	w.Write([]byte(resp))
}

func (s *SchemaRegistryServer) LookupHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	surl, err := url.Parse(id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("unable to parse schema url"))
		return
	}

	if surl.Host != "" && surl.Host != s.Host {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("hostname does not match this registry's hostname %s, %s", surl.Host, s.Host)))
		return
	}

	sch, err := s.sr.Lookup(id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	if sch == nil {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("no schema found!"))
		return
	}

	bytes, err := sch.MarshalJSON()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("error marshalling schema: %s", err.Error())))
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(bytes)
	return

}

func (s *SchemaRegistryServer) AddHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		writeHeaderAndBytes(w, http.StatusBadRequest, "id url arg missing")
		return
	}
	slurp, err := ioutil.ReadAll(r.Body)
	if err != nil {
		writeHeaderAndBytes(w, http.StatusBadRequest, "unable to read request body")
		return
	}

	sch := &jsonschema.Schema{}
	if err = sch.UnmarshalJSON(slurp); err != nil {
		writeHeaderAndBytes(w, http.StatusInternalServerError, fmt.Sprintf("unable to unmarshal body json into schema: %s", err.Error()))
		return
	}

	if err = s.sr.Add(id, sch); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	writeHeaderAndBytes(w, http.StatusOK, "success!")
}

func (s *SchemaRegistryServer) DeleteHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		writeHeaderAndBytes(w, http.StatusBadRequest, "id url arg missing")
		return
	}

	// should check id field in schema here too
	if err := s.sr.Delete(id); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	writeHeaderAndBytes(w, http.StatusOK, "success!")
}

func (s *SchemaRegistryServer) Start(ctx context.Context, port string) {
	r := mux.NewRouter()
	r.HandleFunc("/lookup", s.LookupHandler)
	r.HandleFunc("/add", s.AddHandler)
	r.HandleFunc("/delete", s.DeleteHandler)
	log.Info().Msgf("Starting source server on %s", port)
	if err := http.ListenAndServe(port, r); err != nil {
		log.Fatal().Err(err)
	}
}

func main() {
	sr := sources.NewSchemaRegistry(compass.LocalSchemaPath())
	srServer := NewSchemaRegistryServer("localhost", ":8767", sr)
	srServer.Start(context.Background(), ":8767")
}
