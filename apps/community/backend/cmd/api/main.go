package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strconv"
	"time"

	"github.com/eagraf/habitat/pkg/compass"
	"github.com/eagraf/habitat/pkg/habitatctl/commands"
	"github.com/eagraf/habitat/pkg/ipfs"
	"github.com/eagraf/habitat/structs/community"
	"github.com/eagraf/habitat/structs/ctl"
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

var NodeConfig = &ipfs.IPFSConfig{
	CommunitiesPath: compass.CommunitiesPath(),
	StartCmd:        filepath.Join(compass.HabitatPath(), "apps", "ipfs", "start"),
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
func CreateCommunity(name string, id string, path string, createIpfs bool) ([]byte, error) {
	res, err := commands.SendRequest(ctl.CommandCommunityCreate, []string{name, id, strconv.FormatBool(createIpfs)}) // need to get address from somewhere
	var comm community.Community
	err = json.Unmarshal([]byte(res.Message), &comm)
	if err != nil {
		log.Err(err).Msg(fmt.Sprintf("unable to get community id %s", name))
	}

	if createIpfs {
		time.Sleep(1 * time.Second) // TODO: @arushibandi need to remove this at some point --> basically wait til ipfs comm is created before connecting
		conf, err := ConnectCommunity(name, comm.Id)
		if err != nil {
			log.Err(err).Msg(fmt.Sprintf("unable to connect to community %s", name))
			return nil, err
		}
		bytes, err := json.Marshal(conf)
		return bytes, err
	} else {
		return []byte("did not create ipfs node, only raft"), nil
	}

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

	bytes, err := CreateCommunity(name, args.Get("id"), name, args.Get("ipfs") == "true")
	log.Info().Msg(fmt.Sprintf("Comm is %s", string(bytes)))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		log.Error().Err(err)
	}

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
func JoinCommunity(name string, path string, key string, btstpaddr string, raftaddr string, commId string) ([]byte, error) {
	res, err := commands.SendRequest(ctl.CommandCommunityJoin, []string{name, key, btstpaddr, raftaddr, commId}) // need to get address from somewhere

	var comm community.Community
	err = json.Unmarshal([]byte(res.Message), &comm)
	if err != nil {
		log.Err(err).Msg(fmt.Sprintf("unable to get community id %s", commId))
	}

	time.Sleep(1 * time.Second) // TODO: @arushibandi need to remove this at some point --> basically wait til ipfs comm is created before connecting
	conf, err := ConnectCommunity(name, commId)
	if err != nil {
		return nil, err
	}
	bytes, err := json.Marshal(conf)
	return bytes, err
}

func JoinHandler(w http.ResponseWriter, r *http.Request) {
	args := r.URL.Query()
	name := args.Get("name")
	key := args.Get("key")
	btstpaddr := args.Get("btstpaddr")
	raftaddr := args.Get("raftaddr")
	comm := args.Get("comm")
	if name == "" || key == "" || btstpaddr == "" || raftaddr == "" || comm == "" {
		log.Error().Msg("Error in community join handler: name or key or addr arg is not supplied")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("name or key or addr arg is not supplied"))
		return
	}

	bytes, err := JoinCommunity(name, name, key, btstpaddr, raftaddr, comm)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		log.Error().Err(err)
		return
	}

	w.Write(bytes)
}

type AddMemberResponse struct {
	MemberId string
	NodeId   string
}

// Add Member to a community
// expects node and comm params
func AddMemberHandler(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg(fmt.Sprintf("got request to add node at member %s", compass.NodeID()))
	args := r.URL.Query()
	node := args.Get("node")
	addr := args.Get("address")
	comm := args.Get("comm")
	if node == "" || comm == "" || addr == "" {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Need node id and comm params"))
		return
	}

	res, err := commands.SendRequest(ctl.CommandCommunityAddMember, []string{comm, node, addr}) // need to get address from somewhere
	if err != nil {
		log.Error().Err(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	if res.Status == ctl.StatusOK {
		log.Info().Msg(fmt.Sprintf("successfully added member with message %s", res.Message))
		bytes, err := json.Marshal(
			&AddMemberResponse{
				MemberId: node,
				NodeId:   compass.NodeID(),
			},
		)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			log.Error().Err(err)
			return
		}
		w.Write(bytes)
	} else {
		log.Error().Msg(res.Message)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(res.Message))
	}

}

/*
 ConnectCommunity:
 - meant to be used by nodes that are already in a community
 - just run the daemon & return the API or IPFS Client
*/
func ConnectCommunity(name string, id string) (*ipfs.ConnectedConfig, error) {
	log.Info().Msg(fmt.Sprintf("Connect community called with %s %s", name, id))
	return NodeConfig.ConnectCommunityIPFSNode(name, id)
}

func ConnectHandler(w http.ResponseWriter, r *http.Request) {
	args := r.URL.Query()
	name := args.Get("name")
	id := args.Get("id")
	if name == "" || id == "" {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Need name and id params"))
		return
	}
	conf, err := ConnectCommunity(name, id)
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
	comms := compass.CommunitiesPath()
	communityFiles, err := ioutil.ReadDir(comms)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		log.Error().Err(err)
	} else {
		communityNames := []string{}
		for _, file := range communityFiles {
			communityNames = append(communityNames, file.Name())
		}
		bytes, _ := json.Marshal(&CommunitiesListResponse{Communities: communityNames})
		w.Write(bytes)
	}
}

type NodeIdInfo struct {
	Id string `json:"id"`
}

func NodeIdHandler(w http.ResponseWriter, r *http.Request) {
	info := &NodeIdInfo{Id: compass.NodeID()}
	bytes, err := json.Marshal(info)
	if err == nil {
		w.Write(bytes)
	} else {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		log.Error().Err(err)
	}
}

func main() {
	log.Info().Msg("starting communities api root is" + compass.HabitatPath() + " communnities is " + compass.CommunitiesPath() + "\n ===== Node id is " + compass.NodeID())

	r := mux.NewRouter()
	// r.HandleFunc("/home", HomeHandler)
	r.HandleFunc("/create", CreateHandler)
	r.HandleFunc("/join", JoinHandler)
	// at some point this should be abstracted away from user
	// I'm imagining a side panel and when you click on a community name it connects
	r.HandleFunc("/connect", ConnectHandler)
	r.HandleFunc("/communities", GetCommunitiesHandler)
	r.HandleFunc("/add", AddMemberHandler)
	r.HandleFunc("/node", NodeIdHandler)
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
