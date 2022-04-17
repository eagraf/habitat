package state

import (
	"errors"
	"fmt"
)

const (
	TransitionTypeInitializeCounter = "initialize_counter"
	TransitionTypeIncrementCounter  = "increment_counter"
)

type TransitionWrapper struct {
	Type       string `json:"type"`
	Transition []byte `json:"transition"`
}

type CommunityStateTransition interface {
	Type() string
	Patch(oldState *CommunityState) ([]byte, error)
	Validate(oldState *CommunityState) error
}

type InitializeCounterTransition struct {
	InitialCount int
}

func (t *InitializeCounterTransition) Type() string {
	return TransitionTypeInitializeCounter
}

func (t *InitializeCounterTransition) Patch(oldState *CommunityState) ([]byte, error) {
	return []byte(fmt.Sprintf(`[{
		"op": "add",
		"path": "/counter",
		"value": %d
	}]`, t.InitialCount)), nil
}

func (t *InitializeCounterTransition) Validate(oldState *CommunityState) error {
	if oldState.Counter != 0 {
		return errors.New("counter is not 0")
	}
	return nil
}

type IncrementCounterTransition struct{}

func (t *IncrementCounterTransition) Type() string {
	return TransitionTypeIncrementCounter
}

func (t *IncrementCounterTransition) Patch(oldState *CommunityState) ([]byte, error) {
	return []byte(fmt.Sprintf(`[{
		"op": "replace",
		"path": "/counter",
		"value": %d 
	}]`, oldState.Counter+1)), nil
}

func (t *IncrementCounterTransition) Validate(oldState *CommunityState) error {
	return nil
}
