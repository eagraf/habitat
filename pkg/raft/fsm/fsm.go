package fsm

import (
	"bufio"
	"io"

	"github.com/hashicorp/raft"
	"github.com/rs/zerolog/log"
)

type CommunityStateMachine struct {
	Log []string
}

func NewCommunityStateMachine() *CommunityStateMachine {
	return &CommunityStateMachine{
		Log: make([]string, 0),
	}
}

// Apply log is invoked once a log entry is committed.
// It returns a value which will be made available in the
// ApplyFuture returned by Raft.Apply method if that
// method was called on the same Raft node as the FSM.
func (sm *CommunityStateMachine) Apply(entry *raft.Log) interface{} {
	sm.Log = append(sm.Log, string(entry.Data))
	log.Info().Msgf("applying state transition %v", entry)
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
