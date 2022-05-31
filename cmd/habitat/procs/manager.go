package procs

import (
	"fmt"
	"path/filepath"
	"strings"
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

func (m *Manager) StartProcess(app, communityID string, args, env, flags []string) (string, error) {
	m.lock.Lock()
	defer m.lock.Unlock()

	procID := app
	if communityID != "" {
		procID = fmt.Sprintf("%s-%s", app, communityID)
	}

	appConfig, ok := m.AppConfigs.Apps[app]
	if !ok {
		return "", fmt.Errorf("no app with name %s in app configurations", app)
	}

	if _, ok := m.Procs[procID]; ok {
		return "", fmt.Errorf("process with name %s already exists", procID)
	}

	cmdPath := filepath.Join(compass.BinPath(), appConfig.Bin)
	dataPath := filepath.Join(compass.DataPath(), app)
	proc := NewProc(app, cmdPath, dataPath, m.errChan, env, flags, args)
	err := proc.Start()
	if err != nil {
		return "", err
	}

	// Update reverse proxy ruleset
	for _, ruleConfig := range appConfig.ProxyRules {
		rule, err := proxy.GetRuleFromConfig(ruleConfig, m.ProcDir)
		if err != nil {
			return "", err
		}
		m.ProxyRules.Add(ruleConfig.Hash(), rule)
	}

	m.Procs[procID] = proc
	return procID, nil
}

/*
// TODO: @arushibandi revisit this when adding ipfs to process manager
// see: https://github.com/eagraf/habitat/pull/14#discussion_r827603513
func (m *Manager) StartArbitraryProcess(name string, cmdPath string, dataPath string, env []string, flags []string, args []string) (error, string) {
	proc := &Proc{
		Name:     name,
		CmdPath:  cmdPath,
		DataPath: dataPath,
		Env:      env,
		Flags:    flags,
		Args:     args,

		errChan: m.errChan,
	}

	err := proc.Start()
	if err != nil {
		return err, ""
	}

	// TODO: @arushibandi ensure that all names are unique
	m.Procs[name] = proc
	return nil, name
}
*/

func (m *Manager) StopProcess(procID string) error {
	m.lock.Lock()
	defer m.lock.Unlock()

	if _, ok := m.Procs[procID]; !ok {
		return fmt.Errorf("process with name %s does not exist", procID)
	}
	err := m.Procs[procID].Stop()
	if err != nil {
		return err
	}
	delete(m.Procs, procID)

	name := strings.Split(procID, "-")[0]
	appConfig, ok := m.AppConfigs.Apps[name]
	if !ok {
		return fmt.Errorf("no app with name %s in app configurations", name)
	}

	// Remove from proxy ruleset
	for _, rule := range appConfig.ProxyRules {
		m.ProxyRules.Remove(rule.Hash())
	}

	return nil
}

// Return a readonly list of processes
func (m *Manager) listProcesses() ([]*Proc, error) {
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
		log.Error().Msg("proc err listener got: " + procErr.String())

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
