package main

import (
	"fmt"
	"net/http"
	"time"

	ma "github.com/multiformats/go-multiaddr"

	"github.com/eagraf/habitat/cmd/habitat/api"
	"github.com/eagraf/habitat/cmd/habitat/community"
	"github.com/eagraf/habitat/cmd/habitat/procs"
	"github.com/eagraf/habitat/pkg/p2p"
	"github.com/eagraf/habitat/structs/ctl"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
)

const (
	HabitatCTLHost = "0.0.0.0"
	HabitatCTLPort = "2040"
)

func getRouter(i *Instance) *mux.Router {
	router := mux.NewRouter()
	router.HandleFunc(ctl.GetRoute(ctl.CommandInspect), i.inspectHandler)
	router.HandleFunc(ctl.GetRoute(ctl.CommandListProcesses), i.ProcessManager.ListProcessesHandler)
	router.HandleFunc(ctl.GetRoute(ctl.CommandStart), i.ProcessManager.StartProcessHandler)
	router.HandleFunc(ctl.GetRoute(ctl.CommandStop), i.ProcessManager.StopProcessHandler)

	router.HandleFunc(ctl.GetRoute(ctl.CommandCommunityCreate), i.CommunityManager.CommunityCreateHandler)
	router.HandleFunc(ctl.GetRoute(ctl.CommandCommunityJoin), i.CommunityManager.CommunityJoinHandler)
	router.HandleFunc(ctl.GetRoute(ctl.CommandCommunityAddMember), i.CommunityManager.CommunityAddMemberHandler)
	router.HandleFunc(ctl.GetRoute(ctl.CommandCommunityPropose), i.CommunityManager.CommunityProposeHandler)
	router.HandleFunc(ctl.GetRoute(ctl.CommandCommunityState), i.CommunityManager.CommunityStateHandler)
	router.HandleFunc(ctl.GetRoute(ctl.CommandCommunityList), i.CommunityManager.CommunityListHandler)

	return router
}

func serveHabitatAPI(router *mux.Router) {
	srv := &http.Server{
		Handler:      router,
		Addr:         fmt.Sprintf("%s:%s", HabitatCTLHost, HabitatCTLPort),
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Info().Msgf("starting Habitat API listening on %s", srv.Addr)
	err := srv.ListenAndServe()
	log.Fatal().Err(err)
}

type Instance struct {
	CommunityManager *community.Manager
	ProcessManager   *procs.Manager
	P2PNode          *p2p.Node
}

func (i *Instance) inspectHandler(w http.ResponseWriter, r *http.Request) {
	libp2pProxyMultiaddr, err := ma.NewMultiaddr(i.P2PNode.Host().Addrs()[0].String() + "/p2p/" + i.P2PNode.Host().ID().String())
	if err != nil {
		api.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	res := &ctl.InspectResponse{
		LibP2PProxyMultiaddr: libp2pProxyMultiaddr.String(),
		LibP2PPeerID:         i.P2PNode.Host().ID().String(),
	}

	api.WriteResponse(w, res)
}
