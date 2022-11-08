package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/eagraf/habitat/pkg/compass"
	client "github.com/eagraf/habitat/pkg/habitat_client"
	"github.com/eagraf/habitat/structs/community"
	"github.com/eagraf/habitat/structs/ctl"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
)

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
func CreateCommunity(name string, id string, path string, createIpfs bool) error {

	createCommunityReq := &ctl.CommunityCreateRequest{
		CommunityName:     name,
		CreateIPFSCluster: createIpfs,
	}

	var comm community.Community
	err, apiErr := client.PostRequest(createCommunityReq, &comm, ctl.GetRoute(ctl.CommandCommunityCreate)) // need to get address from somewhere
	if err != nil {
		log.Error().Err(err).Msg("Unable to send request to habitatctl client")
	} else if apiErr != nil {
		log.Error().Err(err).Msg("api error")
	}

	return nil
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

	err := CreateCommunity(name, args.Get("id"), name, args.Get("ipfs") == "true")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		log.Error().Err(err)
	}

	w.WriteHeader(ctl.StatusOK)
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
func JoinCommunity(name string, path string, key string, btstpaddr string, raftaddr string, commId string) error {
	joinCommunityReq := &ctl.CommunityJoinRequest{
		CommunityID:       commId,
		CommunityName:     name,
		AcceptingNodeAddr: raftaddr,
		SwarmKey:          key,
		BootstrapPeers:    []string{btstpaddr}, // This should include the entire list
	}

	var joinCommunityRes ctl.CommunityJoinResponse
	err, apiErr := client.PostRequest(joinCommunityReq, &joinCommunityRes, ctl.GetRoute(ctl.CommandCommunityJoin)) // need to get address from somewhere
	if err != nil {
		log.Error().Err(err).Msg("Unable to send request to habitatctl client")
	} else if apiErr != nil {
		log.Error().Err(err).Msg("api error")
	}

	return nil
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

	err := JoinCommunity(name, name, key, btstpaddr, raftaddr, comm)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		log.Error().Err(err)
		return
	}

	w.WriteHeader(http.StatusOK)
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

	addMemberReq := &ctl.CommunityAddMemberRequest{
		CommunityID:        comm,
		NodeID:             node,
		JoiningNodeAddress: addr,
	}
	var addRes ctl.CommunityAddMemberResponse
	err, apiErr := client.PostRequest(*addMemberReq, &addRes, ctl.GetRoute(ctl.CommandCommunityAddMember)) // need to get address from somewhere
	if err != nil {
		log.Error().Err(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	if apiErr != nil {
		bytes, err := json.Marshal( // TODO unify this response type with the ctl structs
			&AddMemberResponse{
				MemberId: node,
				NodeId:   compass.NodeID(),
			},
		)
		if err != nil {
			log.Error().Err(err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		w.Write(bytes)
	} else {
		log.Error().Err(err).Msgf("error adding member to community")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
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
