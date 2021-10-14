package event

import (
	"errors"
	"strconv"
)

// Event interface
type Event interface {
	Name() string
	// Target() interface{}
	Get(index int) interface{}
	Set(index int, val interface{}) error
	Add(val interface{})
	Data() []interface{}
	SetData(M) Event
	Abort(bool)
	IsAborted() bool
}

// BasicEvent a basic event struct define.
type BasicEvent struct {
	// event name
	name string

	// user data, just array.
	data []interface{}

	// target, use id for every thing.
	target string // for handle func:

	// mark is aborted
	aborted bool
}

// NewBasic new an basic event instance
func NewBasic(name string, data M) *BasicEvent {
	if data == nil {
		data = make([]interface{}, 1)
	}

	return &BasicEvent{
		name: name,
		data: data,
	}
}

// Abort abort event loop exec
func (e *BasicEvent) Abort(abort bool) {
	e.aborted = abort
}

// Fill event data
func (e *BasicEvent) Fill(target string, data M) *BasicEvent {
	if data != nil {
		e.data = data
	}

	e.target = target
	return e
}

// AttachTo add current event to the event manager.
func (e *BasicEvent) AttachTo(em ManagerFace) {
	em.AddEvent(e)
}

// Get get data by index
func (e *BasicEvent) Get(index int) interface{} {
	if len(e.data) < (index + 1) {
		return nil
	}

	return e.data[index]
}

// Add value by key
func (e *BasicEvent) Add(val interface{}) {
	e.data = append(e.data, val)
}

// Set value by key
func (e *BasicEvent) Set(index int, val interface{}) error {
	if len(e.data) == 0 || len(e.data) < index+1 {
		return errors.New("data not exist for " + strconv.Itoa(index))
	}
	for i := 0; i < len(e.data); i++ {
		if i == index {
			e.data[i] = val
			break
		}
	}
	return nil
}

// Name get event name
func (e *BasicEvent) Name() string {
	return e.name
}

// Data get all data
func (e *BasicEvent) Data() []interface{} {
	return e.data
}

// IsAborted check.
func (e *BasicEvent) IsAborted() bool {
	return e.aborted
}

// Target get target
func (e *BasicEvent) Target() interface{} {
	return e.target
}

// SetName set event name
func (e *BasicEvent) SetName(name string) *BasicEvent {
	e.name = name
	return e
}

// SetData set data to the event
func (e *BasicEvent) SetData(data M) Event {
	if data != nil {
		e.data = data
	}
	return e
}

// SetTarget set event target
func (e *BasicEvent) SetTarget(target string) *BasicEvent {
	e.target = target
	return e
}
