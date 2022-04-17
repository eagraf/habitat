package state

import (
	"encoding/base64"

	"github.com/eagraf/habitat/cmd/habitat/community/consensus/raft"
)

type Dispatcher interface {
	Dispatch(patch []byte) error
}

type LocalDispatcher struct {
	jsonState *JSONState
}

func (d *LocalDispatcher) Dispatch(patch []byte) error {
	return d.jsonState.ApplyPatch(patch)
}

type RaftDispatcher struct {
	communityID    string
	clusterService *raft.ClusterService
}

func (d *RaftDispatcher) Dispatch(patch []byte) error {
	encoded := base64.StdEncoding.EncodeToString(patch)
	return d.clusterService.ProposeTransition(d.communityID, []byte(encoded))
}
