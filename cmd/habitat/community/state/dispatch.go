package state

type Dispatcher interface {
	Dispatch(patch []byte) error
}

type LocalDispatcher struct {
	jsonState *JSONState
}

func (d *LocalDispatcher) Dispatch(patch []byte) error {
	return d.jsonState.ApplyPatch(patch)
}
