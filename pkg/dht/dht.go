package dht

import (
	"context"

	"github.com/ipfs/go-datastore"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/host"
)

type DHT struct {
	*dht.IpfsDHT
}

func NewDht(h host.Host) (*DHT, error) {
	kaddht, err := dht.New(
		context.Background(),
		h,
		dht.Datastore(datastore.NewLogDatastore(datastore.NewMapDatastore(), "datastore")),
		dht.NamespacedReducer("crdt", CrdtReducer{}),
		dht.ProtocolPrefix("/hab"),
		dht.Mode(dht.ModeServer),
	)

	if err != nil {
		return nil, err
	}

	return &DHT{
		IpfsDHT: kaddht,
	}, nil
}
