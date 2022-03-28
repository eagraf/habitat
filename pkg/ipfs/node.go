package ipfs

import (
	"bufio"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	mrand "math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	config "github.com/ipfs/go-ipfs-config"
	"github.com/rs/zerolog/log"
)

// from https://github.com/Kubuxu/go-ipfs-swarm-key-gen/blob/master/ipfs-swarm-key-gen/main.go
func KeyGen() string {
	key := make([]byte, 32)
	_, err := rand.Read(key)
	if err != nil {
		fmt.Println("While trying to read random source:", err)
	}

	return hex.EncodeToString(key)
}

// internal structs for IPFSNode (may remove)
type IPFSConfig struct {
	CommunitiesPath string
	StartCmd        string
}

func (c *IPFSConfig) NewCommunityIPFSNode(name string, path string) (error, string, string, []string) {
	cmdCreate := &exec.Cmd{
		Path:   c.StartCmd,
		Args:   []string{c.StartCmd, filepath.Join(path, "ipfs")},
		Stderr: os.Stderr,
	}

	s := make([]string, 0)
	if err := cmdCreate.Run(); err != nil {
		return err, "", "", s
	}

	bytes, _ := ioutil.ReadFile(filepath.Join(path, "ipfs", "config"))
	var data config.Config
	err := json.Unmarshal(bytes, &data)

	if err != nil {
		return err, "", "", s
	}

	// json struct of config : here we can modify it and write back
	// ignore the peers for now (connect after bootstrapping?)
	/* choose a random port to be used by ipfs config (so that having)
	   multiple ipfs instances doesn't cause a conflict */
	addr := data.Addresses.API
	addrstring := addr[0]
	parts := strings.Split(addrstring, "/")
	parts[len(parts)-1] = fmt.Sprint(mrand.Intn(65536-9999) + 9999)
	data.Addresses.API = []string{strings.Join(parts, "/")}
	bytes, err = json.Marshal(data)
	log.Info().Msg("data " + string(bytes))
	ioutil.WriteFile(filepath.Join(path, "ipfs", "/config"), bytes, 0755)

	// json struct of config : here we can modify it and write back

	key := KeyGen()
	keyBytes := []byte("/key/swarm/psk/1.0.0/\n/base16/\n" + key + "\n")
	err = ioutil.WriteFile(filepath.Join(path, "ipfs", "swarm.key"), keyBytes, 0755)

	return nil, key, data.Identity.PeerID, data.Addresses.Swarm
}

func (c *IPFSConfig) JoinCommunityIPFSNode(name string, path string, key string, btsppeers []string) (string, error) {
	commPath := filepath.Join(c.CommunitiesPath, path)
	cmdJoin := &exec.Cmd{
		Path:   c.StartCmd,
		Args:   []string{c.StartCmd, commPath},
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}

	if err := cmdJoin.Run(); err != nil {
		return "", err
	}

	bytes, _ := ioutil.ReadFile(filepath.Join(commPath, "ipfs", "config"))
	var data config.Config
	err := json.Unmarshal(bytes, &data)

	if err != nil {
		return "", err
	}

	// json struct of config : here we can modify it and write back
	// ignore the peers for now (connect after bootstrapping?)
	data.Bootstrap = btsppeers
	addr := data.Addresses.API
	addrstring := addr[0]
	parts := strings.Split(addrstring, "/")
	parts[len(parts)-1] = fmt.Sprint(mrand.Intn(65536-9999) + 9999)
	data.Addresses.API = []string{strings.Join(parts, "/")}
	bytes, err = json.Marshal(data)
	log.Info().Msg("data " + string(bytes))
	ioutil.WriteFile(filepath.Join(commPath, "ipfs", "config"), bytes, 0755)

	keyBytes := []byte("/key/swarm/psk/1.0.0/\n/base16/\n" + key + "\n")
	err = ioutil.WriteFile(filepath.Join(commPath, "ipfs", "/swarm.key"), keyBytes, 0755)

	return data.Identity.PeerID, err
}

type IPFSConnectedConfig struct {
	Id              string   `json:"ID"`
	PublicKey       string   `json:"PublicKey"`
	Addresses       []string `json:"Addresses"`
	AgentVersion    string   `json:"AgentVersion"`
	ProtocolVersion string   `json:"ProtocolVersion"`
	Protocols       []string `json:"Protocols"`
	SwarmKey        string   `json:"SwarmKey"`
}

type ConnectedConfig struct {
	PeerId          string   `json:"PeerId"`
	CommId          string   `json:"CommId"`
	Name            string   `json:"Name"`
	PublicKey       string   `json:"PublicKey"`
	Addresses       []string `json:"Addresses"`
	AgentVersion    string   `json:"AgentVersion"`
	ProtocolVersion string   `json:"ProtocolVersion"`
	Protocols       []string `json:"Protocols"`
	SwarmKey        string   `json:"SwarmKey"`
}

func (c *IPFSConfig) ConnectCommunityIPFSNode(name string, id string) (*ConnectedConfig, error) {
	// TODO: either delete connect script or make this use it
	key := ""
	keyPath := filepath.Join(c.CommunitiesPath, id, "ipfs", "swarm.key")
	keyFile, err := os.Open(keyPath)
	if err != nil {
		log.Error().Msg(fmt.Sprintf("unable to open swarm key file for community id: %s, path: %s", id, keyPath))
		return nil, err
	} else {
		fileScanner := bufio.NewScanner(keyFile)
		fileScanner.Split(bufio.ScanLines)
		lineNum := 0
		for fileScanner.Scan() {
			if lineNum == 2 { // third line
				key = fileScanner.Text()
				break
			}
			lineNum++
		}
	}

	pathEnv := fmt.Sprintf("IPFS_PATH=%s", filepath.Join(c.CommunitiesPath, id, "ipfs"))
	cmdConnect := exec.Command("ipfs", "daemon")
	stdout, _ := cmdConnect.StdoutPipe()

	cmdConnect.Stderr = os.Stderr
	cmdConnect.Env = []string{pathEnv}

	done := make(chan struct{})

	scanner := bufio.NewScanner(stdout)

	// Use the scanner to scan the output line by line and log it
	// It's running in a goroutine so that it doesn't block
	go func() {
		for scanner.Scan() {
			line := scanner.Text()
			if strings.Contains(line, "Daemon is ready") {
				done <- struct{}{}
				return
			}
		}
	}()

	// Start the command and check for errors
	err = cmdConnect.Start()
	if err != nil {
		log.Err(err).Msg("error running ipfs daemon")
		return nil, err
	}

	// Wait for 'Daemon is ready'
	<-done

	cmdId := exec.Command("ipfs", "id")
	cmdId.Env = []string{pathEnv}
	out, err := cmdId.Output()

	if err != nil {
		log.Err(err)
		return nil, err
	}

	var data IPFSConnectedConfig
	err = json.Unmarshal(out, &data)
	if err != nil {
		log.Fatal().Err(err)
		return nil, err
	}
	data.SwarmKey = key

	return &ConnectedConfig{
		PeerId:          data.Id,
		CommId:          id,
		Name:            name,
		PublicKey:       data.PublicKey,
		Addresses:       data.Addresses,
		AgentVersion:    data.AgentVersion,
		ProtocolVersion: data.ProtocolVersion,
		Protocols:       data.Protocols,
		SwarmKey:        data.SwarmKey,
	}, nil
}

/*
	The following functions use the go core API implementation in golang. which is not very well supported right now.
	They are left commented because including them causes a lot of build problems due to all the imports from
	people's personal github repos in the ipfs libraries.
*/

/*
// general function to create a new ipfs node
func NewIPFSNode_Lib(path string) (*IPFSConfig, error) {
	node := &IPFSConfig{
		path,
	}
	return node, nil
}

func createNode(ctx context.Context, repoPath string, private bool) (icore.CoreAPI, error) {
	// Open the repo
	repo, err := fsrepo.Open(repoPath)
	if err != nil {
		return nil, err
	}

	// if it's a private network node, remove bootstrap addresses from config
	if private {
		cfg, err := repo.Config()
		if err != nil {
			return nil, err
		}

		cfg.SetBootstrapPeers(nil)
		err = repo.SetConfig(cfg)
		if err != nil {
			return nil, err
		}
	}

	// Construct the node

	nodeOptions := &core.BuildCfg{
		Online:  true,
		Routing: libp2p.DHTOption, // This option sets the node to be a full DHT node (both fetching and storing DHT Records)
		// Routing: libp2p.DHTClientOption, // This option sets the node to be a client DHT node (only fetching records)
		Repo: repo,
	}

	node, err := core.NewNode(ctx, nodeOptions)
	if err != nil {
		return nil, err
	}

	// Attach the Core API to the constructed node
	return coreapi.NewCoreAPI(node)
}

func connectToPeers(ctx context.Context, ipfs icore.CoreAPI, peers []string) error {
	var wg sync.WaitGroup
	peerInfos := make(map[peer.ID]*peer.AddrInfo, len(peers))
	for _, addrStr := range peers {
		addr, err := ma.NewMultiaddr(addrStr)
		if err != nil {
			return err
		}
		pii, err := peer.AddrInfoFromP2pAddr(addr)
		if err != nil {
			return err
		}
		pi, ok := peerInfos[pii.ID]
		if !ok {
			pi = &peer.AddrInfo{ID: pii.ID}
			peerInfos[pi.ID] = pi
		}
		pi.Addrs = append(pi.Addrs, pii.Addrs...)
	}

	wg.Add(len(peerInfos))
	for _, peerInfo := range peerInfos {
		go func(peerInfo *peer.AddrInfo) {
			defer wg.Done()
			err := ipfs.Swarm().Connect(ctx, *peerInfo)
			if err != nil {
				log.Printf("failed to connect to %s: %s", peerInfo.ID, err)
			}
		}(peerInfo)
	}
	wg.Wait()
	return nil
}
*/
