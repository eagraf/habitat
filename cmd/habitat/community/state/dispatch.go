package state

import (
	"github.com/eagraf/habitat/structs/community"
)

type Dispatcher interface {
	Dispatch(patch []byte) (*community.CommunityState, error)
}
