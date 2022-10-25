package state

import "github.com/eagraf/habitat/structs/community"

type Dispatcher interface {
	Dispatch(patch []byte) (*community.CommunityState, error)
}

type LocalDispatcher struct {
	jsonState *JSONState
}

func (d *LocalDispatcher) Dispatch(patch []byte) error {
	return d.jsonState.ApplyPatch(patch)
}
