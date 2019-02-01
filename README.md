# Overview
Deque is a fast double-ended queue.

[![Build Status](https://travis-ci.com/edwingeng/deque.svg?branch=master)](https://travis-ci.com/edwingeng/deque)

# Benchmark
```
Benchmark_PushBack/deque         	20000000	        75.3 ns/op	      24 B/op	       1 allocs/op
Benchmark_PushBack/list          	 5000000	       280 ns/op	      56 B/op	       2 allocs/op
Benchmark_PushFront/deque        	30000000	        68.5 ns/op	      24 B/op	       1 allocs/op
Benchmark_PushFront/list         	 5000000	       279 ns/op	      56 B/op	       2 allocs/op
Benchmark_Random/deque           	50000000	        28.8 ns/op	       3 B/op	       0 allocs/op
Benchmark_Random/list            	30000000	        51.0 ns/op	      27 B/op	       0 allocs/op
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
