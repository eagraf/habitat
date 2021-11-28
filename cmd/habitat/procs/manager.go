package procs

import (
	"fmt"
	"path/filepath"
	"sync"

	"github.com/eagraf/habitat/cmd/habitat/proxy"
	"github.com/eagraf/habitat/pkg/compass"
	"github.com/eagraf/habitat/structs/configuration"
	"github.com/rs/zerolog/log"
)

type Manager struct {
	ProcDir    string
	Procs      map[string]*Proc
	ProxyRules proxy.RuleSet
	AppConfigs *configuration.AppConfiguration

	errChan chan ProcError
	lock    *sync.Mutex

	archOS string
}

func NewManager(procDir string, rules proxy.RuleSet, appConfigs *configuration.AppConfiguration) *Manager {
	return &Manager{
		ProcDir:    compass.ProcsPath(),
		Procs:      make(map[string]*Proc),
		ProxyRules: rules,
		AppConfigs: appConfigs,

		errChan: make(chan ProcError),
		lock:    &sync.Mutex{},
	}
}

func (m *Manager) StartProcess(name string) error {
	m.lock.Lock()
	defer m.lock.Unlock()

	appConfig, ok := m.AppConfigs.Apps[name]
	if !ok {
		return fmt.Errorf("no app with name %s in app configurations", name)
	}

	if _, ok := m.Procs[name]; ok {
		return fmt.Errorf("process with name %s already exists", name)
	}

	cmdPath := filepath.Join(compass.BinPath(), appConfig.Bin)
	dataPath := filepath.Join(compass.DataPath(), name)
	proc := NewProc(name, cmdPath, dataPath, m.errChan)
	err := proc.Start()
	if err != nil {
		return err
	}

	// Update reverse proxy ruleset
	for _, ruleConfig := range appConfig.ProxyRules {
		rule, err := proxy.GetRuleFromConfig(ruleConfig, m.ProcDir)
		if err != nil {
			return err
		}
		m.ProxyRules.Add(ruleConfig.Hash(), rule)
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

	appConfig, ok := m.AppConfigs.Apps[name]
	if !ok {
		return fmt.Errorf("no app with name %s in app configurations", name)
	}

	// Remove from proxy ruleset
	for _, rule := range appConfig.ProxyRules {
		fmt.Printf("removing rule %s", rule.Hash())
		m.ProxyRules.Remove(rule.Hash())
	}

	return nil
}

// Return a readonly list of processes
func (m *Manager) ListProcesses() ([]*Proc, error) {
	m.lock.Lock()
	defer m.lock.Unlock()

	res := make([]*Proc, len(m.Procs))
	i := 0
	for _, p := range m.Procs {
		res[i] = &Proc{
			Name: p.Name,
		}
		i++
	}

	return res, nil
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
