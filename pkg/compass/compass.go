package compass

import (
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"runtime"

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
	hostname, err := os.Hostname()
	if err != nil {
		panic(fmt.Sprintf("can't get hostname: %s", err))
	}
	return hostname
}

func LocalIPv4() (net.IP, error) {
	// Dial a dummy connection to get the default local IP address
	// This solution is better than using net.Interfaces() because its possible the device
	// is using multiple network interfaces with different IP addresses, which would make it
	// difficult to establish which address it actually uses to communicate with the internet.
	// Establishing the dummy connection is a good workaround for extracting the default IP address used.
	conn, err := net.Dial("udp", "1.2.3.4:1")
	if err != nil {
		return nil, fmt.Errorf("error getting local IP address: %s", err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP, nil
}

func InDockerContainer() bool {
	_, err := os.Stat("/.dockerenv")
	return err == nil
}
