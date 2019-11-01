package deque

import (
	"container/list"
	"fmt"
	"math/rand"
	"reflect"
	"regexp"
	"strings"
	"testing"
)

func benchNameSuffix() string {
	var elem Elem
	t := reflect.TypeOf(elem)
	if t == nil {
		return ""
	}
	rex1 := regexp.MustCompile(`[a-zA-Z0-9_]+\.`)
	str1 := rex1.ReplaceAllString(t.String(), "")
	str2 := strings.ReplaceAll(str1, "interface {}", "interface{}")
	str3 := fmt.Sprintf("<%s>", str2)
	return str3
}

func Benchmark_PushBack(b *testing.B) {
	b.Run("Deque"+benchNameSuffix(), func(b *testing.B) {
		dq := NewDeque()
		for i := 0; i < b.N; i++ {
			dq.PushBack(Elem(i))
		}
	})
	b.Run("list.List", func(b *testing.B) {
		lst := list.New()
		for i := 0; i < b.N; i++ {
			lst.PushBack(i)
		}
	})
}

func Benchmark_PushFront(b *testing.B) {
	b.Run("Deque"+benchNameSuffix(), func(b *testing.B) {
		dq := NewDeque()
		for i := 0; i < b.N; i++ {
			dq.PushFront(Elem(i))
		}
	})
	b.Run("list.List", func(b *testing.B) {
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

	b.Run("Deque"+benchNameSuffix(), func(b *testing.B) {
		dq := NewDeque()
		for i := 0; i < b.N; i++ {
			switch a[i%na] % 4 {
			case 0:
				dq.PushBack(Elem(i))
			case 1:
				dq.PushFront(Elem(i))
			case 2:
				dq.PopBack()
			case 3:
				dq.PopFront()
			}
		}
	})
	b.Run("list.List", func(b *testing.B) {
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
