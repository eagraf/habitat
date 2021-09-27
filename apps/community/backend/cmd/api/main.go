package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"time"
	"os/exec"

	"github.com/eagraf/habitat/pkg/ipfs"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
)

//  bootstrap peers are expected to be "always online"
//  does ipfs have a built in way to connect to all peers from the bootstrap?
//  we might need some protocol to do this : get peers from bootstrap node and try to connect
type CommunityConfig struct {
	Name           string
	SwarmKey       string
	BootstrapPeers []string // peer identities of nodes that are bootstrap
	Peers          []string // peer identities of nodes that are just peers
}

// This is a data structure that represents all the communities the user is a part of
type UserCommunities struct {
	Communities []CommunityConfig
}


func main() {
	log.Info().Msg("starting communities api")

	r := mux.NewRouter()
	r.HandleFunc("/home", HomeHandler)
	r.HandleFunc("/create", CreateHandler)
	r.HandleFunc("/join", JoinHandler)
	// at some point this should be abstracted away from user
	// I'm imagining a side panel and when you click on a community name it connects
	r.HandleFunc("/connect", ConnectHandler)
	http.Handle("/", r)

	srv := &http.Server{
		Handler:      r,
		Addr:         "0.0.0.0:8008",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Info().Msg("communities api listening on localhost:8000")
	log.Fatal().Err(srv.ListenAndServe())
}

// from https://github.com/Kubuxu/go-ipfs-swarm-key-gen/blob/master/ipfs-swarm-key-gen/main.go
func KeyGen() string {
	key := make([]byte, 32)
	_, err := rand.Read(key)
	if err != nil {
		fmt.Println("While trying to read random source:", err)
	}

	// when writing to swarm.key, add these to the top:
	// fmt.Println("/key/swarm/psk/1.0.0/")
	// fmt.Println("/base16/")
	return hex.EncodeToString(key)
}

/*
 CreateCommunity: create a node with new peers, serves as default bootstrap for community
 - Create Node
 - Delete all peers
 - Create swarm key and add to swarm.key
 - Run daemon
 - return swarm + address to broadcast
 - this node automatically becomes the bootstrap peer (for now)
 can either use the api returned by createNode or connect to a new client
 TODO: 	create CommunityConfig struct which contains globals for the community like
		swarm key and name of it and peer ids in it
*/
func CreateCommunity(name string, path string) (error, string, string) {

	// how to get / set env var for root dir?
	cmdCreate := &exec.Cmd {
		Path: root + "/procs/bin/" + ostype + "/commstart.sh",
		Args: []string{ bashcmd , path, },
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}

	fmt.Println("Command is ", cmdCreate.String())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	api, err := ipfs.createNode(ctx, path, true) // change this to use CLI NOT coreapi
	if err != nil {
		return err, "", ""
	}

	return nil, KeyGen(), api.identity
}

type CommunityInfo struct {
	Name string 		  `json:"name"`
	Key string 			  `json:"key"`
	BootstrapPeer string  `json:"bootstrap"`
}

func CreateHandler(w http.ResponseWriter, r http.Request) {
	args := r.URL.Query()
	name := args.Get("name")
	if name == nil {
		// error here
		return
	}

	err, key, peerid := CreateCommunity(name, name)

	resComm = &CommunityInfo{
		Name: 			name,
		Key: 			key,
		BootstrapPeer:  peerid
	}
}

/*
 JoinCommunity: the client doesn't have an IPFS node for this network yet:
 - Create Node
 - Delete all peers
 - Add swarm to swarm.key
 - Add bootstrap peers
 - Run daemon
 - Add regular peers
 - Return peer id: need to kick off some way for all other nodes to add this node
  can either use the api returned by createNode or connect to a new client
*/
func JoinCommunity(name string, path string, key string, btsppeers []string, peers []string) string {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	api, err := ipfs.createNode(ctx, path, true)
	if err != nil {
		return err, "", ""
	}

	return nil, api.identity
}

/*
 ConnectCommunity:
 - meant to be used by nodes that are already in a community
 - just run the daemon & return the API or IPFS Client
*/
func ConnectCommunity() {

}
