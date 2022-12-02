package state

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"io/ioutil"

	"github.com/eagraf/habitat/structs/community"
	"github.com/hashicorp/raft"
	"github.com/rs/zerolog/log"
)

// TODO refactor this into the raft package
type RaftFSMAdapter struct {
	jsonState  *JSONState
	updateChan chan StateUpdate
}

func NewRaftFSMAdapter(commState []byte) (*RaftFSMAdapter, error) {
	jsonState, err := NewJSONState(community.CommunityStateSchema, commState)
	if err != nil {
		return nil, err
	}

	return &RaftFSMAdapter{
		jsonState:  jsonState,
		updateChan: make(chan StateUpdate),
	}, nil
}

func (sm *RaftFSMAdapter) JSONState() *JSONState {
	return sm.jsonState
}

func (sm *RaftFSMAdapter) UpdateChan() <-chan StateUpdate {
	return sm.updateChan
}

// Apply log is invoked once a log entry is committed.
// It returns a value which will be made available in the
// ApplyFuture returned by Raft.Apply method if that
// method was called on the same Raft node as the FSM.
func (sm *RaftFSMAdapter) Apply(entry *raft.Log) interface{} {
	buf, err := base64.StdEncoding.DecodeString(string(entry.Data))
	if err != nil {
		log.Error().Msgf("error decoding log entry data: %s", err)
	}

	var wrappers []*TransitionWrapper
	err = json.Unmarshal(buf, &wrappers)
	if err != nil {
		log.Error().Msgf("error unmarshaling transition wrapper: %s", err)
	}

	for _, w := range wrappers {
		err = sm.jsonState.ApplyPatch(w.Patch)
		if err != nil {
			log.Error().Msgf("error applying patch: %s", err)
		}

		sm.updateChan <- StateUpdate{
			TransitionType: w.Type,
			Transition:     w.Transition,
			NewState:       sm.jsonState.Bytes(),
		}
	}

	var state community.CommunityState
	err = json.Unmarshal(sm.jsonState.Bytes(), &state)
	if err != nil {
		log.Error().Msgf("error unmarshaling state after applying transitions: %s", err)
	}

	return &state
}

// Snapshot is used to support log compaction. This call should
// return an FSMSnapshot which can be used to save a point-in-time
// snapshot of the FSM. Apply and Snapshot are not called in multiple
// threads, but Apply will be called concurrently with Persist. This means
// the FSM should be implemented in a fashion that allows for concurrent
// updates while a snapshot is happening.
func (sm *RaftFSMAdapter) Snapshot() (raft.FSMSnapshot, error) {
	return &FSMSnapshot{
		state: sm.jsonState.Bytes(),
	}, nil
}

// Restore is used to restore an FSM from a snapshot. It is not called
// concurrently with any other command. The FSM must discard all previous
// state.
func (sm *RaftFSMAdapter) Restore(reader io.ReadCloser) error {
	/*newLog := make([]string, 0)
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		newLog = append(newLog, scanner.Text())
	}
	sm.Log = newLog
	return scanner.Err()*/
	buf, err := ioutil.ReadAll(reader)
	if err != nil {
		return err
	}

	state, err := NewJSONState(community.CommunityStateSchema, buf)
	if err != nil {
		return err
	}

	sm.jsonState = state
	return nil
}

func (sm *RaftFSMAdapter) State() ([]byte, error) {
	return sm.jsonState.Bytes(), nil
}

type FSMSnapshot struct {
	state []byte
}

// Persist should dump all necessary state to the WriteCloser 'sink',
// and call sink.Close() when finished or call sink.Cancel() on error.
func (s *FSMSnapshot) Persist(sink raft.SnapshotSink) error {
	/*for _, entry := range s.Log {
		_, err := sink.Write([]byte(entry + "\n"))
		if err != nil {
			return err
		}
	}*/
	sink.Write(s.state)
	return sink.Close()
}

// Release is invoked when we are finished with the snapshot.
func (s *FSMSnapshot) Release() {
}
