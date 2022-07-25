**Please use [the generic version](v2), if you have golang 1.18 or above.**

# Overview
Deque is a highly optimized double-ended queue, which is
much efficient compared with `list.List` when adding or removing elements from
the beginning or the end.

# Benchmark
```
PushBack/Deque<harden>       100000000       12.0 ns/op       8 B/op      0 allocs/op
PushBack/Deque                20000000       55.5 ns/op      24 B/op      1 allocs/op
PushBack/list.List             5000000      158.7 ns/op      56 B/op      1 allocs/op

PushFront/Deque<harden>      195840157        9.2 ns/op       8 B/op      0 allocs/op
PushFront/Deque               30000000       49.2 ns/op      24 B/op      1 allocs/op
PushFront/list.List            5000000      159.2 ns/op      56 B/op      1 allocs/op

Random/Deque<harden>          65623633       15.1 ns/op       0 B/op      0 allocs/op
Random/Deque                  50000000       24.7 ns/op       4 B/op      0 allocs/op
Random/list.List              30000000       46.9 ns/op      28 B/op      1 allocs/op
```

# Getting started
```
go get -u github.com/edwingeng/deque
```

# Usage
``` go
import "github.com/edwingeng/deque"

dq := deque.NewDeque()
dq.PushBack(100)
dq.PushBack(200)
dq.PushBack(300)
for !dq.Empty() {
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

# Harden the element data type
``` bash
./harden.sh <outputDir> <packageName> [elemType]
```
