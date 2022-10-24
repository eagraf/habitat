package raft

import "encoding/base64"

type RaftDispatcher struct {
	communityID    string
	clusterService *ClusterService
}

func (d *RaftDispatcher) Dispatch(json []byte) error {
	encoded := base64.StdEncoding.EncodeToString(json)
	return d.clusterService.ProposeTransitions(d.communityID, []byte(encoded))
}
