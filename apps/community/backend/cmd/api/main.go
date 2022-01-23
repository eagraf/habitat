package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/eagraf/habitat/pkg/compass"
	"github.com/eagraf/habitat/pkg/ipfs"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
)

//  bootstrap peers are expected to be "always online"
//  does ipfs have a built in way to connect to all peers from the bootstrap?
//  we might need some protocol to do this : get peers from bootstrap node and try to connect
type CommunityConfig struct {
	Name           string   `json:"name"`
	SwarmKey       string   `json:"swarm_key"`
	BootstrapPeers []string `json:"btstp_peers"` // addresses of nodes that are bootstrap
	Peers          []string `json:"peers"`       // peer ids of nodes
}

// This is a data structure that represents all the communities the user is a part of
type UserCommunities struct {
	Communities []CommunityConfig
}

/*
 CreateCommunity: create a node with new peers, serves as default bootstrap for community
 - Create Node
 - Delete all peers
 - Create swarm key and add to swarm.key
 - Run daemon
 - return error, secret key, peerid,swarm address to broadcast
 - this node automatically becomes the bootstrap peer (for now)
 can either use the api returned by createNode or connect to a new client
 TODO: 	create CommunityConfig struct which contains globals for the community like
		swarm key and name of it and peer ids in it
*/
func CreateCommunity(name string, path string) (error, string, string, []string) {
	return ipfs.NewCommunityIPFSNode(name, path)
}

type CommunityInfo struct {
	Name          string `json:"name"`
	Key           string `json:"key"`
	BootstrapPeer string `json:"bootstrap"`
}

func CreateHandler(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("Create handler called")
	args := r.URL.Query()
	name := args.Get("name")
	if name == "" {
		// error here
		log.Error().Msg("Error in community create handler: name argument not suppled")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("no name argument supplied in the request"))
		return
	}

	err, key, peerid, addrs := CreateCommunity(name, name)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		log.Error().Err(err)
	}

	CommConfig := &CommunityConfig{
		Name:           name,
		SwarmKey:       key,
		BootstrapPeers: addrs,
		Peers:          []string{peerid},
	}

	commstr, _ := json.Marshal(CommConfig)
	log.Info().Msg("Community Config is " + string(commstr))
	bytes, err := json.Marshal(*CommConfig)
	w.Write(bytes)
}

/*
 JoinCommunity: the client doesn't have an IPFS node for this network yet:
 - Create Node
 - Delete all peers
 - Add swarm to swarm.key
 - Add bootstrap peers
 - Run daemon
 - Add regular peers ?
 - Return peer id: need to kick off some way for all other nodes to add this node
  can either use the api returned by createNode or connect to a new client
*/
func JoinCommunity(name string, path string, key string, btsppeers []string, peers []string) (error, string) {
	return ipfs.JoinCommunityIPFSNode(name, path, key, btsppeers, peers)
}

func JoinHandler(w http.ResponseWriter, r *http.Request) {
	args := r.URL.Query()
	name := args.Get("name")
	key := args.Get("key")
	addr := args.Get("addr")
	if name == "" || key == "" || addr == "" {
		log.Error().Msg("Error in community join handler: name or key or addr arg is not supplied")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("name or key or addr arg is not supplied"))
		return
	}

	btsppeers := []string{addr}
	err, peerid := JoinCommunity(name, name, key, btsppeers, make([]string, 0))

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		log.Error().Err(err)
	}

	CommConfig := &CommunityConfig{
		Name:           name,
		SwarmKey:       key,
		BootstrapPeers: btsppeers,
		Peers:          []string{peerid},
	}

	bytes, err := json.Marshal(CommConfig)
	w.Write(bytes)
}

/*
 ConnectCommunity:
 - meant to be used by nodes that are already in a community
 - just run the daemon & return the API or IPFS Client
*/
func ConnectCommunity(name string) (ipfs.ConnectedConfig, error) {
	return ipfs.ConnectCommunityIPFSNode(name)
}

func ConnectHandler(w http.ResponseWriter, r *http.Request) {
	args := r.URL.Query()
	name := args.Get("name")
	conf, err := ConnectCommunity(name)
	if err == nil {
		bytes, _ := json.Marshal(conf)
		w.Write(bytes)
	} else {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		log.Error().Err(err)
	}
}

type CommunitiesListResponse struct {
	Communities []string `json:"Communities"`
}

/*
 GetCommunities:
 - return all communities in user's data/ipfs folder
*/
func GetCommunitiesHandler(w http.ResponseWriter, r *http.Request) {
	root := compass.HabitatPath()
	ipfsDir, err := os.Open(filepath.Join(root, "/data/ipfs/"))
	fmt.Println("Get communities from path ", root, "/data/ipfs/")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		log.Error().Err(err)
		return
	}
	defer ipfsDir.Close()
	communityNames, err := ipfsDir.Readdirnames(0)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		log.Error().Err(err)
	} else {
		fmt.Println("returning :", communityNames)
		bytes, _ := json.Marshal(&CommunitiesListResponse{Communities: communityNames})
		w.Write(bytes)
	}
}

func main() {
	log.Info().Msg("starting communities api root is" + compass.HabitatPath() + " communnities is " + compass.CommunitiesPath())

	r := mux.NewRouter()
	// r.HandleFunc("/home", HomeHandler)
	r.HandleFunc("/create", CreateHandler)
	r.HandleFunc("/join", JoinHandler)
	// at some point this should be abstracted away from user
	// I'm imagining a side panel and when you click on a community name it connects
	r.HandleFunc("/connect", ConnectHandler)
	r.HandleFunc("/communities", GetCommunitiesHandler)
	http.Handle("/", r)

	srv := &http.Server{
		Handler:      r,
		Addr:         "0.0.0.0:8008",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Info().Msg("communities api listening on localhost:8008")
	log.Fatal().Err(srv.ListenAndServe())
}
