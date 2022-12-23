package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/eagraf/habitat/cmd/habitat/community"
	"github.com/eagraf/habitat/cmd/habitat/node"
	"github.com/eagraf/habitat/structs/ctl"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
)

const (
	HabitatCTLHost = "0.0.0.0"
	HabitatCTLPort = "2040"
)

func getRouter(n *node.Node, cm *community.Manager) *mux.Router {
	router := mux.NewRouter()
	router.HandleFunc(ctl.GetRoute(ctl.CommandInspect), n.InspectHandler)
	router.HandleFunc(ctl.GetRoute(ctl.CommandListProcesses), n.ProcessManager.ListProcessesHandler)
	router.HandleFunc(ctl.GetRoute(ctl.CommandStart), n.ProcessManager.StartProcessHandler)
	router.HandleFunc(ctl.GetRoute(ctl.CommandStop), n.ProcessManager.StopProcessHandler)

	router.HandleFunc(ctl.GetRoute(ctl.CommandCommunityCreate), cm.CommunityCreateHandler)
	router.HandleFunc(ctl.GetRoute(ctl.CommandCommunityJoin), cm.CommunityJoinHandler)
	router.HandleFunc(ctl.GetRoute(ctl.CommandCommunityAddMember), cm.CommunityAddMemberHandler)
	router.HandleFunc(ctl.GetRoute(ctl.CommandCommunityPropose), cm.CommunityProposeHandler)
	router.HandleFunc(ctl.GetRoute(ctl.CommandCommunityState), cm.CommunityStateHandler)
	router.HandleFunc(ctl.GetRoute(ctl.CommandCommunityList), cm.CommunityListHandler)
	router.HandleFunc(ctl.GetRoute(ctl.CommandCommunityPS), cm.CommunityPSHandler)
	router.HandleFunc(ctl.GetRoute(ctl.CommandCommunityStartProcess), cm.CommunityStartProcessHandler)
	router.HandleFunc(ctl.GetRoute(ctl.CommandCommunityStopProcess), cm.CommunityStopProcessHandler)

	router.HandleFunc(ctl.GetRoute(ctl.CommandAddFile), n.FS.AddHandler)
	router.HandleFunc(ctl.GetRoute(ctl.CommandGetFile), n.FS.GetHandler)

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
