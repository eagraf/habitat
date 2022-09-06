package sources

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"io/ioutil"
	"net/http"
	"os"
	"path"

	"github.com/gorilla/mux"
	"github.com/qri-io/jsonschema"
	"github.com/rs/zerolog/log"
)

/*
	The Schema server is used to implement the $id keyword of the JSON schema spec.
	Specifically, on a sources request for JSON Schema data, reads and writes will be verified against the schema at the $id URL.
	The URL will point to the Schema server.
*/

type SchemaServer struct {
	schemasDir string
}

func NewSchemaServer(path string) *SchemaServer {
	return &SchemaServer{
		schemasDir: path,
	}
}

func (s *SchemaServer) path(hash string) string {
	return path.Join(s.schemasDir, hash+".json")
}

// /get endpoint gets a JSON Schema from the given hash
func (s *SchemaServer) GetHandler(w http.ResponseWriter, r *http.Request) {
	qry := r.URL.Query()
	hash := qry.Get("hash")
	bytes, err := os.ReadFile(s.path(hash))
	if err != nil {
		w.Write([]byte(fmt.Sprintf("error: can't read file at %s: %s", s.path(hash), err.Error())))
		return
	}
	w.Write(bytes)
}

// /validate endpoint validates a JSON object against the given hash
func (s *SchemaServer) ValidateHandler(w http.ResponseWriter, r *http.Request) {
	qry := r.URL.Query()
	hash := qry.Get("hash")
	schema, err := os.ReadFile(s.path(hash))
	if err != nil {
		w.Write([]byte(fmt.Sprintf("error: can't read file at %s: %s", s.path(hash), err.Error())))
		return
	}

	rs := &jsonschema.Schema{}
	if err := json.Unmarshal(schema, rs); err != nil {
		w.Write([]byte(fmt.Sprintf("error: can't unmarshal json at %s: %s", s.path(hash), err.Error())))
		return
	}

	bytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.Write([]byte("error: can't read request body: " + err.Error()))
		return
	}
	errs, err := rs.ValidateBytes(context.Background(), bytes)
	if err != nil {
		w.Write([]byte("error: unable to validate json: " + err.Error()))
		return
	}

	if len(errs) > 0 {
		w.Write([]byte("error: json validation failed: " + errs[0].Error()))
		return
	}

	w.Write([]byte("success!"))
}

func byteSliceEqual(bytes1 []byte, bytes2 []byte) bool {
	if len(bytes1) != len(bytes2) {
		return false
	}
	for i, b := range bytes1 {
		if b != bytes2[i] {
			return false
		}
	}
	return true
}

// /store endpoint saves JSON Schema and returns the hash
func (s *SchemaServer) StoreHandler(w http.ResponseWriter, r *http.Request) {
	bytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.Write([]byte("error: can't read request body: " + err.Error()))
		return
	}
	hash := hashBytes(bytes)
	p := s.path(hash)

	if _, err := os.Stat(p); errors.Is(err, os.ErrNotExist) {
		// path/to/whatever does not exist
		_, err := os.Create(p)
		if err != nil {
			w.Write([]byte(fmt.Sprintf("error: unable to create file at %s: %s", p, err.Error())))
			return
		}
	} else {
		fbytes, err := os.ReadFile(p)
		if err != nil {
			w.Write([]byte(fmt.Sprintf("error: file %s exists and cannot be read: %s", p, err.Error())))
			return
		}
		if !byteSliceEqual(bytes, fbytes) {
			w.Write([]byte(fmt.Sprintf("error: schema exists at path %s exists and is not the same as given schema: %s", p, err.Error())))
			return
		}
	}

	// TODO: is this the right permissions?
	err = os.WriteFile(p, bytes, fs.FileMode(0644))
	if err != nil {
		w.Write([]byte(fmt.Sprintf("error: unable to write to file %s: %s", p, err.Error())))
		return
	}

	w.Write([]byte(hash))
}

// /delete endpoint deletes a JSON Schema
func (s *SchemaServer) DeleteHandler(w http.ResponseWriter, r *http.Request) {
	qry := r.URL.Query()
	hash := qry.Get("hash")
	err := os.Remove(s.path(hash))
	if err != nil {
		w.Write([]byte(fmt.Sprintf("error: unable to delete file at %s: %s", s.path(hash), err.Error())))
		return
	}
	w.Write([]byte("success!"))
}

// getall endpoint returns all the stored JSON Schemas at this server
func (s *SchemaServer) GetAllHandler(w http.ResponseWriter, r *http.Request) {
	files, err := ioutil.ReadDir(s.schemasDir)
	if err != nil {
		w.Write([]byte(fmt.Sprintf("error: unable to read files at %s: %s", s.schemasDir, err.Error())))
		return
	}
	for _, file := range files {
		p := path.Join(s.schemasDir, file.Name())
		bytes, err := os.ReadFile(p)
		if err != nil {
			w.Write([]byte(fmt.Sprintf("error: can't read file at %s: %s", p, err.Error())))
			return
		}
		w.Write(bytes)
		w.Write([]byte("\n"))
	}
}

func (s *SchemaServer) Start(ctx context.Context, port string) {
	err := os.MkdirAll(s.schemasDir, os.ModePerm)
	if err != nil {
		panic("unable to create schema dir: " + err.Error())
	}
	r := mux.NewRouter()
	r.HandleFunc("/get", s.GetHandler)
	r.HandleFunc("/validate", s.ValidateHandler)
	r.HandleFunc("/store", s.StoreHandler)
	r.HandleFunc("/delete", s.DeleteHandler)
	r.HandleFunc("/getall", s.GetAllHandler)
	log.Info().Msgf("Starting schema server on %s", port)
	if err := http.ListenAndServe(port, r); err != nil {
		log.Fatal().Err(err)
	}
}
