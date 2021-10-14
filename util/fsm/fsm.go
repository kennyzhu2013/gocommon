/*
@Time : 2019/12/25 16:28
@Author : kenny zhu
@File : fsm
@Software: GoLand
@Others:
*/
package fsm

import (
	// "strings"
	// "sync"
	"sync"
)

// interface for the FSM  transition function.
// type transitionFunc interface {
// 	transition(*FSM) error
// }

// FSM is the state machine that holds the current state.
// Not lock and goroutines-unsafe.
type FSM struct {
	// current is the state that the FSM is currently in.
	current   int
	statusAny int
	// eKey = events and source states -> destination states.
	transitions []eKey
	// checkBeforeCallbacks CheckCall // must use []checkCall
	checkBeforeCallbacks []checkCall
	// target state -> callback functions.
	// callbacks map[int]Callback
	callbacks []callBack //

	// current transition func.
	// transition func()
	// // transitionerObj calls the FSM' transition() function.
	// transitionerObj transitionFunc

	// stateMu guards access to the current state.
	stateMu sync.Mutex // RWMutex
	// eventMu guards access to Event() and Transition().
	eventMu sync.Mutex
}

// EventDesc represents an event when initializing the FSM.
type EventDesc struct {
	// Name is the event name used when calling for a transition.
	Name string

	// Src is a slice of source states that the FSM must be in to perform a state transition.
	Src []int

	// Dst is the destination state that the FSM will be in if the transition success.
	Dst int
}

// Callback is a function type that callbacks should use. Event is the current
// event info as the callback happens.
type Callback func(*Event)
type CheckCall func(*Event) bool // false not process any more.

// Events is a shorthand for defining the transition map in NewFSM.
type Events []EventDesc

// Callbacks is a shorthand for defining the callbacks in NewFSM.a
type Callbacks map[int]Callback
type CheckCalls map[int]CheckCall

// NewFSM constructs a FSM from events and callbacks.
// Callbacks are added as a map specified as Callbacks where the key is parsed
// as the callback event as follows, and called in the same order:
// 1. before_event - called before event
// 2. leave_state - called before leaving old_state
// 3. enter_state - called after entering new_state
// 4. after_event - called after event
//
// Default
// enter_state
//
func NewFSM(initial int, events []EventDesc, callbacks map[int]Callback, eventCheckCallbacks map[int]CheckCall) *FSM {
	f := &FSM{
		// transitionerObj: &transitionerStruct{},
		current: initial,
		// transitions:     new([]eKey),
		// callbacks:       make([]callBack, len(callbacks)),
		// checkBeforeCallbacks:  make(map[int]CheckCall, len(eventCheckCallbacks)),
	}

	// Build transition map and store sets of all events and states.
	// allEvents := make(map[string]bool)
	// allStates := make(map[int]bool)
	for _, e := range events {
		for _, src := range e.Src {
			ts := eKey{e.Name, src, e.Dst}
			f.transitions = append(f.transitions, ts)
			// f.transitions[eKey{e.Name, src, e.Dst}] = e.Dst
			// allStates[src] = true
			// allStates[e.Dst] = true
		}
		// allEvents[e.Name] = true
	}

	// Map all callbacks to events/states.
	for dst, fn := range callbacks {
		if dst > 0 {
			cb := callBack{fn, dst}
			// f.callbacks[dst] = fn
			f.callbacks = append(f.callbacks, cb)
		}
	}

	// map all before check callbacks
	for dst, fn := range eventCheckCallbacks {
		if dst >= 0 {
			cb := checkCall{fn, dst}
			f.checkBeforeCallbacks = append(f.checkBeforeCallbacks, cb)
		}
	}
	f.statusAny = -1
	return f
}

// Current returns the current state of the FSM.
func (f *FSM) Current() int {
	return f.current
}

// Current returns the current state of the FSM.
func (f *FSM) Reset() {
	f.current = 0
}

// Is returns true if state is the current state.
func (f *FSM) Is(state int) bool {
	return state == f.current
}

func (f *FSM) SetAnyState(state int) {
	f.statusAny = state
}

// SetState allows the user to move to the given state from current state.
// The call does not trigger any callbacks, if defined.
func (f *FSM) SetState(state int) {
	f.stateMu.Lock()
	defer f.stateMu.Unlock()
	f.current = state
	return
}

// Can returns dest state if event can occur in the current state.
func (f *FSM) GetDestState(event string) (int, bool) {
	for _, ts := range f.transitions {
		if ts.event == event && (ts.src == f.current || ts.src == f.statusAny) {
			return ts.dst, true
		}
	}
	// state, ok := f.transitions[eKey{event, f.current}]
	return 0, false //  && (f.transition == nil)
}

// Can returns true if event can occur in the current state.
func (f *FSM) Can(event string) bool {
	// f.stateMu.RLock()
	// defer f.stateMu.RUnlock()
	for _, ts := range f.transitions {
		if ts.event == event && (ts.src == f.current || ts.src == f.statusAny) {
			return true
		}
	}

	return false //  && (f.transition == nil)
}

// AvailableTransitions returns a list of transitions available in the current state.
func (f *FSM) AvailableTransitions() []string {
	// f.stateMu.RLock()
	// defer f.stateMu.RUnlock()
	var transitions []string
	for _, key := range f.transitions {
		if key.src == f.current || key.src == f.statusAny {
			transitions = append(transitions, key.event)
		}
	}
	return transitions
}

// Cannot returns true if event can not occure in the current state.
// It is a convenience method to help code read nicely.
func (f *FSM) Cannot(event string) bool {
	return !f.Can(event)
}

// not change, not thread safe.
func (f *FSM) EventNoArgs(e Event) error {
	// f.stateMu.Lock()
	// defer f.stateMu.Unlock()
	// f.eventMu.Lock()
	// defer f.eventMu.Unlock()

	dst, ok := f.GetDestState(e.Event)
	if !ok {
		for _, eke := range f.transitions {
			if eke.event == e.Event {
				return InvalidEventError{e.Event, f.current}
			}
		}
		return UnknownEventError{e.Event}
	}

	if f.beforeEventCallbacks(&e) == false {
		return InvalidEventError{e.Event, f.current}
	}

	// Perform the transition synchronous.
	// f.stateMu.RUnlock()
	// f.current = dst
	f.SetState(dst)
	err := f.enterStateCallbacks(&e)

	// f.stateMu.RLock()
	if err != nil {
		return err
	}

	return e.Err
}

// Event initiates a state transition with the named event.
func (f *FSM) Event(event string, seq int, args ...interface{}) error {
	f.eventMu.Lock()
	defer f.eventMu.Unlock()
	// f.stateMu.Lock()
	// defer f.stateMu.Unlock()
	// if f.transition != nil {
	// 	return InTransitionError{event}
	// }

	// dst, ok := f.transitions[eKey{event, f.current}]
	dst, ok := f.GetDestState(event)
	if !ok {
		for _, eke := range f.transitions {
			if eke.event == event {
				return InvalidEventError{event, f.current}
			}
		}
		return UnknownEventError{event}
	}

	e := &Event{f, event, f.current, dst, nil, seq, args}

	if f.beforeEventCallbacks(e) == false {
		return InvalidEventError{event, f.current}
	}
	// if f.current == dst {
	// 	f.afterEventCallbacks(e)
	// 	return NoTransitionError{e.Err}
	// }
	// Setup the transition, call it later.
	// f.transition = func() {
	// 	// f.stateMu.Lock()
	// 	// f.current = dst
	// 	// f.stateMu.Unlock()
	// 	f.enterStateCallbacks(e)
	// 	// f.afterEventCallbacks(e)
	// }

	// Perform the transition synchronous.
	// f.stateMu.RUnlock()
	// f.current = dst
	f.SetState(dst)
	err := f.enterStateCallbacks(e)

	// f.stateMu.RLock()
	if err != nil {
		return err
	}

	return e.Err
}

// Transition wraps transitioner.transition.
// func (f *FSM) Transition() error {
// 	// f.eventMu.Lock()
// 	// defer f.eventMu.Unlock()
// 	return f.doTransition()
// }
// doTransition wraps transitioner.transition.
// func (f *FSM) doTransition() error {
// 	return f.transitionerObj.transition(f)
// }
// // transition struct is the default implementation of the transition-er
// // interface. Other implementations can be swapped in for testing.
// type transitionerStruct struct{}
//
// // Transition completes an asynchronous state change.
// // The callback for leave_<STATE> must prviously have called Async on its
// // event to have initiated an asynchronous state transition.
// func (t transitionerStruct) transition(f *FSM) error {
// 	if f.transition == nil {
// 		return NotInTransitionError{}
// 	}
// 	f.transition()
// 	f.transition = nil
// 	return nil
// }
//
func (f *FSM) enterStateCallbacks(e *Event) error {
	for _, ts := range f.callbacks {
		if ts.status == f.current {
			ts.cb(e)
			if e.Err != nil {
				return e.Err
			}

			// only call once
			return nil
		}
	}
	// if fn, ok := f.callbacks[f.current]; ok {
	// 	// state change in fn function
	// 	fn(e)
	// 	// if e.canceled {
	// 	// 	return CanceledError{e.Err}
	// 	// } else if e.async {
	// 	// 	return AsyncError{e.Err}
	// 	// }
	// 	if e.Err != nil {
	// 		return e.Err
	// 	}
	// }

	// // default call
	// if fn, ok := f.callbacks[0]; ok {
	// 	fn(e)
	// }
	return nil
}

// 0 is default for all check if defined.
func (f *FSM) beforeEventCallbacks(e *Event) bool {
	for _, ts := range f.checkBeforeCallbacks {
		if ts.status == f.current {
			return ts.cc(e)
		} else if ts.status == 0 {
			return ts.cc(e)
		}
	}

	// if fn, ok := f.checkBeforeCallbacks[f.current]; ok {
	// 	return fn(e)
	// }
	//
	// if fn, ok := f.checkBeforeCallbacks[0]; ok {
	// 	return fn(e)
	// }

	return true
}

// eKey is a struct key used for storing the transition map.
type eKey struct {
	// event is the name of the event that the keys refers to.
	event string

	// src is the source from where the event can transition.
	src int
	dst int
}

type callBack struct {
	// event is the name of the event that the keys refers to.
	cb Callback

	// src is the source from where the event can transition.
	status int
}

type checkCall struct {
	cc     CheckCall
	status int
}
