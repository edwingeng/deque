# Overview
Deque is a highly optimized double-ended queue.

# Benchmark
```
PushBack/Deque                20000000       13.6 ns/op       8 B/op      0 allocs/op
PushBack/list.List             5000000      158.7 ns/op      56 B/op      1 allocs/op

PushFront/Deque               30000000        9.8 ns/op       8 B/op      0 allocs/op
PushFront/list.List            5000000      159.2 ns/op      56 B/op      1 allocs/op

Random/Deque                  50000000       13.9 ns/op       0 B/op      0 allocs/op
Random/list.List              30000000       46.9 ns/op      28 B/op      1 allocs/op
```

# Installation
```
go get -u github.com/edwingeng/deque/v2
```

# Usage
``` go
import "github.com/edwingeng/deque/v2"

dq := deque.NewDeque[int]()
dq.PushBack(100)
dq.PushBack(200)
dq.PushBack(300)
for !dq.IsEmpty() {
    fmt.Println(dq.PopFront())
}

dq.PushFront(100)
dq.PushFront(200)
dq.PushFront(300)
for i, n := 0, dq.Len(); i < n; i++ {
    fmt.Println(dq.PopFront())
}

// Output:
// 100
// 200
// 300
// 300
// 200
// 100
```

# Documentation
```
type Deque[T any] struct {
        // Has unexported fields.
}

func NewDeque[T any]() *Deque[T]
    NewDeque creates a new Deque instance.

func (dq *Deque[T]) PushBack(v T)
    PushBack adds a new value v at the back of Deque.

func (dq *Deque[T]) PushFront(v T)
    PushFront adds a new value v at the front of Deque.

func (dq *Deque[T]) PopBack() T
    PopBack removes a value from the back of Deque and returns the removed
    value. It panics if Deque is empty.

func (dq *Deque[T]) PopFront() T
    PopFront removes a value from the front of Deque and returns the removed
    value. It panics if Deque is empty.

func (dq *Deque[T]) TryPopBack() (_ T, ok bool)
    TryPopBack tries to remove a value from the back of Deque and returns the
    removed value. ok is false if Deque is empty.

func (dq *Deque[T]) TryPopFront() (_ T, ok bool)
    TryPopFront tries to remove a value from the front of Deque and returns the
    removed value. ok is false if Deque is empty.

func (dq *Deque[T]) Back() (_ T, ok bool)
    Back returns the last value of Deque. ok is false if Deque is empty.

func (dq *Deque[T]) Front() (_ T, ok bool)
    Front returns the first value of Deque. ok is false if Deque is empty.

func (dq *Deque[T]) IsEmpty() bool
    IsEmpty returns whether Deque is empty.

func (dq *Deque[T]) Len() int
    Len returns the number of values in Deque.

func (dq *Deque[T]) Enqueue(v T)
    Enqueue is an alias of PushBack.

func (dq *Deque[T]) Dequeue() T
    Dequeue is an alias of PopFront.

func (dq *Deque[T]) TryDequeue() (_ T, ok bool)
    TryDequeue is an alias of TryPopFront.

func (dq *Deque[T]) DequeueMany(max int) []T
    DequeueMany removes a number of values from the front of Deque and returns
    the removed values or nil if the Deque is empty. If max <= 0, DequeueMany
    removes and returns all the values in Deque.

func (dq *Deque[T]) DequeueManyWithBuffer(max int, buf []T) []T
    DequeueManyWithBuffer is similar to DequeueMany except that it uses buf to
    store the removed values as long as it has enough space.

func (dq *Deque[T]) Range(f func(i int, v T) bool)
    Range iterates all the values in Deque.

func (dq *Deque[T]) Peek(idx int) T
    Peek returns the value at idx.

func (dq *Deque[T]) Replace(idx int, v T)
    Replace replaces the value at idx.

func (dq *Deque[T]) Dump() []T
    Dump returns all the values in Deque.
```