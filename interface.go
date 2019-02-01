package deque

type Deque interface {
	PushBack(v interface{})
	PushFront(v interface{})
	PopBack() interface{}
	PopFront() interface{}
	Back() interface{}
	Front() interface{}
	Empty() bool
	Len() int
}
