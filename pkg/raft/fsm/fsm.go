package fsm

import (
	"bufio"
	"encoding/base64"
	"io"

	jsonpatch "github.com/evanphx/json-patch"
	"github.com/hashicorp/raft"
	"github.com/rs/zerolog/log"
)

type CommunityStateMachine struct {
	Log       []string `json:"log"`
	StateJSON []byte
}

func NewCommunityStateMachine() *CommunityStateMachine {
	return &CommunityStateMachine{
		Log:       make([]string, 0),
		StateJSON: []byte("{}"),
	}
}

// Apply log is invoked once a log entry is committed.
// It returns a value which will be made available in the
// ApplyFuture returned by Raft.Apply method if that
// method was called on the same Raft node as the FSM.
func (sm *CommunityStateMachine) Apply(entry *raft.Log) interface{} {
	sm.Log = append(sm.Log, string(entry.Data))
	log.Info().Msgf("applying state transition %v", entry)

	// Base64 decode so that no weird JSON quoting interferes along the way
	decoded, err := base64.StdEncoding.DecodeString(string(entry.Data))
	if err != nil {
		log.Error().Msgf("error decoding log entry data: %s")
	}

	// TODO patching should be validated by the leader beforehand so there
	// is no chance of an error here
	patch, err := jsonpatch.DecodePatch(decoded)
	if err != nil {
		log.Error().Msgf("error decoding JSON patch: %s", err)
	}

	modified, err := patch.Apply(sm.StateJSON)
	if err != nil {
		log.Error().Msgf("error applying JSON patch: %s", err)
	}

	sm.StateJSON = modified

	return nil
}

// Snapshot is used to support log compaction. This call should
// return an FSMSnapshot which can be used to save a point-in-time
// snapshot of the FSM. Apply and Snapshot are not called in multiple
// threads, but Apply will be called concurrently with Persist. This means
// the FSM should be implemented in a fashion that allows for concurrent
// updates while a snapshot is happening.
func (sm *CommunityStateMachine) Snapshot() (raft.FSMSnapshot, error) {
	return &CSMSnapshot{Log: sm.Log}, nil
}

// Restore is used to restore an FSM from a snapshot. It is not called
// concurrently with any other command. The FSM must discard all previous
// state.
func (sm *CommunityStateMachine) Restore(reader io.ReadCloser) error {
	newLog := make([]string, 0)
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		newLog = append(newLog, scanner.Text())
	}
	sm.Log = newLog
	return scanner.Err()
}

func (sm *CommunityStateMachine) State() ([]byte, error) {
	return sm.StateJSON, nil
}

type CSMSnapshot struct {
	Log []string
}

// Persist should dump all necessary state to the WriteCloser 'sink',
// and call sink.Close() when finished or call sink.Cancel() on error.
func (s *CSMSnapshot) Persist(sink raft.SnapshotSink) error {
	for _, entry := range s.Log {
		_, err := sink.Write([]byte(entry + "\n"))
		if err != nil {
			return err
		}
	}
	return sink.Close()
}

// Release is invoked when we are finished with the snapshot.
func (s *CSMSnapshot) Release() {
	return
}
