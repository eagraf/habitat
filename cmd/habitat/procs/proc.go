package procs

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
)

type Proc struct {
	Name string
	Path string

	cmd *exec.Cmd
}

func NewProc(name, path string) *Proc {
	return &Proc{
		Name: name,
		Path: path,
	}
}

func (p *Proc) Start() error {
	// make sure that proc dir exists
	fileInfo, err := os.Stat(p.Path)
	if err != nil {
		return fmt.Errorf("couldn't find process dir %s", p.Path)
	}
	if !fileInfo.IsDir() {
		return fmt.Errorf("%s is not a dir", p.Path)
	}

	// start process
	startPath := filepath.Join(p.Path, "start.sh")
	cmd := exec.Command(startPath)

	// start this process with a groupd id equal to its pid. this allows for all of its subprocesses to be killed
	// at once by passing in the negative pid to syscall.Kill
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	p.cmd = cmd

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	scanner := bufio.NewScanner(stdout)
	errScanner := bufio.NewScanner(stderr)

	err = cmd.Start()
	if err != nil {
		return err
	}

	go func() {
		for scanner.Scan() {
			line := scanner.Text()
			fmt.Println(line)
		}
	}()

	go func() {
		for errScanner.Scan() {
			line := errScanner.Text()
			fmt.Println(line)
		}
	}()

	return nil
}

func (p *Proc) Stop() error {
	// if stop.sh script exists, execute it
	stopPath := filepath.Join(p.Path, "stop.sh")
	if _, err := os.Stat(stopPath); err == nil {
		// stop process
		cmd := exec.Command(stopPath)

		stdout, err := cmd.StdoutPipe()
		if err != nil {
			return err
		}

		stderr, err := cmd.StderrPipe()
		if err != nil {
			return err
		}

		scanner := bufio.NewScanner(stdout)
		errScanner := bufio.NewScanner(stderr)

		err = cmd.Start()
		if err != nil {
			return err
		}

		go func() {
			for scanner.Scan() {
				line := scanner.Text()
				fmt.Println(line)
			}
		}()

		go func() {
			for errScanner.Scan() {
				line := errScanner.Text()
				fmt.Println(line)
			}
		}()
	}

	// force kill process afterwards
	// TODO make sure this works on all operating systems
	// passing in negative pid makes sure all child processes are killed as well
	err := syscall.Kill(-p.cmd.Process.Pid, syscall.SIGKILL)
	if err != nil {
		return err
	}
	return nil
}
