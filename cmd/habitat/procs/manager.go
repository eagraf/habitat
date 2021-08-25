package procs

import (
	"fmt"
	"path/filepath"
	"runtime"
	"sync"

	"github.com/rs/zerolog/log"
)

type Manager struct {
	ProcDir string
	Procs   map[string]*Proc

	errChan chan ProcError
	lock    *sync.Mutex

	archOS string
}

func NewManager(procDir string) *Manager {
	return &Manager{
		ProcDir: procDir,
		Procs:   make(map[string]*Proc),

		errChan: make(chan ProcError),
		lock:    &sync.Mutex{},

		archOS: runtime.GOARCH + "-" + runtime.GOOS,
	}
}

func (m *Manager) StartProcess(name string) error {
	m.lock.Lock()
	defer m.lock.Unlock()

	if _, ok := m.Procs[name]; ok {
		return fmt.Errorf("process with name %s already exists", name)
	}

	cmdPath := filepath.Join(m.ProcDir, "bin", m.archOS, name)
	dataPath := filepath.Join(m.ProcDir, "data", name)
	proc := NewProc(name, cmdPath, dataPath, m.errChan)
	err := proc.Start()
	if err != nil {
		return err
	}

	m.Procs[name] = proc
	return nil
}

func (m *Manager) StopProcess(name string) error {
	m.lock.Lock()
	defer m.lock.Unlock()

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
		m.StopProcess(procErr.proc.Name)
	}
}

func (m *Manager) StopAllProcesses() {
	m.lock.Lock()
	defer m.lock.Unlock()

	for _, proc := range m.Procs {
		proc.Stop()
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
