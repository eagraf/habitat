package procs

import "github.com/eagraf/habitat/structs/ctl"

func (m *Manager) StartProcessHandler(req *ctl.RequestWrapper) (*ctl.ResponseWrapper, error) {
	var startReq ctl.StartRequest
	err := req.Deserialize(&startReq)
	if err != nil {
		return nil, err
	}

	procID, err := m.startProcess(startReq.App, startReq.CommunityID, startReq.Args, startReq.Env, startReq.Flags)
	if err != nil {
		return nil, err
	}

	startRes := &ctl.StartResponse{
		ProcID: procID,
	}
	res, err := ctl.NewResponseWrapper(startRes, ctl.StatusOK, "")
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (m *Manager) StopProcessHandler(req *ctl.RequestWrapper) (*ctl.ResponseWrapper, error) {
	var stopReq ctl.StopRequest
	err := req.Deserialize(&stopReq)
	if err != nil {
		return nil, err
	}

	err = m.stopProcess(stopReq.ProcID)
	if err != nil {
		return nil, err
	}

	stopRes := &ctl.StopResponse{}
	res, err := ctl.NewResponseWrapper(stopRes, ctl.StatusOK, "")
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (m *Manager) ListProcessesHandler(req *ctl.RequestWrapper) (*ctl.ResponseWrapper, error) {
	var psReq ctl.PSRequest
	err := req.Deserialize(&psReq)
	if err != nil {
		return nil, err
	}

	procs, err := m.listProcesses()
	if err != nil {
		return nil, err
	}

	psRes := &ctl.PSResponse{
		ProcIDs: make([]string, 0),
	}
	for _, p := range procs {
		psRes.ProcIDs = append(psRes.ProcIDs, p.Name)
	}
	res, err := ctl.NewResponseWrapper(psRes, ctl.StatusOK, "")
	if err != nil {
		return nil, err
	}

	return res, nil
}
