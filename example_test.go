package deque

import "fmt"

func Example() {
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
}
