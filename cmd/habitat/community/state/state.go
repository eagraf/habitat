package state

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/eagraf/habitat/structs/community"
	jsonpatch "github.com/evanphx/json-patch"
	"github.com/qri-io/jsonschema"
	"github.com/rs/zerolog/log"
)

func keyError(errs []jsonschema.KeyError) error {
	s := strings.Builder{}
	for _, e := range errs {
		s.WriteString(fmt.Sprintf("%s\n", e.Error()))
	}
	return errors.New(s.String())
}

type Executor interface {
	Execute(*StateUpdate)
}

type NOOPExecutor struct{}

func (e *NOOPExecutor) Execute(update *StateUpdate) {
	log.Info().Msgf("executing %s update", update.TransitionType)
}

type StateUpdate struct {
	NewState       []byte
	TransitionType string
}

func (s *StateUpdate) State() (*community.CommunityState, error) {
	var res community.CommunityState
	err := json.Unmarshal(s.NewState, &res)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

type CommunityStateMachine struct {
	jsonState *JSONState // this JSONState is maintained in addition to
	//state *community.CommunityState
	dispatcher Dispatcher
	executor   Executor
	//transitionChan <-chan *CommunityStateTransition
	updateChan <-chan StateUpdate
	doneChan   chan bool
}

func NewCommunityStateMachine(initState *community.CommunityState, updateChan <-chan StateUpdate, dispatcher Dispatcher, executor Executor) (*CommunityStateMachine, error) {
	marshaled, err := json.Marshal(initState)
	if err != nil {
		return nil, err
	}
	jsonState, err := NewJSONState(community.CommunityStateSchema, marshaled)
	if err != nil {
		return nil, err
	}
	return &CommunityStateMachine{
		jsonState:  jsonState,
		updateChan: updateChan,
		dispatcher: dispatcher,
		doneChan:   make(chan bool),
		executor:   executor,
	}, nil
}

func (csm *CommunityStateMachine) StartListening() {
	go func() {
		for {
			select {
			case stateUpdate := <-csm.updateChan:
				// execute state update
				jsonState, err := NewJSONState(community.CommunityStateSchema, stateUpdate.NewState)
				if err != nil {
					log.Error().Err(err).Msgf("error getting new state from state update chan")
				}
				csm.jsonState = jsonState
				csm.executor.Execute(&stateUpdate)
			case <-csm.doneChan:
				return
			}
		}
	}()
}

func (csm *CommunityStateMachine) StopListening() {
	csm.doneChan <- true
}

func (csm *CommunityStateMachine) ProposeTransitions(transitions []CommunityStateTransition) (*community.CommunityState, error) {
	currentState, err := csm.State()
	if err != nil {
		return nil, err
	}

	jsonStateBranch, err := csm.jsonState.Copy()
	if err != nil {
		return nil, err
	}

	wrappers := make([]*TransitionWrapper, 0)

	for _, t := range transitions {

		err = t.Validate(currentState)
		if err != nil {
			return nil, fmt.Errorf("transition validation failed: %s", err)
		}

		patch, err := t.Patch(currentState)
		if err != nil {
			return nil, err
		}

		err = jsonStateBranch.ApplyPatch(patch)
		if err != nil {
			return nil, err
		}

		newStateBytes := jsonStateBranch.state

		var newState community.CommunityState
		err = json.Unmarshal(newStateBytes, &newState)
		if err != nil {
			return nil, err
		}

		wrapped, err := wrapTransition(t, currentState)
		if err != nil {
			return nil, err
		}

		wrappers = append(wrappers, wrapped)

		currentState = &newState
	}

	transitionsJSON, err := json.Marshal(wrappers)
	if err != nil {
		return nil, err
	}

	state, err := csm.dispatcher.Dispatch(transitionsJSON)
	if err != nil {
		return nil, err
	}

	return state, nil
}

func (csm *CommunityStateMachine) State() (*community.CommunityState, error) {
	var res community.CommunityState
	err := csm.jsonState.Unmarshal(&res)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

type JSONState struct {
	schema *jsonschema.Schema
	state  []byte

	*sync.Mutex
}

func NewJSONState(jsonSchema []byte, initState []byte) (*JSONState, error) {
	rs := &jsonschema.Schema{}
	err := json.Unmarshal(jsonSchema, rs)
	if err != nil {
		return nil, fmt.Errorf("invalid JSON schema: %s", err)
	}
	keyErrs, err := rs.ValidateBytes(context.Background(), initState)
	if err != nil {
		return nil, fmt.Errorf("error validating initial state: %s", err)
	}
	if len(keyErrs) != 0 {
		return nil, keyError(keyErrs)
	}

	return &JSONState{
		schema: rs,
		state:  initState,
		Mutex:  &sync.Mutex{},
	}, nil
}

func (s *JSONState) ApplyPatch(patchJSON []byte) error {
	updated, err := s.applyImpl(patchJSON)
	if err != nil {
		return err
	}

	// only update state if everything worked out
	s.Lock()
	defer s.Unlock()

	s.state = updated

	return nil
}

func (s *JSONState) ValidatePatch(patchJSON []byte) ([]byte, error) {
	updated, err := s.applyImpl(patchJSON)
	if err != nil {
		return nil, err
	}

	return updated, err
}

func (s *JSONState) applyImpl(patchJSON []byte) ([]byte, error) {
	patch, err := jsonpatch.DecodePatch(patchJSON)
	if err != nil {
		return nil, fmt.Errorf("invalid JSON patch: %s", err)
	}
	updated, err := patch.Apply(s.state)
	if err != nil {
		return nil, fmt.Errorf("error applying patch to current state")
	}
	// check that updated state still fulfills the schema
	keyErrs, err := s.schema.ValidateBytes(context.Background(), updated)
	if err != nil {
		return nil, fmt.Errorf("error validating updated state: %s", err)
	}
	if len(keyErrs) != 0 {
		return nil, keyError(keyErrs)
	}
	return updated, nil
}

func (s *JSONState) Unmarshal(dest interface{}) error {
	return json.Unmarshal(s.state, dest)
}

func (s *JSONState) Bytes() []byte {
	return s.state
}

func (s *JSONState) Copy() (*JSONState, error) {
	schema, err := json.Marshal(s.schema)
	if err != nil {
		return nil, err
	}
	return NewJSONState(schema, s.state)
}
