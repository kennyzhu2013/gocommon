package event

import (
	"fmt"
	"sort"
)

// Listener interface
type Listener interface {
	Handle(e Event) error
}

// ListenerFunc func definition.
type ListenerFunc func(e Event) error

// Handle event. implements the Listener interface
func (fn ListenerFunc) Handle(e Event) error {
	return fn(e)
}

// you can register multi event listeners in a struct func.
// 外部订阅者接收event接口, 一系列Listener集合.
type Subscriber interface {
	// SubscribedEvents register event listeners
	// key: is event name
	// value: can be Listener or ListenerItem interface
	SubscribedEvents() []ListenerItem
}

// ListenerItem storage a event listener and it's priority value.
type ListenerItem struct {
	Key      string // event name.
	Priority int
	Listener Listener
}

/*************************************************************
 * Listener Queue
 *************************************************************/
// ListenerQueue storage sorted Listener instance.
type ListenerQueue struct {
	items []*ListenerItem
}

// Len get items length
func (lq *ListenerQueue) Len() int {
	return len(lq.items)
}

// IsEmpty get items length == 0
func (lq *ListenerQueue) IsEmpty() bool {
	return len(lq.items) == 0
}

// Push get items length
func (lq *ListenerQueue) Push(li *ListenerItem) *ListenerQueue {
	lq.items = append(lq.items, li)
	return lq
}

// Sort the queue items by ListenerItem's priority.
// Priority:
// 	High > Low
func (lq *ListenerQueue) Sort() *ListenerQueue {
	// if lq.IsEmpty() {
	// 	return lq
	// }
	ls := ByPriorityItems(lq.items)

	// check items is sorted
	if !sort.IsSorted(ls) {
		sort.Sort(ls)
	}

	return lq
}

// Items get all ListenerItem
func (lq *ListenerQueue) Items() []*ListenerItem {
	return lq.items
}

// Remove a listener from the queue
func (lq *ListenerQueue) Remove(listener Listener) {
	if listener == nil {
		return
	}

	ptrVal := fmt.Sprintf("%p", listener)

	// pointer vs..
	var newItems []*ListenerItem
	for _, li := range lq.items {
		if fmt.Sprintf("%p", li.Listener) == ptrVal {
			continue
		}

		newItems = append(newItems, li)
	}

	lq.items = newItems
}

// Clear clear all listeners
func (lq *ListenerQueue) Clear() {
	lq.items = lq.items[:0]
}

/*************************************************************
 * Sorted PriorityItems
 *************************************************************/
// There are some default priority constants
const (
	Normal = 0
	High   = 200
	Max    = 300
)

// ByPriorityItems type. implements the sort.Interface
type ByPriorityItems []*ListenerItem

// Len get items length
func (ls ByPriorityItems) Len() int {
	return len(ls)
}

// Less implements the sort.Interface.Less.
func (ls ByPriorityItems) Less(i, j int) bool {
	return ls[i].Priority > ls[j].Priority
}

// Swap implements the sort.Interface.Swap.
func (ls ByPriorityItems) Swap(i, j int) {
	ls[i], ls[j] = ls[j], ls[i]
}
