# Overview
Deque is a highly optimized double-ended queue.

# Benchmark
```
PushBack/Deque<harden>       100000000       10.3 ns/op       9 B/op      0 allocs/op
PushBack/Deque                20000000       81.3 ns/op      24 B/op      1 allocs/op
PushBack/list.List             5000000      281   ns/op      56 B/op      2 allocs/op

PushFront/Deque<harden>      195840157        8.0 ns/op       9 B/op      0 allocs/op
PushFront/Deque               30000000       70.6 ns/op      24 B/op      1 allocs/op
PushFront/list.List            5000000      276   ns/op      56 B/op      2 allocs/op

Random/Deque<harden>          65623633       17.3 ns/op       0 B/op      0 allocs/op
Random/Deque                  50000000       32.1 ns/op       4 B/op      0 allocs/op
Random/list.List              30000000      123   ns/op      28 B/op      1 allocs/op
```

# Usage
``` go
import "github.com/edwingeng/deque"

dq := NewDeque()
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
