package dex

import (
	"encoding/json"
	"net"
	"net/http"
	"os"

	"github.com/rs/zerolog/log"

	"github.com/gorilla/mux"
)

type Server struct {
	sockPath string
	driver   Driver
}

/*
 * Architecture: use unix sockets for IPC between Habitat and DEX driver process
 * Each API endpoint gets its own Unix socket
 * Therefore, multiple concurrent listeners are created, each able to spawn multiple connections
 */
func NewServer(sockPath string, driver Driver) (*Server, error) {
	return &Server{
		sockPath: sockPath,
		driver:   driver,
	}, nil
}

func httpError(w http.ResponseWriter, status int, msg string) {
	w.WriteHeader(status)
	w.Write([]byte(msg))
}

// Handler funcs wrap the Driver implementation
func (s *Server) handleGet(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	cID, ok := vars["content_identifier"]
	if !ok {
		httpError(w, http.StatusBadRequest, "no content identifier provided in request")
		return
	}

	getRes, err := s.driver.Get(cID)
	if err != nil {
		httpError(w, http.StatusInternalServerError, err.Error())
		return
	}

	buf, err := json.Marshal(getRes)
	if err != nil {
		httpError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Write(buf)
}

func (s *Server) handleSchema(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	hash, ok := vars["hash"]
	if !ok {
		httpError(w, http.StatusBadRequest, "no hash provided in request")
		return
	}

	// This should be marshaled JSON

	schemaRes, err := s.driver.Schema(hash)
	if err != nil {
		httpError(w, http.StatusInternalServerError, err.Error())
		return
	}

	buf, err := json.Marshal(schemaRes)
	if err != nil {
		httpError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Write(buf)
}

func (s *Server) handleInterface(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	hash, ok := vars["hash"]
	if !ok {
		httpError(w, http.StatusBadRequest, "no hash provided in request")
		return
	}

	// This should be marshaled JSON

	interfaceRes, err := s.driver.Interface(hash)
	if err != nil {
		httpError(w, http.StatusInternalServerError, err.Error())
		return
	}

	buf, err := json.Marshal(interfaceRes)
	if err != nil {
		httpError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Write(buf)
}

func (s *Server) handleImplementations(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	hash, ok := vars["interface_hash"]
	if !ok {
		httpError(w, http.StatusBadRequest, "no hash provided in request")
		return
	}

	// This should be marshaled JSON

	implementationsRes, err := s.driver.Implementations(hash)
	if err != nil {
		httpError(w, http.StatusInternalServerError, err.Error())
		return
	}

	buf, err := json.Marshal(implementationsRes)
	if err != nil {
		httpError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Write(buf)
}

func (s *Server) Start() {
	err := os.RemoveAll(s.sockPath)
	if err != nil {
		log.Fatal().Msgf("Unable to remove socket path: %s", err)
	}

	r := mux.NewRouter()
	r.HandleFunc("/dex/driver/get/{content_identifier}", s.handleGet)
	r.HandleFunc("/dex/driver/schema/{hash}", s.handleSchema)
	r.HandleFunc("/dex/driver/interface/{hash}", s.handleInterface)
	r.HandleFunc("/dex/driver/implementations/{interface_hash}", s.handleImplementations)
	http.Handle("/", r)

	listener, err := net.Listen("unix", s.sockPath)
	if err != nil {
		log.Fatal().Msgf("failed to start listener: %s", err)
	}
	log.Info().Msgf("DEX driver starting")
	err = http.Serve(listener, r)
	if err != nil {
		log.Fatal().Msgf("error serving DEX driver: %s", err)
	}
}
