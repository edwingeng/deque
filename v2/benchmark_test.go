package deque

import (
	"container/list"
	"math/rand"
	"testing"
)

func BenchmarkPushBack(b *testing.B) {
	b.Run("Deque", func(b *testing.B) {
		dq := NewDeque[int]()
		for i := 0; i < b.N; i++ {
			dq.PushBack(i)
		}
	})
	b.Run("list.List", func(b *testing.B) {
		lst := list.New()
		for i := 0; i < b.N; i++ {
			lst.PushBack(i)
		}
	})
}

func BenchmarkPushFront(b *testing.B) {
	b.Run("Deque", func(b *testing.B) {
		dq := NewDeque[int]()
		for i := 0; i < b.N; i++ {
			dq.PushFront(i)
		}
	})
	b.Run("list.List", func(b *testing.B) {
		lst := list.New()
		for i := 0; i < b.N; i++ {
			lst.PushFront(i)
		}
	})
}

func BenchmarkRandom(b *testing.B) {
	const nn = 10000
	a := make([]int, nn)
	for i := 0; i < nn; i++ {
		a[i] = rand.Int()
	}

	b.Run("Deque", func(b *testing.B) {
		dq := NewDeque[int]()
		for i := 0; i < b.N; i++ {
			switch a[i%nn] % 4 {
			case 0:
				dq.PushBack(i)
			case 1:
				dq.PushFront(i)
			case 2:
				dq.TryPopBack()
			case 3:
				dq.TryPopFront()
			}
		}
	})
	b.Run("list.List", func(b *testing.B) {
		lst := list.New()
		for i := 0; i < b.N; i++ {
			switch a[i%nn] % 4 {
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
