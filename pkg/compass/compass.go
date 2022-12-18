package compass

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"runtime"

	"github.com/google/uuid"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
	ma "github.com/multiformats/go-multiaddr"
)

const (
	osLinux  = "linux"
	osDarwin = "darwin"

	habitatPathEnv         = "HABITAT_PATH"
	habitatIdentityPathEnv = "HABITAT_IDENTITY_PATH"

	nodeIDRelativePath = "node_id"

	apiPort     = "2040"
	p2pPort     = "6000"
	sourcesPort = "8765"
)

func HabitatPath() string {
	switch runtime.GOOS {
	case osLinux:
		fallthrough
	case osDarwin:
		habitatPathEnv := os.Getenv(habitatPathEnv)
		if habitatPathEnv == "" {
			userHome, err := os.UserHomeDir()
			if err != nil {
				panic("can't get user home directory")
			}

			return filepath.Join(userHome, ".habitat")
		}
		return habitatPathEnv
	default:
		panic(fmt.Sprintf("operating system %s not supported", runtime.GOOS))
	}
}

func HabitatIdentityPath() string {
	identityPathEnv := os.Getenv(habitatIdentityPathEnv)

	if identityPathEnv == "" {
		return filepath.Join(HabitatPath(), "identity")
	}
	return identityPathEnv
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

func LocalSourcesPath() string {
	return filepath.Join(DataPath(), "sources")
}

func LocalSchemaPath() string {
	return filepath.Join(DataPath(), "schema")
}

func SourcesServerPort() string {
	return sourcesPort
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

func PublicIP() (net.IP, error) {
	// TODO if we are in dockerland, fake this
	// TODO use SSL
	url := "http://api64.ipify.org?format=text"
	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	ip := net.ParseIP(string(body))
	if ip == nil {
		return nil, errors.New("invalid IP address")
	}
	return ip, nil
}

func IPMultiaddr() (multiaddr.Multiaddr, error) {
	var ip net.IP
	if InDockerContainer() {
		localIP, err := LocalIPv4()
		if err != nil {
			return nil, err
		}
		ip = localIP
	} else {
		pubIP, err := PublicIP()
		if err != nil {
			return nil, err
		}
		ip = pubIP
	}

	ipVersion := "ip4"
	if ip.To4() == nil {
		ipVersion = "ip6"
	}
	addr, err := multiaddr.NewMultiaddr(fmt.Sprintf("/%s/%s", ipVersion, ip.String()))
	if err != nil {
		return nil, err
	}
	return addr, nil
}

func IPFSSwarmAddr() (multiaddr.Multiaddr, error) {
	ipMA, err := IPMultiaddr()
	if err != nil {
		return nil, err
	}
	tail, err := multiaddr.NewMultiaddr(fmt.Sprintf("/tcp/4001/p2p/%s", PeerID().String()))
	if err != nil {
		return nil, err
	}
	ma := ipMA.Encapsulate(tail)
	return ma, nil
}

func PublicLibP2PMultiaddr() (multiaddr.Multiaddr, error) {
	ipMA, err := IPMultiaddr()
	if err != nil {
		return nil, err
	}
	tail, err := multiaddr.NewMultiaddr(fmt.Sprintf("/tcp/%s/p2p/%s", p2pPort, PeerID().String()))
	if err != nil {
		return nil, err
	}
	ma := ipMA.Encapsulate(tail)
	return ma, nil
}

// TODO this should really be something like InTestingDockerContainer
// You'd need a way to know that however
func InDockerContainer() bool {
	_, err := os.Stat("/.dockerenv")
	return err == nil
}

func DecomposeNodeMultiaddr(multiaddr string) (peer.ID, ma.Multiaddr, error) {
	remoteMA, err := ma.NewMultiaddr(multiaddr)
	if err != nil {
		return "", nil, err
	}

	b58PeerID, err := remoteMA.ValueForProtocol(ma.P_P2P)
	if err != nil {
		return "", nil, fmt.Errorf("couldn't retrieve p2p value in multiaddr: %s", multiaddr)
	}

	addrStr := ""
	ip, err := remoteMA.ValueForProtocol(ma.P_IP4)
	if err != nil {
		ip, err = remoteMA.ValueForProtocol(ma.P_IP6)
		if err != nil {
			return "", nil, fmt.Errorf("couldn't retrieve ip4 or ip6 value in multiaddr: %s", multiaddr)
		} else {
			addrStr += "/ip6/" + ip
		}
	} else {
		addrStr += "/ip4/" + ip
	}

	port, err := remoteMA.ValueForProtocol(ma.P_TCP)
	if err != nil {
		return "", nil, fmt.Errorf("couldn't retrieve tcp value in multiaddr: %s", multiaddr)
	}
	addrStr += "/tcp/" + port

	addr, err := ma.NewMultiaddr(addrStr)
	if err != nil {
		return "", nil, err
	}

	// decode base58 encoded peer id for setting addresses
	peerID, err := peer.Decode(b58PeerID)
	if err != nil {
		return "", nil, err
	}

	return peerID, addr, nil
}
