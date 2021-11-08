package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"time"

	"github.com/gorilla/mux"
	config "github.com/ipfs/go-ipfs-config"
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
func CreateCommunity(name string, path string) (error, string, string, []string) {

	root := os.Getenv("ROOT")
	// how to get / set env var for root dir?
	cmdCreate := &exec.Cmd{
		Path:   root + "/procs/start.sh",
		Args:   []string{root + "/procs/start.sh", root + "/ipfs/" + path},
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}

	s := make([]string, 0)
	fmt.Println(cmdCreate.String())
	if err := cmdCreate.Run(); err != nil {
		return err, "", "", s
	}

	bytes, _ := ioutil.ReadFile(root + "/ipfs/" + path + "/config")
	var data config.Config
	err := json.Unmarshal(bytes, &data)

	if err != nil {
		return err, "", "", s
	}

	// json struct of config : here we can modify it and write back

	key := KeyGen()
	keyBytes := []byte("/key/swarm/psk/1.0.0/\n/base16/\n" + key + "\n")
	err = ioutil.WriteFile(root+"/ipfs/"+path+"/swarm.key", keyBytes, 0755)

	return nil, key, data.Identity.PeerID, data.Addresses.Swarm
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
		log.Error().Msg("create handler: name argument not suppled")
		return
	}

	err, key, peerid, addrs := CreateCommunity(name, name)

	if err != nil {
		fmt.Println(err.Error())
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
	fmt.Println("returning ", string(bytes))
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
	root := os.Getenv("ROOT")
	// how to get / set env var for root dir?
	cmdJoin := &exec.Cmd{
		Path:   root + "/procs/start.sh",
		Args:   []string{root + "/procs/start.sh", root + "/ipfs/" + path},
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}

	fmt.Println(cmdJoin.String())
	if err := cmdJoin.Run(); err != nil {
		return err, ""
	}

	bytes, _ := ioutil.ReadFile(root + "/ipfs/" + path + "/config")
	var data config.Config
	err := json.Unmarshal(bytes, &data)

	if err != nil {
		return err, ""
	}

	// json struct of config : here we can modify it and write back
	// ignore the peers for now (connect after bootstrapping?)
	data.Bootstrap = btsppeers
	bytes, err = json.Marshal(data)
	log.Info().Msg("data " + string(bytes))
	ioutil.WriteFile(root+"/ipfs/"+path+"/config", bytes, 0755)

	keyBytes := []byte("/key/swarm/psk/1.0.0/\n/base16/\n" + key + "\n")
	err = ioutil.WriteFile(root+"/ipfs/"+path+"/swarm.key", keyBytes, 0755)

	return nil, data.Identity.PeerID
}

func JoinHandler(w http.ResponseWriter, r *http.Request) {
	args := r.URL.Query()
	name := args.Get("name")
	key := args.Get("key")
	addr := args.Get("addr")
	if name == "" || key == "" || addr == "" {
		// error here
		fmt.Errorf("Error: name or key or addr arg is empty string")
		return
	}

	btsppeers := []string{addr}
	err, peerid := JoinCommunity(name, name, key, btsppeers, make([]string, 0))

	if err != nil {
		fmt.Println(err.Error())
	}

	CommConfig := &CommunityConfig{
		Name:           name,
		SwarmKey:       key,
		BootstrapPeers: btsppeers,
		Peers:          []string{peerid},
	}

	fmt.Println("Community Config is ", CommConfig)
	bytes, err := json.Marshal(CommConfig)
	w.Write(bytes)
}

/*
 ConnectCommunity:
 - meant to be used by nodes that are already in a community
 - just run the daemon & return the API or IPFS Client
*/
func ConnectCommunity(name string) error {
	root := os.Getenv("ROOT")
	// how to get / set env var for root dir?
	cmdConnect := &exec.Cmd{
		Path:   root + "/procs/connect.sh",
		Args:   []string{root + "/procs/connect.sh", root + "/ipfs/" + name},
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}
	fmt.Println(cmdConnect.String())
	go cmdConnect.Run()
	/*
		if err := cmdConnect.Run(); err != nil {
			return err
		}
	*/
	return nil
}

func ConnectHandler(w http.ResponseWriter, r *http.Request) {
	args := r.URL.Query()
	name := args.Get("name")
	err := ConnectCommunity(name)
	if err == nil {
		w.Write([]byte("success!"))
	}
}

func main() {
	log.Info().Msg("starting communities api root is" + os.Getenv("ROOT"))

	r := mux.NewRouter()
	// r.HandleFunc("/home", HomeHandler)
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

	log.Info().Msg("communities api listening on localhost:8008")
	log.Fatal().Err(srv.ListenAndServe())
}
