package sources

type ReadRequest struct {
	Requester string `json:"requester"` // for ex: name of app
	Source    Source `json:"source"`    // request by schema for v0
}

type SourceReader interface {
	Read(Source) (error, SourceData) // return (error, data)
}

type Reader struct {
	SourceReader       SourceReader
	PermissionsManager PermissionsManager
}

func NewReader(S SourceReader, P PermissionsManager) *Reader {
	return &Reader{SourceReader: S, PermissionsManager: P}
}

// return (allowed, error, data)
func (R *Reader) Read(r *ReadRequest) (bool, error, SourceData) {
	if !R.PermissionsManager.CheckCanRead(r) {
		return false, nil, ""
	}

	err, data := R.SourceReader.Read(r.Source)
	return true, err, data
}
