// Package deque implements a highly optimized double-ended queue, which is
// much efficient compared with list.List when adding or removing elements from
// the beginning or the end.
package deque

type Elem = interface{}

// Deque is a highly optimized double-ended queue.
type Deque interface {
	// PushBack adds a new value v at the back of Deque.
	PushBack(v Elem)
	// PushFront adds a new value v at the front of Deque.
	PushFront(v Elem)
	// PopBack removes a value from the back of Deque and returns
	// the removed value or nil if the Deque is empty.
	PopBack() Elem
	// PopFront removes a value from the front of Deque and returns
	// the removed value or nil if the Deque is empty.
	PopFront() Elem
	// Back returns the last value of Deque or nil if the Deque is empty.
	Back() Elem
	// Front returns the first value of Deque or nil if the Deque is empty.
	Front() Elem
	// Empty returns whether Deque is empty.
	Empty() bool
	// Len returns the number of values in Deque.
	Len() int

	// Enqueue is an alias of PushBack.
	Enqueue(v Elem)
	// Dequeue is an alias of PopFront.
	Dequeue() Elem
	// DequeueMany removes a number of values from the front of Deque and
	// returns the removed values or nil if the Deque is empty.
	// If max <= 0, DequeueMany removes and returns all the values in Deque.
	DequeueMany(max int) []Elem
	// DequeueManyWithBuffer is similar to DequeueMany except that it uses
	// buf to store the removed values as long as it has enough space.
	DequeueManyWithBuffer(max int, buf []Elem) []Elem

	// Range iterates all the values in Deque.
	Range(f func(i int, v Elem) bool)

	// Peek returns the value at idx.
	Peek(idx int) Elem
	// Replace replaces the value at idx.
	Replace(idx int, v Elem)
}
