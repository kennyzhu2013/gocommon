package main

import (
	"common/event"
	"fmt"
)

func main() {
	// Register event listener
	event.On("evt1", event.ListenerFunc(func(e event.Event) error {
		fmt.Printf("handle event 1: %s parms:%s\n", e.Name(), e.Data())
		return nil
	}), event.Normal)

	// Register multiple listeners
	event.On("evt1", event.ListenerFunc(func(e event.Event) error {
		fmt.Printf("handle event 2: %s parms:%s\n", e.Name(), e.Data())
		return nil
	}), event.High)

	// ... ...

	// Trigger event
	// Note: The second listener has a higher priority, so it will be executed first.
	event.MustFire("evt1", event.M{"arg0", "arg1"})

	evt1 := event.NewBasic("evt1", nil).Fill("123456789", event.M{"inhere", "outofhere"})
	event.AddEvent(evt1)

	event.FireEvent(evt1)

}
