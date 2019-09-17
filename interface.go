package deque

// Deque is a fast double-ended queue.
type Deque interface {
	// PushBack adds a new value v at the back of Deque.
	PushBack(v interface{})
	// PushFront adds a new value v at the front of Deque.
	PushFront(v interface{})
	// PopBack removes a value from the back of Deque and returns
	// the removed value or nil if the Deque is empty.
	PopBack() interface{}
	// PopFront removes a value from the front of Deque and returns
	// the removed value or nil if the Deque is empty.
	PopFront() interface{}
	// Back returns the last value of Deque or nil if the Deque is empty.
	Back() interface{}
	// Front returns the first value of Deque or nil if the Deque is empty.
	Front() interface{}
	// Empty returns whether Deque is empty.
	Empty() bool
	// Len returns the number of values in Deque.
	Len() int

	// PopManyBack removes a number of values from the back of Deque and
	// returns the removed values or nil if the Deque is empty.
	// If max <= 0, PopManyBack removes and returns all the values in Deque.
	PopManyBack(max int) []interface{}
	// PopManyFront removes a number of values from the front of Deque and
	// returns the removed values or nil if the Deque is empty.
	// If max <= 0, PopManyFront removes and returns all the values in Deque.
	PopManyFront(max int) []interface{}

	// Enqueue is an alias of PushBack.
	Enqueue(v interface{})
	// Dequeue is an alias of PopFront.
	Dequeue() interface{}
	// DequeueMany is an alias of PopManyFront.
	DequeueMany(max int) []interface{}

	// Range iterates all of the values in Deque.
	Range(f func(v interface{}) bool)
}
