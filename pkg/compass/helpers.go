package compass

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
)

// readSingleValueConfigFile is useful for reading single value config files (e.g. /etc/hostname)
// This function will panic under any error, as it assumes that it is opening critical config
// that the program cannot run without
func readSingleValueConfigFile(path string) string {
	file, err := os.Open(path)
	if err != nil {
		panic(fmt.Sprintf("unable to open %s: %s", path, err))
	}
	defer file.Close()

	buf, err := ioutil.ReadAll(file)
	if err != nil {
		panic(fmt.Sprintf("unable to read %s: %s", path, err))
	}

	return string(buf)
}

// singleValueCommand is good for running commands that return a single value config (e.g. hostname)
//nolint
func singleValueCommand(cmd string) string {
	command := exec.Command(cmd)
	output, err := command.CombinedOutput()
	if err != nil {
		panic(fmt.Sprintf("error running %s: %s", cmd, err))
	}
	return string(output)
}
