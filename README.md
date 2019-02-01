# Overview
Deque is a fast double-ended queue.

# Benchmark
```
Benchmark_PushBack/deque         	20000000	        77.2 ns/op
Benchmark_PushBack/list          	 5000000	       280 ns/op
Benchmark_PushFront/deque        	30000000	        71.2 ns/op
Benchmark_PushFront/list         	 5000000	       276 ns/op
Benchmark_Random/deque           	50000000	        32.3 ns/op
Benchmark_Random/list            	30000000	       146 ns/op
```

# Interface
``` go
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
```
