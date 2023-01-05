package dht

import (
	"github.com/eagraf/habitat/pkg/crdt"
	"github.com/libp2p/go-libp2p-kad-dht/reducer"
)

type CrdtReducer struct{}

func (r CrdtReducer) Validate(key string, value []byte) error {
	return nil
}
func (r CrdtReducer) Reduce(key string, values [][]byte) ([]byte, int, error) {
	doc := crdt.NewDoc(values[0])
	for _, v := range values[1:] {
		doc.Merge(v)
	}

	return doc.State(), -1, nil
}

var _ reducer.Reducer = CrdtReducer{}
