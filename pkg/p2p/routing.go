package p2p

import (
	"context"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/routing"
)

type HabitatPeerRouting struct {
	// The peer routing table maintains a map of all Habitat nodes that
	// we know the address of
	peerRoutingTable map[peer.ID]*peer.AddrInfo
}

func (r *HabitatPeerRouting) FindPeer(ctx context.Context, peerID peer.ID) (peer.AddrInfo, error) {
	addrInfo, ok := r.peerRoutingTable[peerID]
	if !ok {
		return peer.AddrInfo{}, routing.ErrNotFound
	}
	return *addrInfo, nil
}
