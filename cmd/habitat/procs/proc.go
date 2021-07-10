package procs

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

type Proc struct {
	Name   string
	Path   string
	cancel context.CancelFunc
}

func NewProc(name, path string) *Proc {
	return &Proc{
		Name: name,
		Path: path,
	}
}

func (p *Proc) Start() error {
	ctx, cancel := context.WithCancel(context.Background())
	p.cancel = cancel

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
	cmd := exec.CommandContext(ctx, startPath)
	buf, err := cmd.Output()
	if err != nil {
		return err
	}
	fmt.Println(string(buf))

	return nil
}

func (p *Proc) Stop() error {
	p.cancel()
	return nil
}
