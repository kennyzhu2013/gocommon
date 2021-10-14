/*
@Time : 2019/12/25 16:32
@Author : kenny zhu
@File : event
@Software: GoLand
@Others:
*/
package fsm

// event based fsm
type Event struct {
	// reference to the current FSM.
	FSM *FSM
	// Event is the event name.
	Event string

	// Src is the state before the transition.
	Src int
	// Dst is the state after the transition.
	Dst int

	// Err is an optional error that can be returned from a callback.
	Err error
	// Args is a optional list of arguments passed to the callback.
	Seq int
	Args []interface{}


	// Extends :
	// async bool  // an internal flag set if the transition should be asynchronous
	// canceled bool // an internal flag set if the transition is canceled.
}

// Cancel can be called to cancel the current transition before it happens.
// func (e *Event) Cancel(err ...error) {
// 	e.canceled = true
//
// 	if len(err) > 0 {
// 		e.Err = err[0]
// 	}
// }
//
// // Async can be called to do an asynchronous state transition.
// func (e *Event) Async() {
// 	e.async = true
// }