package procs

import (
	"fmt"
	"path/filepath"
)

type Manager struct {
	ProcDir string
	Procs   map[string]*Proc
}

func NewManager(procDir string) *Manager {
	return &Manager{
		ProcDir: procDir,
		Procs:   make(map[string]*Proc),
	}
}

func (m *Manager) StartProcess(name string) error {
	if _, ok := m.Procs[name]; ok {
		return fmt.Errorf("process with name %s already exists", name)
	}
	procDir := filepath.Join(m.ProcDir, name)

	proc := NewProc(name, procDir)
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
