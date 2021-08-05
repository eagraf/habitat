package procs

import (
	"fmt"
	"path/filepath"

	"github.com/rs/zerolog/log"
)

type Manager struct {
	ProcDir string
	Procs   map[string]*Proc

	errChan chan ProcError
}

func NewManager(procDir string) *Manager {
	return &Manager{
		ProcDir: procDir,
		Procs:   make(map[string]*Proc),

		errChan: make(chan ProcError),
	}
}

func (m *Manager) StartProcess(name string) error {
	if _, ok := m.Procs[name]; ok {
		return fmt.Errorf("process with name %s already exists", name)
	}
	procDir := filepath.Join(m.ProcDir, name)

	proc := NewProc(name, procDir, m.errChan)
	err := proc.Start()
	if err != nil {
		return err
	}

	m.Procs[name] = proc
	return nil
}

func (m *Manager) StopProcess(name string) error {
	if _, ok := m.Procs[name]; !ok {
		return fmt.Errorf("process with name %s does not exist", name)
	}
	err := m.Procs[name].Stop()
	if err != nil {
		return err
	}
	delete(m.Procs, name)
	return nil
}

func (m *Manager) ListenForErrors() {
	for {
		procErr := <-m.errChan
		log.Error().Msg(procErr.String())
		// try stop command in case it has any clean up, don't worry too much about errors
		err := m.Procs[procErr.proc.Name].Stop()
		if err != nil {
			log.Error().Msg(err.Error())
		}
		// remove process, since we know it exited
		delete(m.Procs, procErr.proc.Name)
	}
}

type ProcError struct {
	message string
	proc    *Proc
}

func (pe ProcError) Error() string {
	return fmt.Sprintf("error in process %s: %s", pe.proc.Name, pe.message)
}

func (pe ProcError) String() string {
	return pe.Error()
}
