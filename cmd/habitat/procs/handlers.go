package procs

import (
	"net/http"

	"github.com/eagraf/habitat/cmd/habitat/api"
	"github.com/eagraf/habitat/structs/ctl"
)

func (m *Manager) StartProcessHandler(w http.ResponseWriter, r *http.Request) {
	var startReq ctl.StartRequest
	err := api.BindPostRequest(r, &startReq)
	if err != nil {
		api.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	procID := RandomProcessID()
	if err != nil {
		api.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	processInstanceID, err := m.StartProcessInstance(startReq.CommunityID, procID, startReq.App, startReq.Args, startReq.Env)
	if err != nil {
		api.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	startRes := &ctl.StartResponse{
		ProcessInstanceID: processInstanceID,
	}

	api.WriteResponse(w, startRes)
}

func (m *Manager) StopProcessHandler(w http.ResponseWriter, r *http.Request) {
	var stopReq ctl.StopRequest
	err := api.BindPostRequest(r, &stopReq)
	if err != nil {
		api.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	err = m.StopProcessInstance(stopReq.ProcID)
	if err != nil {
		api.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	stopRes := &ctl.StopResponse{}

	api.WriteResponse(w, stopRes)
}

func (m *Manager) ListProcessesHandler(w http.ResponseWriter, r *http.Request) {
	procs, err := m.listProcessInstances()
	if err != nil {
		api.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	psRes := &ctl.PSResponse{
		ProcIDs: make([]string, 0),
	}
	for _, p := range procs {
		psRes.ProcIDs = append(psRes.ProcIDs, p.Name)
	}

	api.WriteResponse(w, psRes)
}
