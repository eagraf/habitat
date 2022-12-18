package procs

import (
	"fmt"
	"hash/fnv"
	"path/filepath"
	"sync"

	"github.com/eagraf/habitat/cmd/habitat/proxy"
	"github.com/eagraf/habitat/pkg/compass"
	"github.com/eagraf/habitat/structs/configuration"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

type Manager struct {
	ProcDir    string
	Procs      map[string]*Proc
	ProxyRules proxy.RuleSet

	errChan chan ProcError
	lock    *sync.Mutex
}

func NewManager(procDir string, rules proxy.RuleSet) *Manager {
	return &Manager{
		ProcDir:    compass.ProcsPath(),
		Procs:      make(map[string]*Proc),
		ProxyRules: rules,

		errChan: make(chan ProcError),
		lock:    &sync.Mutex{},
	}
}

func (m *Manager) StartProcessInstance(communityID, processID, app string, args, env, flags []string) (string, error) {
	m.lock.Lock()
	defer m.lock.Unlock()

	processInstanceID, err := GetProcessInstanceID(communityID, compass.NodeID(), processID)
	if err != nil {
		return "", err
	}

	appConfig, appPath, err := configuration.GetAppConfig(app)
	if err != nil {
		return "", err
	}

	binPath := filepath.Join(appPath, "bin", appConfig.Bin)

	if _, ok := m.Procs[processInstanceID]; ok {
		return "", fmt.Errorf("process with name %s already exists", processInstanceID)
	}

	proc := NewProc(processInstanceID, binPath, m.errChan, env, flags, args, appConfig)
	err = proc.Start()
	if err != nil {
		return "", err
	}

	// Update reverse proxy ruleset
	for _, ruleConfig := range appConfig.ProxyRules {
		rule, err := proxy.GetRuleFromConfig(ruleConfig, appPath)
		if err != nil {
			return "", err
		}
		m.ProxyRules.Add(ruleConfig.Hash(), rule)
	}

	m.Procs[processInstanceID] = proc
	return processInstanceID, nil
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

func (m *Manager) StopProcessInstance(procID string) error {
	m.lock.Lock()
	defer m.lock.Unlock()

	if _, ok := m.Procs[procID]; !ok {
		return fmt.Errorf("process with name %s does not exist", procID)
	}

	proc := m.Procs[procID]

	err := proc.Stop()
	if err != nil {
		return err
	}
	delete(m.Procs, procID)

	// Remove from proxy ruleset
	for _, rule := range proc.config.ProxyRules {
		m.ProxyRules.Remove(rule.Hash())
	}

	return nil
}

// Return a readonly list of processes
func (m *Manager) listProcessInstances() ([]*Proc, error) {
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
		m.StopProcessInstance(procErr.proc.Name)
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

func RandomProcessID() string {
	return uuid.New().String()
}

func GetProcessInstanceID(communityID, nodeID, processID string) (string, error) {
	concatenated := fmt.Sprintf("%s-%s-%s", communityID, nodeID, processID)
	hasher := fnv.New128a()
	_, err := hasher.Write([]byte(concatenated))
	if err != nil {
		return "", err
	}

	res := hasher.Sum([]byte{})
	return string(res), nil
}
