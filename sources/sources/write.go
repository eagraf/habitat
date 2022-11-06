package sources

import "github.com/eagraf/habitat/pkg/sources"

type WriteRequest struct {
	Requester string             `json:"requester"` // for ex: name of app
	Source    sources.Source     `json:"source"`    // request by name and hash/version
	Community string             `json:"community"`
	Data      sources.SourceData `json:"data"`
}

type SourceWriter interface {
	Write(sources.Source, sources.SourceData) error // take in write request, return (allowed, error)
}

type Writer struct {
	SourceWriter       SourceWriter
	PermissionsManager PermissionsManager
}

func NewWriter(S SourceWriter, P PermissionsManager) *Writer {
	return &Writer{SourceWriter: S, PermissionsManager: P}
}

// return (allowed, error, data)
func (W *Writer) Write(w *WriteRequest) (bool, error) {
	if !W.PermissionsManager.CheckCanWrite(w) {
		return false, nil
	}

	return true, W.SourceWriter.Write(w.Source, w.Data)
}
