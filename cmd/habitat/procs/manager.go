package procs

import "fmt"

type Proc struct {
	Name string
}

type Manager struct {
	Procs map[string]*Proc
}

func NewManager() *Manager {
	return &Manager{
		Procs: make(map[string]*Proc),
	}
}

func (m *Manager) StartProcess(name string) error {
	if _, ok := m.Procs[name]; ok {
		return fmt.Errorf("process with name %s already exists", name)
	}
	m.Procs[name] = &Proc{
		Name: name,
	}
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
