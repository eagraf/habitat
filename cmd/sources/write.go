package sources

import "encoding/json"

type WriteRequest struct {
	// Token      []byte     `json:"token"`     // token for permissions
	SourceName SourceName      `json:"id"` // eventually this should be a unique source id
	Data       json.RawMessage `json:"data"`
}

type SourceWriter interface {
	Write(SourceName, []byte) error // take in write request, return (allowed, error)
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

	return true, W.SourceWriter.Write(w.SourceName, w.Data)
}
