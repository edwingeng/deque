package deque

// Deque is a fast double-ended queue.
type Deque interface {
	// PushBack adds a new value v at the back of the Deque.
	PushBack(v interface{})
	// PushFront adds a new value v at the front of the Deque.
	PushFront(v interface{})
	// PopBack removes a value from the back of the Deque, and then returns
	// the removed value or nil if the Deque is empty.
	PopBack() interface{}
	// PopFront removes a value from the front of the Deque, and then returns
	// the removed value or nil if the Deque is empty.
	PopFront() interface{}
	// Back returns the last value at the back of the Deque or nil if the Deque is empty.
	Back() interface{}
	// Front returns the first value at the front of the Deque or nil if the Deque is empty.
	Front() interface{}
	// Empty returns whether the Deque is empty.
	Empty() bool
	// Len returns the number of values in the Deque.
	Len() int
}
