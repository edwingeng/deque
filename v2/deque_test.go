package deque

import (
	"fmt"
	"math/rand"
	"reflect"
	"testing"
	"time"
	"unsafe"
)

func invariant(dq *Deque[int], t *testing.T) {
	t.Helper()
	if dq.sFree < 0 || dq.sFree > len(dq.chunkPitch) {
		t.Fatal("dq.sFree < 0 || dq.sFree > len(dq.chunkPitch)")
	}
	if dq.eFree < 0 || dq.eFree > len(dq.chunkPitch) {
		t.Fatal("dq.eFree < 0 || dq.eFree > len(dq.chunkPitch)")
	}
	if dq.sFree+dq.eFree+len(dq.chunks) != len(dq.chunkPitch) {
		t.Fatal("dq.sFree+dq.eFree+len(dq.chunks) != len(dq.chunkPitch)")
	}

	if dq.chunks != nil && dq.chunkPitch != nil {
		s1 := (*reflect.SliceHeader)(unsafe.Pointer(&dq.chunks))
		s2 := (*reflect.SliceHeader)(unsafe.Pointer(&dq.chunkPitch))
		if s1.Data < s2.Data || s1.Data >= s2.Data+uintptr(s2.Cap)*unsafe.Sizeof(*new(int)) {
			t.Fatal(`s1.Data < s2.Data || s1.Data >= s2.Data+uintptr(s2.Cap)*unsafe.Sizeof(*new(int))`)
		}
	}

	if n := dq.Len(); n <= dq.chunkSize {
		if len(dq.chunks) > 2 {
			t.Fatal("len(dq.chunks) > 2")
		}
	} else {
		if len(dq.chunks) < n/dq.chunkSize {
			t.Fatal("len(dq.chunks) < n/dq.chunkSize")
		}
		if len(dq.chunks) > n/dq.chunkSize+2 {
			t.Fatal("len(dq.chunks) > n/dq.chunkSize+2")
		}
	}

	for i, c := range dq.chunks {
		if c.s < 0 || c.s > dq.chunkSize {
			t.Fatal("c.s < 0 || c.s > dq.chunkSize")
		}
		if c.e < 0 || c.e > dq.chunkSize {
			t.Fatal("c.e < 0 || c.e > dq.chunkSize")
		}
		if len(c.data) != dq.chunkSize {
			t.Fatal("len(c.data) != dq.chunkSize")
		}
		if c != dq.chunkPitch[dq.sFree+i] {
			t.Fatal("c != dq.chunkPitch[dq.sFree+i]")
		}

		if len(dq.chunks) >= 2 {
			switch i {
			case 0:
				if c.e != dq.chunkSize {
					t.Fatal("c.e != dq.chunkSize")
				}
			case len(dq.chunks) - 1:
				if c.s != 0 {
					t.Fatal("c.s != 0")
				}
			default:
				if c.s != 0 {
					t.Fatal("c.s != 0")
				}
				if c.e != dq.chunkSize {
					t.Fatal("c.e != dq.chunkSize")
				}
			}
		}

		for j, v := range c.data {
			if j < c.s || j >= c.e {
				if v != 0 {
					t.Fatal("any value beyond [s, e) in a chunk should always be the default value of its type")
				}
			}
		}
	}
}

func TestChunkSize(t *testing.T) {
	if NewDeque[int64]().chunkSize != 128 {
		t.Fatal(`NewDeque[int64]().chunkSize != 128`)
	}
	if NewDeque[[2]int64]().chunkSize != 64 {
		t.Fatal(`NewDeque[[2]int64]().chunkSize != 64`)
	}
	if NewDeque[[3]int64]().chunkSize != 42 {
		t.Fatal(`NewDeque[[3]int64]().chunkSize != 42`)
	}
	if NewDeque[[8]int64]().chunkSize != 16 {
		t.Fatal(`NewDeque[[8]int64]().chunkSize != 16`)
	}
	if NewDeque[[9]int64]().chunkSize != 16 {
		t.Fatal(`NewDeque[[9]int64]().chunkSize != 16`)
	}
}

func TestChunk(t *testing.T) {
	dq1 := NewDeque[int]()
	total1 := dq1.chunkSize * 3
	for i := 0; i < total1; i++ {
		dq1.PushBack(i)
		invariant(dq1, t)
		if v, ok := dq1.Back(); !ok || v != i {
			t.Fatal("!ok || v != i")
		}
		if v, ok := dq1.Front(); !ok || v != 0 {
			t.Fatal("!ok || v != 0")
		}
	}

	for i := 0; i < total1; i++ {
		dq1.PopFront()
	}
	if len(dq1.chunks) != 0 {
		t.Fatal(`len(dq1.chunks) != 0`)
	}
	if _, ok := dq1.Back(); ok {
		t.Fatal("ok should be false")
	}
	if _, ok := dq1.Front(); ok {
		t.Fatal("ok should be false")
	}

	dq1.PushBack(9)
	dq1.PopFront()
	if len(dq1.chunks) != 1 {
		t.Fatal(`len(dq1.chunks) != 1`)
	}
	if _, ok := dq1.chunks[0].back(); ok {
		t.Fatal("ok should be false")
	}
	if _, ok := dq1.chunks[0].front(); ok {
		t.Fatal("ok should be false")
	}

	dq2 := NewDeque[int]()
	total2 := dq2.chunkSize * 3
	for i := 0; i < total2; i++ {
		dq2.PushFront(i)
		invariant(dq2, t)
		if v, ok := dq2.Back(); !ok || v != 0 {
			t.Fatal("!ok || v != 0")
		}
		if v, ok := dq2.Front(); !ok || v != i {
			t.Fatal("!ok || v != i")
		}
	}

	for i := 0; i < total2; i++ {
		dq2.PopBack()
	}
	if len(dq2.chunks) != 0 {
		t.Fatal(`len(dq2.chunks) != 0`)
	}
	if _, ok := dq2.Back(); ok {
		t.Fatal("ok should be false")
	}
	if _, ok := dq2.Front(); ok {
		t.Fatal("ok should be false")
	}

	dq2.PushBack(9)
	dq2.PopFront()
	if len(dq2.chunks) != 1 {
		t.Fatal(`len(dq2.chunks) != 1`)
	}
	if _, ok := dq2.chunks[0].back(); ok {
		t.Fatal("ok should be false")
	}
	if _, ok := dq2.chunks[0].front(); ok {
		t.Fatal("ok should be false")
	}
}

func TestDeque_realloc(t *testing.T) {
	dq1 := NewDeque[int]()
	dq1.PushBack(1)
	for i := 0; i < 3; i++ {
		count := dq1.chunkSize * (dq1.eFree + 1)
		for j := 0; j < count; j++ {
			dq1.PushBack(j)
		}
		pitchLen := 64
		for j := 0; j < i+1; j++ {
			pitchLen *= 2
		}
		if len(dq1.chunkPitch) != pitchLen {
			t.Fatal(`len(dq1.chunkPitch) != pitchLen`)
		}
		if dq1.sFree != 32 {
			t.Fatal(`dq1.sFree != 32`)
		}
	}

	dq2 := NewDeque[int]()
	dq2.PushFront(1)
	for i := 0; i < 3; i++ {
		count := dq2.chunkSize * (dq2.sFree + 1)
		for j := 0; j < count; j++ {
			dq2.PushFront(j)
		}
		pitchLen := 64
		for j := 0; j < i+1; j++ {
			pitchLen *= 2
		}
		if len(dq2.chunkPitch) != pitchLen {
			t.Fatal(`len(dq2.chunkPitch) != pitchLen`)
		}
		if dq2.eFree != 32 {
			t.Fatal(`dq2.eFree != 32`)
		}
	}
}

func TestDeque_expandEnd(t *testing.T) {
	dq := NewDeque[int]()
	dq.sFree += 10
	dq.eFree -= 10
	sf := dq.sFree
	ef := dq.eFree
	dq.expandEnd()
	invariant(dq, t)
	if len(dq.chunks) != 1 {
		t.Fatal("len(dq.chunks) != 1")
	}
	if dq.sFree != sf {
		t.Fatal("dq.sFree != sf")
	}
	if dq.eFree != ef-1 {
		t.Fatal("dq.eFree != ef-1")
	}
	if dq.chunks[0] == nil {
		t.Fatal("dq.chunks[0] == nil")
	}
}

func TestDeque_expandStart(t *testing.T) {
	dq := NewDeque[int]()
	dq.sFree += 10
	dq.eFree -= 10
	sf := dq.sFree
	ef := dq.eFree
	dq.expandStart()
	invariant(dq, t)
	if len(dq.chunks) != 1 {
		t.Fatal("len(dq.chunks) != 1")
	}
	if dq.sFree != sf-1 {
		t.Fatal("dq.sFree != sf-1")
	}
	if dq.eFree != ef {
		t.Fatal("dq.eFree != ef")
	}
	if dq.chunks[0] == nil {
		t.Fatal("dq.chunks[0] == nil")
	}
}

func TestDeque_shrinkEnd(t *testing.T) {
	dq := NewDeque[int]()
	for i := 0; i < len(dq.chunkPitch); i++ {
		dq.chunkPitch[i] = &chunk[int]{
			data: make([]int, dq.chunkSize),
			e:    dq.chunkSize,
		}
	}
	dq.eFree -= 10
	sf := dq.sFree
	ef := dq.eFree
	dq.shrinkEnd()
	invariant(dq, t)
	if len(dq.chunks) != 9 {
		t.Fatal("len(dq.chunks) != 9")
	}
	if dq.sFree != sf {
		t.Fatal("dq.sFree != sf")
	}
	if dq.eFree != ef+1 {
		t.Fatal("dq.eFree != ef+1")
	}

	var cnt, idx int
	for i := 0; i < len(dq.chunkPitch); i++ {
		if dq.chunkPitch[i] != nil {
			cnt++
		} else {
			idx = i
		}
	}
	if cnt != len(dq.chunkPitch)-1 {
		t.Fatal("cnt != len(dq.chunkPitch)-1")
	}
	if idx != len(dq.chunkPitch)-ef-1 {
		t.Fatal("idx != len(dq.chunkPitch)-ef-1")
	}

	dq.sFree = 60
	dq.eFree = len(dq.chunkPitch) - dq.sFree - 1
	dq.shrinkEnd()
	invariant(dq, t)
	if dq.sFree != 32 {
		t.Fatal("dq.sFree != 32")
	}
	if dq.eFree != 32 {
		t.Fatal("dq.eFree != 32")
	}
}

func TestDeque_shrinkStart(t *testing.T) {
	dq := NewDeque[int]()
	for i := 0; i < len(dq.chunkPitch); i++ {
		dq.chunkPitch[i] = &chunk[int]{
			data: make([]int, dq.chunkSize),
			e:    dq.chunkSize,
		}
	}
	dq.eFree -= 10
	sf := dq.sFree
	ef := dq.eFree
	dq.shrinkStart()
	invariant(dq, t)
	if len(dq.chunks) != 9 {
		t.Fatal("len(dq.chunks) != 9")
	}
	if dq.sFree != sf+1 {
		t.Fatal("dq.sFree != sf+1")
	}
	if dq.eFree != ef {
		t.Fatal("dq.eFree != ef")
	}

	var cnt, idx int
	for i := 0; i < len(dq.chunkPitch); i++ {
		if dq.chunkPitch[i] != nil {
			cnt++
		} else {
			idx = i
		}
	}
	if cnt != len(dq.chunkPitch)-1 {
		t.Fatal("cnt != len(dq.chunkPitch)-1")
	}
	if idx != sf {
		t.Fatal("idx != sf")
	}

	dq.sFree = 60
	dq.eFree = len(dq.chunkPitch) - dq.sFree - 1
	dq.shrinkEnd()
	invariant(dq, t)
	if dq.sFree != 32 {
		t.Fatal("dq.sFree != 32")
	}
	if dq.eFree != 32 {
		t.Fatal("dq.eFree != 32")
	}
}

func TestDeque_PushBack(t *testing.T) {
	dq := NewDeque[int]()
	const total = 10000
	for i := 0; i < total; i++ {
		if dq.Len() != i {
			t.Fatal("dq.Len() != i")
		}
		dq.PushBack(i)
	}
	for i := 0; i < total; i++ {
		if v := dq.PopFront(); v != i {
			invariant(dq, t)
			t.Fatal("v != i")
		}
	}

	for i := 0; i < total; i++ {
		dq.PushBack(i)
	}
	for i := 0; i < total; i++ {
		if v := dq.PopBack(); v != total-i-1 {
			invariant(dq, t)
			t.Fatal("v != total-i-1")
		}
	}
}

func TestDeque_PushFront(t *testing.T) {
	dq := NewDeque[int]()
	const total = 10000
	for i := 0; i < total; i++ {
		if dq.Len() != i {
			t.Fatal("dq.Len() != i")
		}
		dq.PushFront(i)
	}
	for i := 0; i < total; i++ {
		if v := dq.PopFront(); v != total-i-1 {
			invariant(dq, t)
			t.Fatal("v != total-i-1")
		}
	}

	for i := 0; i < total; i++ {
		dq.PushFront(i)
	}
	for i := 0; i < total; i++ {
		if v := dq.PopBack(); v != i {
			invariant(dq, t)
			t.Fatal("v != i")
		}
	}
}

func TestDeque_PopBack(t *testing.T) {
	dq := NewDeque[int]()
	if _, ok := dq.TryPopBack(); ok {
		t.Fatal("ok should be false")
	}
	func() {
		defer func() {
			_ = recover()
		}()
		dq.PopBack()
		t.Fatal("PopBack should panic")
	}()

	dq.PushBack(1)
	dq.PopFront()
	if _, ok := dq.TryPopBack(); ok {
		t.Fatal("ok should be false")
	}
	func() {
		defer func() {
			_ = recover()
		}()
		dq.PopBack()
		t.Fatal("PopBack should panic")
	}()
}

func TestDeque_PopFront(t *testing.T) {
	dq := NewDeque[int]()
	if _, ok := dq.TryPopFront(); ok {
		t.Fatal("ok should be false")
	}
	func() {
		defer func() {
			_ = recover()
		}()
		dq.PopFront()
		t.Fatal("PopFront should panic")
	}()

	dq.PushBack(1)
	dq.PopFront()
	if _, ok := dq.TryPopFront(); ok {
		t.Fatal("ok should be false")
	}
	func() {
		defer func() {
			_ = recover()
		}()
		dq.PopFront()
		t.Fatal("PopFront should panic")
	}()
}

func TestDeque_DequeueMany(t *testing.T) {
	dq1 := NewDeque[int]()
	for i := -1; i <= 1; i++ {
		if dq1.DequeueMany(i) != nil {
			t.Fatalf("dq1.DequeueMany(%d) should return nil while dq1 is empty", i)
		}
		invariant(dq1, t)
	}

	dq2 := NewDeque[int]()
	for i := 0; i < 1000; i += 5 {
		for j := 0; j < i; j++ {
			dq2.PushBack(j)
		}
		invariant(dq2, t)
		if len(dq2.DequeueMany(0)) != i {
			t.Fatalf("dq2.DequeueMany(0) should return %d values", i)
		}
		invariant(dq2, t)
	}

	for i := 0; i < 2000; i += 5 {
		for j := 5; j < 600; j += 25 {
			dq3 := NewDeque[int]()
			for k := 0; k < i; k++ {
				dq3.PushBack(k)
			}
			invariant(dq3, t)
			left := i
			for left > 0 {
				c := j
				if left < j {
					c = left
				}
				vals := dq3.DequeueMany(j)
				if len(vals) != c {
					t.Fatalf("len(vals) != c. len: %d, c: %d, i: %d, j: %d", len(vals), c, i, j)
				}
				left -= c
				invariant(dq3, t)
			}
			if dq3.DequeueMany(0) != nil {
				t.Fatalf("dq3.DequeueMany(0) != nil")
			}
		}
	}
}

func compareBufs(bufA, bufB []int, suffix string, t *testing.T) {
	t.Helper()
	if bufB == nil {
		return
	}

	ptrA := (*reflect.SliceHeader)(unsafe.Pointer(&bufA)).Data
	ptrB := (*reflect.SliceHeader)(unsafe.Pointer(&bufB)).Data
	if len(bufB) <= cap(bufA) {
		if ptrB != ptrA {
			t.Fatal("ptrB != ptrA. " + suffix)
		}
	} else {
		if ptrB == ptrA {
			t.Fatal("ptrB != ptrA. " + suffix)
		}
	}
}

func TestDeque_DequeueManyWithBuffer(t *testing.T) {
	dq1 := NewDeque[int]()
	for i := -1; i <= 1; i++ {
		if dq1.DequeueManyWithBuffer(i, nil) != nil {
			t.Fatalf("dq1.DequeueManyWithBuffer(%d, nil) should return nil while dq1 is empty", i)
		}
		invariant(dq1, t)
	}

	dq2 := NewDeque[int]()
	for i := 0; i < 1000; i += 5 {
		for j := 0; j < i; j++ {
			dq2.PushBack(j)
		}
		invariant(dq2, t)
		bufA := make([]int, 64, 64)
		bufB := dq2.DequeueManyWithBuffer(0, bufA)
		if len(bufB) != i {
			t.Fatalf("dq2.DequeueManyWithBuffer(0, bufA) should return %d values", i)
		}
		invariant(dq2, t)
		compareBufs(bufA, bufB, fmt.Sprintf("i: %d", i), t)
	}

	for i := 0; i < 2000; i += 5 {
		for j := 5; j < 600; j += 25 {
			dq3 := NewDeque[int]()
			for k := 0; k < i; k++ {
				dq3.PushBack(k)
			}
			invariant(dq3, t)
			left := i
			for left > 0 {
				c := j
				if left < j {
					c = left
				}
				bufA := make([]int, 64, 64)
				bufB := dq3.DequeueManyWithBuffer(j, bufA)
				if len(bufB) != c {
					t.Fatalf("len(bufB) != c. len: %d, c: %d, i: %d, j: %d", len(bufB), c, i, j)
				}
				left -= c
				invariant(dq3, t)
				str := fmt.Sprintf("len: %d, c: %d, i: %d, j: %d", len(bufB), c, i, j)
				compareBufs(bufA, bufB, str, t)
			}
			if dq3.DequeueMany(0) != nil {
				t.Fatalf("dq3.DequeueMany(0) != nil")
			}
		}
	}
}

func TestDeque_Back(t *testing.T) {
	dq := NewDeque[int]()
	if _, ok := dq.Back(); ok {
		t.Fatal("ok should be false")
	}
	for i := 0; i < 1000; i++ {
		dq.PushBack(i)
		if v, ok := dq.Back(); !ok || v != i {
			t.Fatal("!ok || v != i")
		}
	}
}

func TestDeque_Front(t *testing.T) {
	dq := NewDeque[int]()
	if _, ok := dq.Front(); ok {
		t.Fatal("ok should be false")
	}
	for i := 0; i < 1000; i++ {
		dq.PushFront(i)
		if v, ok := dq.Front(); !ok || v != i {
			t.Fatal("!ok || v != i")
		}
	}
}

func TestDeque_Random(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	dq := NewDeque[int]()
	var n int
	for i := 0; i < 100000; i++ {
		invariant(dq, t)
		switch rand.Int() % 4 {
		case 0:
			dq.PushBack(i)
			n++
		case 1:
			dq.PushFront(i)
			n++
		case 2:
			if _, ok := dq.TryPopBack(); ok {
				n--
			}
		case 3:
			if _, ok := dq.TryPopFront(); ok {
				n--
			}
		}
		if dq.Len() != n {
			t.Fatal("dq.Len() != n")
		}
		if n == 0 != dq.IsEmpty() {
			t.Fatal("n == 0 != dq.IsEmpty()")
		}
	}
}

func TestDeque_Dump(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	var a []int
	dq := NewDeque[int]()
	if dq.Dump() != nil {
		t.Fatal(`dq.Dump() != nil`)
	}
	dq.PushBack(1)
	dq.PopFront()
	if dq.Dump() != nil {
		t.Fatal(`dq.Dump() != nil`)
	}

	for i := 0; i < 10000; i++ {
		invariant(dq, t)
		switch rand.Int() % 5 {
		case 0, 1:
			dq.PushBack(i)
			a = append(a, i)
		case 2:
			dq.PushFront(i)
			a = append([]int{i}, a...)
		case 3:
			if _, ok := dq.TryPopBack(); ok {
				a = a[:len(a)-1]
			}
		case 4:
			if _, ok := dq.TryPopFront(); ok {
				a = a[1:]
			}
		}

		b := dq.Dump()
		if len(b) != len(a) {
			t.Fatal("len(b) != len(a)")
		}
		for i, v := range a {
			if b[i] != v {
				t.Fatalf("b[i] != v. i: %d", i)
			}
		}
	}

	for _, idx := range []int{-1, dq.Len()} {
		func() {
			defer func() {
				_ = recover()
			}()
			dq.Peek(idx)
			t.Fatal("Peek should panic")
		}()
	}
}

func TestDeque_Replace(t *testing.T) {
	chunkSize := NewDeque[int]().chunkSize
	for _, n := range []int{1, 100, chunkSize, chunkSize + 1, chunkSize * 2, 1000} {
		a := make([]int, n)
		dq := NewDeque[int]()
		for i := 0; i < n; i++ {
			v := rand.Int()
			a[i] = v
			dq.PushBack(v)
		}
		for i := 0; i < 100; i++ {
			idx := rand.Intn(n)
			if dq.Peek(idx) != a[idx] {
				t.Fatal("dq.Peek(idx) != a[idx]")
			}

			val := rand.Int()
			a[idx] = val
			dq.Replace(idx, val)

			var n1 int
			dq.Range(func(i int, v int) bool {
				if a[i] != v {
					t.Fatal("a[i] != v")
				}
				n1++
				return true
			})
			if n1 != dq.Len() {
				t.Fatal(`n1 != dq.Len()`)
			}

			var n2 int
			dq.Range(func(i int, v int) bool {
				n2++
				return false
			})
			if n2 != 1 {
				t.Fatal(`n2 != 1`)
			}
		}

		for _, idx := range []int{-1, dq.Len()} {
			func() {
				defer func() {
					_ = recover()
				}()
				dq.Replace(idx, 999)
				t.Fatal("Replace should panic")
			}()
		}
	}
}
