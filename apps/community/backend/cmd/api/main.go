package main

import (
	"context"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
	"github.com/eagraf/habitat/pkg/ipfs"
)

type CommunityConfig struct {
	NodePath string
	SwarmKey string
	Address  string
}

func main() {
	log.Info().Msg("starting notes api")

	r := mux.NewRouter()

	http.Handle("/", r)

	srv := &http.Server{
		Handler:      r,
		Addr:         "0.0.0.0:8000",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Info().Msg("notes api listening on localhost:8000")
	log.Fatal().Err(srv.ListenAndServe())
}

func KeyGen() string {
	return ""
}

// CreateCommunity: create a node with new peers, serves as default bootstrap for community
// 1. Create Node
// 2. Delete all peers
// 3. Run daemon
// 4. get swarm + address to broadcast
func CreateCommunity(name string, path string) (error, string, string) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	api, err := ipfs.createNode(ctx, path, true)
	if err != nil {
		return err, "", ""
	}

	return nil, KeyGen(), api.identity
}

func JoinCommunity() {

}

func ConnectCommunity() {

}
