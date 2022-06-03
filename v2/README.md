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

# Usage
``` go
import "github.com/edwingeng/deque/v2"

dq := NewDeque[int]()
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
