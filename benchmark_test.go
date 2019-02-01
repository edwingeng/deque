package deque

import (
	"container/list"
	"math/rand"
	"testing"
)

func Benchmark_PushBack(b *testing.B) {
	b.Run("deque", func(b *testing.B) {
		dq := NewDeque()
		for i := 0; i < b.N; i++ {
			dq.PushBack(i)
		}
	})
	b.Run("list", func(b *testing.B) {
		lst := list.New()
		for i := 0; i < b.N; i++ {
			lst.PushBack(i)
		}
	})
}

func Benchmark_PushFront(b *testing.B) {
	b.Run("deque", func(b *testing.B) {
		dq := NewDeque()
		for i := 0; i < b.N; i++ {
			dq.PushFront(i)
		}
	})
	b.Run("list", func(b *testing.B) {
		lst := list.New()
		for i := 0; i < b.N; i++ {
			lst.PushFront(i)
		}
	})
}

func Benchmark_Random(b *testing.B) {
	const na = 100000
	a := make([]int, na)
	for i := 0; i < na; i++ {
		a[i] = rand.Int()
	}

	b.Run("deque", func(b *testing.B) {
		dq := NewDeque()
		for i := 0; i < b.N; i++ {
			switch a[i%na] % 4 {
			case 0:
				dq.PushBack(i)
			case 1:
				dq.PushFront(i)
			case 2:
				dq.PopBack()
			case 3:
				dq.PopFront()
			}
		}
	})
	b.Run("list", func(b *testing.B) {
		lst := list.New()
		for i := 0; i < b.N; i++ {
			switch a[i%na] % 4 {
			case 0:
				lst.PushBack(i)
			case 1:
				lst.PushFront(i)
			case 2:
				if v := lst.Back(); v != nil {
					lst.Remove(v)
				}
			case 3:
				if v := lst.Front(); v != nil {
					lst.Remove(v)
				}
			}
		}
	})
}
