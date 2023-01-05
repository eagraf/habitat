package dht

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/eagraf/habitat/pkg/crdt"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"
	bhost "github.com/libp2p/go-libp2p/p2p/host/basic"
	swarmt "github.com/libp2p/go-libp2p/p2p/net/swarm/testing"
	ma "github.com/multiformats/go-multiaddr"
)

func setupDHT(ctx context.Context, t *testing.T) *DHT {

	host, err := bhost.NewHost(swarmt.GenSwarm(t, swarmt.OptDisableReuseport), new(bhost.HostOpts))
	if err != nil {
		t.Fatal("error setting up host")
		return nil
	}
	t.Cleanup(func() { host.Close() })

	dht, err := NewDht(host)
	//t.Cleanup(func() { dht.Close() })
	if err != nil {
		t.Log("err", err)
		t.Fatal("error setting up dht", err)
		return nil
	}

	return dht
}

func setupDHTS(t *testing.T, ctx context.Context, n int) []*DHT {
	addrs := make([]ma.Multiaddr, n)
	dhts := make([]*DHT, n)
	peers := make([]peer.ID, n)

	sanityAddrsMap := make(map[string]struct{})
	sanityPeersMap := make(map[string]struct{})

	for i := 0; i < n; i++ {
		dhts[i] = setupDHT(ctx, t)
		peers[i] = dhts[i].PeerID()
		addrs[i] = dhts[i].Host().Addrs()[0]

		if _, lol := sanityAddrsMap[addrs[i].String()]; lol {
			t.Fatal("While setting up DHTs address got duplicated.")
		} else {
			sanityAddrsMap[addrs[i].String()] = struct{}{}
		}
		if _, lol := sanityPeersMap[peers[i].String()]; lol {
			t.Fatal("While setting up DHTs peerid got duplicated.")
		} else {
			sanityPeersMap[peers[i].String()] = struct{}{}
		}
	}

	return dhts
}

func connectNoSync(t *testing.T, ctx context.Context, a, b *DHT) {
	t.Helper()

	idB := b.PeerID()
	addrB := b.Host().Peerstore().Addrs(idB)
	if len(addrB) == 0 {
		t.Fatal("peers setup incorrectly: no local address")
	}

	a.Host().Peerstore().AddAddrs(idB, addrB, peerstore.TempAddrTTL)
	pi := peer.AddrInfo{ID: idB}
	if err := a.Host().Connect(ctx, pi); err != nil {
		t.Fatal(err)
	}
}

func wait(t *testing.T, ctx context.Context, a, b *DHT) {
	t.Helper()

	// loop until connection notification has been received.
	// under high load, this may not happen as immediately as we would like.

	for a.RoutingTable().Find(b.PeerID()) == "" {
		select {
		case <-ctx.Done():
			t.Fatal(ctx.Err())
		case <-time.After(time.Millisecond * 5):
		}
	}
}

func connect(t *testing.T, ctx context.Context, a, b *DHT) {
	t.Helper()
	connectNoSync(t, ctx, a, b)
	wait(t, ctx, a, b)
	wait(t, ctx, b, a)
}

func TestDht(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Second)
	defer cancel()

	dhts := setupDHTS(t, ctx, 3)

	for _, dht := range dhts[1:] {
		t.Log(dht.PeerID())
		connect(t, ctx, dhts[0], dht)
	}

	start, a, b, merged := crdt.TestStates()

	err := dhts[0].PutValue(ctx, "/crdt/test", start)
	if err != nil {
		t.Fatal("error putting start value in dht 0", err)
	}
	t.Log("successfully put start in dht 0")

	value, err := dhts[0].GetValue(ctx, "/crdt/test")

	if err != nil {
		t.Fatal("failed to get value from dht 0", err)
	}
	if !bytes.Equal(value, start) {
		t.Fatal("initial put of start failed")
	}
	t.Log("successfully got back start from dht 0")

	err = dhts[1].PutValue(ctx, "/crdt/test", a)
	if err != nil {
		t.Fatal("error putting value in dht 1", err)
	}
	t.Log("successfully put value in dht 1")

	value, err = dhts[0].GetValue(ctx, "/crdt/test")

	if err != nil {
		t.Fatal("failed to get value from dht 0", err)
	}

	if !bytes.Equal(a, value) {
		t.Fatal("bytes received from dht 0 not equal to a")
	}

	t.Log("got back a from dht 0")

	err = dhts[0].PutValue(ctx, "/crdt/test", b)
	if err != nil {
		t.Fatal("error putting b in dht 0", err)
	}

	value, err = dhts[2].GetValue(ctx, "/crdt/test")

	t.Log(crdt.NewDoc(value).GetTextValue("text"))

	if err != nil {
		t.Fatal("failed to get value from dht 2", err)
	}

	if !bytes.Equal(merged, value) {
		t.Fatal("bytes received from dht 2 not equal to merged")
	}
}
