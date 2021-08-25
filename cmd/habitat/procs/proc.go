package procs

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"github.com/rs/zerolog/log"
)

type Proc struct {
	Name     string
	CmdPath  string
	DataPath string

	cmd     *exec.Cmd
	errChan chan ProcError
}

func NewProc(name, cmdPath, dataPath string, errChan chan ProcError) *Proc {
	return &Proc{
		Name:     name,
		CmdPath:  cmdPath,
		DataPath: dataPath,

		errChan: errChan,
	}
}

func (p *Proc) Start() error {
	cmd := &exec.Cmd{
		Path: p.CmdPath,
		Env: []string{
			fmt.Sprintf("DATA_DIR=%s", p.DataPath),
		},
	}

	// start this process with a groupd id equal to its pid. this allows for all of its subprocesses to be killed
	// at once by passing in the negative pid to syscall.Kill
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	p.cmd = cmd

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Start()
	if err != nil {
		return err
	}

	go func() {
		err := cmd.Wait()
		if err != nil {
			if _, ok := err.(*exec.ExitError); ok {
				procErr := ProcError{
					proc:    p,
					message: err.Error(),
				}
				p.errChan <- procErr
			} else {
				log.Error().Msgf("process %s encountered error: %s", p.Name, err)
			}
		}
	}()

	return nil
}

func (p *Proc) Stop() error {
	terminateProcess := func(pid int) {
		// force kill process afterwards
		// TODO make sure this works on all operating systems
		// passing in negative pid makes sure all child processes are killed as well
		err := syscall.Kill(-pid, syscall.SIGTERM)
		if err != nil {
			log.Error().Msg(err.Error())
		}
	}
	defer terminateProcess(p.cmd.Process.Pid)

	return nil
}
