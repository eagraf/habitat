package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/eagraf/habitat/cmd/habitat/community"
	"github.com/eagraf/habitat/cmd/habitat/procs"
	"github.com/eagraf/habitat/structs/ctl"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
)

const (
	HabitatCTLHost = "0.0.0.0"
	HabitatCTLPort = "2040"
)

func getRouter(pm *procs.Manager, cm *community.Manager) *mux.Router {
	router := mux.NewRouter()
	router.HandleFunc(ctl.GetRoute(ctl.CommandListProcesses), pm.ListProcessesHandler)
	router.HandleFunc(ctl.GetRoute(ctl.CommandStart), pm.StartProcessHandler)
	router.HandleFunc(ctl.GetRoute(ctl.CommandStop), pm.StopProcessHandler)

	router.HandleFunc(ctl.GetRoute(ctl.CommandCommunityCreate), cm.CommunityCreateHandler)
	router.HandleFunc(ctl.GetRoute(ctl.CommandCommunityJoin), cm.CommunityJoinHandler)
	router.HandleFunc(ctl.GetRoute(ctl.CommandCommunityAddMember), cm.CommunityAddMemberHandler)
	router.HandleFunc(ctl.GetRoute(ctl.CommandCommunityPropose), cm.CommunityProposeHandler)
	router.HandleFunc(ctl.GetRoute(ctl.CommandCommunityState), cm.CommunityStateHandler)
	router.HandleFunc(ctl.GetRoute(ctl.CommandCommunityList), cm.CommunityListHandler)

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
