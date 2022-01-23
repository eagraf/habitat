package procs

import (
	"fmt"
	"path/filepath"
	"strings"
	"sync"

	"github.com/eagraf/habitat/cmd/habitat/proxy"
	"github.com/eagraf/habitat/pkg/compass"
	"github.com/eagraf/habitat/structs/configuration"
	"github.com/eagraf/habitat/structs/ctl"
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

func (m *Manager) StartProcess(req *ctl.Request) (error, string) {
	m.lock.Lock()
	defer m.lock.Unlock()

	name := req.Args[0]
	subName := name
	if len(req.Args) > 1 {
		subName = strings.Join(req.Args, "-")
	}

	fmt.Printf("starting process %s \n", subName)
	appConfig, ok := m.AppConfigs.Apps[name]
	if !ok {
		return fmt.Errorf("no app with name %s in app configurations", name), ""
	}

	if _, ok := m.Procs[subName]; ok {
		return fmt.Errorf("process with name %s already exists", subName), ""
	}

	cmdPath := filepath.Join(compass.BinPath(), appConfig.Bin)
	dataPath := filepath.Join(compass.DataPath(), name)
	proc := NewProc(name, cmdPath, dataPath, m.errChan, req.Env, req.Flags, req.Args[1:])
	err := proc.Start()
	if err != nil {
		return err, ""
	}

	// Update reverse proxy ruleset
	for _, ruleConfig := range appConfig.ProxyRules {
		rule, err := proxy.GetRuleFromConfig(ruleConfig, m.ProcDir)
		if err != nil {
			return err, ""
		}
		m.ProxyRules.Add(ruleConfig.Hash(), rule)
	}

	m.Procs[subName] = proc
	return nil, subName
}

func (m *Manager) StopProcess(subName string) error {
	m.lock.Lock()
	defer m.lock.Unlock()

	fmt.Printf("stopping process %s \n", subName)
	if _, ok := m.Procs[subName]; !ok {
		return fmt.Errorf("process with name %s does not exist", subName)
	}
	err := m.Procs[subName].Stop()
	if err != nil {
		return err
	}
	delete(m.Procs, subName)

	name := strings.Split(subName, "-")[0]
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
