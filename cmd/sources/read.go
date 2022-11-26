package sources

import (
	"encoding/json"

	"github.com/rs/zerolog/log"
)

type ReadRequest struct {
	Requester string `json:"requester"` // for ex: name of app
	Community string `json:"community"` // eventually should be community id
	// Token      []byte     `json:"token"`     // token for permissions
	SourceName SourceName `json:"id"` // eventually this should be a unique source id
}

type SourceReader interface {
	Read(SourceName) ([]byte, error) // return (error, data)
}

type Reader struct {
	SourceReader       SourceReader
	PermissionsManager PermissionsManager
}

func NewReader(S SourceReader, P PermissionsManager) *Reader {
	return &Reader{SourceReader: S, PermissionsManager: P}
}

// return (allowed, error, data)
func (R *Reader) Read(r *ReadRequest) (bool, error, json.RawMessage) {
	if !R.PermissionsManager.CheckCanRead(r) {
		return false, nil, nil
	}

	data, err := R.SourceReader.Read(r.SourceName)
	if err != nil {
		log.Error().Msgf("Error reading source: %s", err.Error())
	}
	return true, err, data
}
