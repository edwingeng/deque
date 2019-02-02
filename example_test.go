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

	// Output:
	// 100
	// 200
	// 300
}
