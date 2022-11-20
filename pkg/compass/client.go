package compass

import (
	"fmt"

	"github.com/libp2p/go-libp2p/core/peer"
	ma "github.com/multiformats/go-multiaddr"
)

func DefaultHabitatAPIAddr() string {
	return CustomHabitatAPIAddr("localhost", apiPort)
}

func CustomHabitatAPIAddr(host string, port string) string {
	return fmt.Sprintf("http://%s:%s", host, port)
}

func LibP2PHabitatAPIAddr(proxyMultiaddr string) (peer.ID, ma.Multiaddr, error) {
	remoteMA, err := ma.NewMultiaddr(proxyMultiaddr)
	if err != nil {
		return "", nil, err
	}

	b58PeerID, err := remoteMA.ValueForProtocol(ma.P_P2P)
	if err != nil {
		return "", nil, err
	}

	addrStr := ""
	ip, err := remoteMA.ValueForProtocol(ma.P_IP4)
	if err != nil {
		ip, err = remoteMA.ValueForProtocol(ma.P_IP6)
		if err != nil {
			return "", nil, err
		} else {
			addrStr += "/ip6/" + ip
		}
	} else {
		addrStr += "/ip4/" + ip
	}

	port, err := remoteMA.ValueForProtocol(ma.P_TCP)
	if err != nil {
		return "", nil, err
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
