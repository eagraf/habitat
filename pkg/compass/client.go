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

func DefaultHabitatAPIAddrWebsocket() string {
	return CustomHabitatAPIAddrWebsocket("localhost", apiPort)
}

func CustomHabitatAPIAddrWebsocket(host, port string) string {
	return fmt.Sprintf("ws://%s:%s", host, port)
}

func LibP2PHabitatAPIAddr(proxyMultiaddr string) (peer.ID, ma.Multiaddr, error) {
	return DecomposeNodeMultiaddr(proxyMultiaddr)
}
