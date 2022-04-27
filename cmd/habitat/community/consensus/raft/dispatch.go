package raft

import "encoding/base64"

type RaftDispatcher struct {
	communityID    string
	clusterService *ClusterService
}

func (d *RaftDispatcher) Dispatch(patch []byte) error {
	encoded := base64.StdEncoding.EncodeToString(patch)
	return d.clusterService.ProposeTransition(d.communityID, []byte(encoded))
}
