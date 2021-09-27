package compass

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/google/uuid"
)

const (
	osLinux  = "linux"
	osDarwin = "darwin"

	habitatPathEnv         = "HABITAT_PATH"
	defaultHabitatPathUnix = "~/.habitat"

	nodeIDRelativePath = "node_id"
)

func HabitatPath() string {
	switch runtime.GOOS {
	case osLinux:
		fallthrough
	case osDarwin:
		habitatPathEnv := os.Getenv(habitatPathEnv)
		if habitatPathEnv == "" {
			return defaultHabitatPathUnix
		}
		return habitatPathEnv
	default:
		panic(fmt.Sprintf("operating system %s not supported", runtime.GOOS))
	}
}

func ProcsPath() string {
	return filepath.Join(HabitatPath(), "procs")
}

func BinPath() string {
	archOS := fmt.Sprintf("%s-%s", runtime.GOARCH, runtime.GOOS)
	return filepath.Join(ProcsPath(), "bin", archOS)
}

func DataPath() string {
	return filepath.Join(ProcsPath(), "data")
}

func CommunitiesPath() string {
	return filepath.Join(HabitatPath(), "communities")
}

// NodeID will return the value in the node_id file, creating it if it doesn't exist
func NodeID() string {
	nodeIDPath := filepath.Join(HabitatPath(), nodeIDRelativePath)
	// Check if node id file exists
	_, err := os.Stat(nodeIDPath)
	if errors.Is(err, os.ErrNotExist) {
		// create new node id file
		nodeIDFile, err := os.OpenFile(nodeIDPath, os.O_CREATE|os.O_WRONLY, 0600)
		if err != nil {
			panic(fmt.Sprintf("error writing %s: %s", nodeIDPath, err))
		}
		defer nodeIDFile.Close()

		nodeID := uuid.NewString()

		_, err = nodeIDFile.Write([]byte(nodeID))
		if err != nil {
			panic(fmt.Sprintf("error writing %s: %s", nodeIDPath, err))
		}

		return nodeID
	} else if err != nil {
		panic(fmt.Sprintf("error reading %s: %s", nodeIDPath, err))
	}

	// If node file exists, just read it from the file
	return readSingleValueConfigFile(nodeIDPath)
}

// TODO this should probably figure out public IP stuff too
func Hostname() string {
	switch runtime.GOOS {
	case osLinux:
		fallthrough
	case osDarwin:
		return strings.TrimSpace(singleValueCommand("hostname"))
	default:
		panic(fmt.Sprintf("operating system %s not supported", runtime.GOOS))
	}
}