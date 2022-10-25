package raft

import (
	"encoding/base64"

	"github.com/eagraf/habitat/structs/community"
)

type RaftDispatcher struct {
	communityID    string
	clusterService *ClusterService
}

func (d *RaftDispatcher) Dispatch(json []byte) (*community.CommunityState, error) {
	encoded := base64.StdEncoding.EncodeToString(json)
	return d.clusterService.ProposeTransitions(d.communityID, []byte(encoded))
}
