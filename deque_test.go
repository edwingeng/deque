package deque

import (
	"fmt"
	"math"
	"math/rand"
	"reflect"
	"testing"
	"time"
	"unsafe"
)

func always(dq Deque, t *testing.T) {
	t.Helper()
	dq1 := dq.(*deque)
	if dq1.sFree < 0 || dq1.sFree > len(dq1.ptrPitch) {
		t.Fatal("dq1.sFree < 0 || dq1.sFree > len(dq1.ptrPitch)")
	}
	if dq1.eFree < 0 || dq1.eFree > len(dq1.ptrPitch) {
		t.Fatal("dq1.eFree < 0 || dq1.eFree > len(dq1.ptrPitch)")
	}
	if dq1.sFree+dq1.eFree+len(dq1.chunks) != len(dq1.ptrPitch) {
		t.Fatal("dq1.sFree+dq1.eFree+len(dq1.chunks) != len(dq1.ptrPitch)")
	}
	if n := dq1.Len(); n <= chunkSize {
		if len(dq1.chunks) > 2 {
			t.Fatal("len(dq1.chunks) > 2")
		}
	} else {
		if len(dq1.chunks) < n/chunkSize {
			t.Fatal("len(dq1.chunks) < n/chunkSize")
		}
		if len(dq1.chunks) > n/chunkSize+2 {
			t.Fatal("len(dq1.chunks) > n/chunkSize+2")
		}
	}
	for _, c := range dq1.chunks {
		if c.s < 0 || c.s > chunkSize {
			t.Fatal("c.s < 0 || c.s > chunkSize")
		}
		if c.e < 0 || c.e > chunkSize {
			t.Fatal("c.e < 0 || c.e > chunkSize")
		}
	}
}

func TestChunk(t *testing.T) {
	dq1 := NewDeque()
	for i := 0; i < 5; i++ {
		dq1.PushBack(i)
		always(dq1, t)
		if dq1.Back().(int) != i {
			t.Fatal("dq1.Back().(int) != i")
		}
		if dq1.Front().(int) != 0 {
			t.Fatal("dq1.Front().(int) != 0")
		}
	}

	for dq1.Len() > 0 {
		dq1.PopFront()
	}
	if dq1.(*deque).chunks[0].back() != nil {
		t.Fatal("dq1.(*deque).chunks[0].back() != nil")
	}
	if dq1.(*deque).chunks[0].front() != nil {
		t.Fatal("dq1.(*deque).chunks[0].front() != nil")
	}

	dq2 := NewDeque()
	for i := 0; i < 5; i++ {
		dq2.PushFront(i)
		always(dq2, t)
		if dq2.Back().(int) != 0 {
			t.Fatal("dq2.Back().(int) != 0")
		}
		if dq2.Front().(int) != i {
			t.Fatal("dq2.Front().(int) != i")
		}
	}
}

func TestDeque_realloc(t *testing.T) {
	for pushes := 0; pushes < chunkSize*5; pushes++ {
		dq := NewDeque().(*deque)
		for i := 0; i < pushes; i++ {
			dq.PushBack(i)
		}
		numChunks := pushes / chunkSize
		if pushes%chunkSize != 0 {
			numChunks++
		}
		for i := 0; i < 3; i++ {
			dq.realloc()
			always(dq, t)
			n := 64 * int(math.Pow(2, float64(i+1)))
			if len(dq.ptrPitch) != n {
				t.Fatal("len(dq.ptrPitch) != n")
			}
			if len(dq.chunks) != numChunks {
				t.Fatal("len(dq.chunks) != numChunks")
			}
			if dq.sFree != n/2-numChunks/2 {
				t.Fatal("dq.sFree != n/2-numChunks/2")
			}
			if dq.eFree != n-dq.sFree-numChunks {
				t.Fatal("dq.eFree != n-dq.sFree-numChunks")
			}
		}
	}
}

func TestDeque_expandEnd(t *testing.T) {
	dq := NewDeque().(*deque)
	dq.sFree += 10
	dq.eFree -= 10
	sf := dq.sFree
	ef := dq.eFree
	dq.expandEnd()
	always(dq, t)
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
	dq := NewDeque().(*deque)
	dq.sFree += 10
	dq.eFree -= 10
	sf := dq.sFree
	ef := dq.eFree
	dq.expandStart()
	always(dq, t)
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
	dq := NewDeque().(*deque)
	for i := 0; i < len(dq.ptrPitch); i++ {
		dq.ptrPitch[i] = &chunk{}
	}
	dq.eFree -= 10
	sf := dq.sFree
	ef := dq.eFree
	dq.shrinkEnd()
	always(dq, t)
	if len(dq.chunks) != 9 {
		t.Fatal("len(dq.chunks) != 9")
	}
	if dq.sFree != sf {
		t.Fatal("dq.sFree != sf")
	}
	if dq.eFree != ef+1 {
		t.Fatal("dq.eFree != ef+1")
	}
	b := 0
	idx := 0
	for i := 0; i < len(dq.ptrPitch); i++ {
		if dq.ptrPitch[i] != nil {
			b++
		} else {
			idx = i
		}
	}
	if b != len(dq.ptrPitch)-1 {
		t.Fatal("b != len(dq.ptrPitch)-1")
	}
	if idx != len(dq.ptrPitch)-ef-1 {
		t.Fatal("idx != len(dq.ptrPitch)-ef-1")
	}

	dq.sFree = 60
	dq.eFree = len(dq.ptrPitch) - dq.sFree - 1
	dq.shrinkEnd()
	always(dq, t)
	if dq.sFree != 32 {
		t.Fatal("dq.sFree != 32")
	}
	if dq.eFree != 32 {
		t.Fatal("dq.eFree != 32")
	}
}

func TestDeque_shrinkStart(t *testing.T) {
	dq := NewDeque().(*deque)
	for i := 0; i < len(dq.ptrPitch); i++ {
		dq.ptrPitch[i] = &chunk{}
	}
	dq.eFree -= 10
	sf := dq.sFree
	ef := dq.eFree
	dq.shrinkStart()
	always(dq, t)
	if len(dq.chunks) != 9 {
		t.Fatal("len(dq.chunks) != 9")
	}
	if dq.sFree != sf+1 {
		t.Fatal("dq.sFree != sf+1")
	}
	if dq.eFree != ef {
		t.Fatal("dq.eFree != ef")
	}
	b := 0
	idx := 0
	for i := 0; i < len(dq.ptrPitch); i++ {
		if dq.ptrPitch[i] != nil {
			b++
		} else {
			idx = i
		}
	}
	if b != len(dq.ptrPitch)-1 {
		t.Fatal("b != len(dq.ptrPitch)-1")
	}
	if idx != sf {
		t.Fatal("idx != sf")
	}

	dq.sFree = 60
	dq.eFree = len(dq.ptrPitch) - dq.sFree - 1
	dq.shrinkEnd()
	always(dq, t)
	if dq.sFree != 32 {
		t.Fatal("dq.sFree != 32")
	}
	if dq.eFree != 32 {
		t.Fatal("dq.eFree != 32")
	}
}

func TestDeque_PushBack(t *testing.T) {
	dq := NewDeque()
	const total = 10000
	for i := 0; i < total; i++ {
		if dq.Len() != i {
			t.Fatal("dq.Len() != i")
		}
		dq.PushBack(i)
	}
	for i := 0; i < total; i++ {
		if dq.PopFront().(int) != i {
			always(dq, t)
			t.Fatal("dq.PopFront().(int) != i")
		}
	}

	for i := 0; i < total; i++ {
		dq.PushBack(i)
	}
	for i := 0; i < total; i++ {
		if dq.PopBack().(int) != total-i-1 {
			always(dq, t)
			t.Fatal("dq.PopBack().(int) != total-i-1")
		}
	}
}

func TestDeque_PushFront(t *testing.T) {
	dq := NewDeque()
	const total = 10000
	for i := 0; i < total; i++ {
		if dq.Len() != i {
			t.Fatal("dq.Len() != i")
		}
		dq.PushFront(i)
	}
	for i := 0; i < total; i++ {
		if dq.PopFront().(int) != total-i-1 {
			always(dq, t)
			t.Fatal("dq.PopFront().(int) != total-i-1")
		}
	}

	for i := 0; i < total; i++ {
		dq.PushFront(i)
	}
	for i := 0; i < total; i++ {
		if dq.PopBack().(int) != i {
			always(dq, t)
			t.Fatal("dq.PopBack().(int) != i")
		}
	}
}

func TestDeque_DequeueMany(t *testing.T) {
	dq1 := NewDeque()
	for i := -1; i <= 1; i++ {
		if dq1.DequeueMany(i) != nil {
			t.Fatalf("dq1.DequeueMany(%d) should return nil while dq1 is empty", i)
		}
		always(dq1, t)
	}

	dq2 := NewDeque()
	for i := 0; i < 1000; i += 5 {
		for j := 0; j < i; j++ {
			dq2.PushBack(j)
		}
		always(dq2, t)
		if len(dq2.DequeueMany(0)) != i {
			t.Fatalf("dq2.DequeueMany(0) should return %d values", i)
		}
		always(dq2, t)
	}

	for i := 0; i < 2000; i += 5 {
		for j := 5; j < 600; j += 25 {
			dq3 := NewDeque()
			for k := 0; k < i; k++ {
				dq3.PushBack(k)
			}
			always(dq3, t)
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
				always(dq3, t)
			}
			if dq3.DequeueMany(0) != nil {
				t.Fatalf("dq3.DequeueMany(0) != nil")
			}
		}
	}
}

func compareBufs(bufA, bufB []interface{}, suffix string, t *testing.T) {
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
	dq1 := NewDeque()
	for i := -1; i <= 1; i++ {
		if dq1.DequeueManyWithBuffer(i, nil) != nil {
			t.Fatalf("dq1.DequeueManyWithBuffer(%d, nil) should return nil while dq1 is empty", i)
		}
		always(dq1, t)
	}

	dq2 := NewDeque()
	for i := 0; i < 1000; i += 5 {
		for j := 0; j < i; j++ {
			dq2.PushBack(j)
		}
		always(dq2, t)
		bufA := make([]interface{}, 64, 64)
		bufB := dq2.DequeueManyWithBuffer(0, bufA)
		if len(bufB) != i {
			t.Fatalf("dq2.DequeueManyWithBuffer(0, bufA) should return %d values", i)
		}
		always(dq2, t)
		compareBufs(bufA, bufB, fmt.Sprintf("i: %d", i), t)
	}

	for i := 0; i < 2000; i += 5 {
		for j := 5; j < 600; j += 25 {
			dq3 := NewDeque()
			for k := 0; k < i; k++ {
				dq3.PushBack(k)
			}
			always(dq3, t)
			left := i
			for left > 0 {
				c := j
				if left < j {
					c = left
				}
				bufA := make([]interface{}, 64, 64)
				bufB := dq3.DequeueManyWithBuffer(j, bufA)
				if len(bufB) != c {
					t.Fatalf("len(bufB) != c. len: %d, c: %d, i: %d, j: %d", len(bufB), c, i, j)
				}
				left -= c
				always(dq3, t)
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
	dq := NewDeque()
	if dq.Back() != nil {
		t.Fatal("dq.Back() != nil")
	}
	for i := 0; i < 1000; i++ {
		dq.PushBack(i)
		if dq.Back().(int) != i {
			t.Fatal("dq.Back().(int) != i")
		}
	}
}

func TestDeque_Front(t *testing.T) {
	dq := NewDeque()
	if dq.Front() != nil {
		t.Fatal("dq.Front() != nil")
	}
	for i := 0; i < 1000; i++ {
		dq.PushFront(i)
		if dq.Front().(int) != i {
			t.Fatal("dq.Front().(int) != i")
		}
	}
}

func TestDeque_Random(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	dq := NewDeque()
	var n int
	for i := 0; i < 100000; i++ {
		always(dq, t)
		switch rand.Int() % 4 {
		case 0:
			dq.PushBack(i)
			n++
		case 1:
			dq.PushFront(i)
			n++
		case 2:
			if dq.PopBack() != nil {
				n--
			}
		case 3:
			if dq.PopFront() != nil {
				n--
			}
		}
		if dq.Len() != n {
			t.Fatal("dq.Len() != n")
		}
		if n == 0 != dq.Empty() {
			t.Fatal("n == 0 != dq.Empty()")
		}
	}
}

func TestDeque_Dump(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	var a []Elem
	dq := NewDeque().(*deque)
	for i := 0; i < 10000; i++ {
		always(dq, t)
		switch rand.Int() % 5 {
		case 0, 1:
			dq.PushBack(i)
			a = append(a, i)
		case 2:
			dq.PushFront(i)
			a = append([]Elem{i}, a...)
		case 3:
			if dq.PopBack() != nil {
				a = a[:len(a)-1]
			}
		case 4:
			if dq.PopFront() != nil {
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
}

func TestDeque_Replace(t *testing.T) {
	for _, n := range []int{1, 100, chunkSize, chunkSize + 1, chunkSize * 2, 1000} {
		a := make([]int, n)
		dq := NewDeque()
		for i := 0; i < n; i++ {
			v := rand.Int()
			a[i] = v
			dq.PushBack(v)
		}
		for i := 0; i < 100; i++ {
			idx := rand.Intn(n)
			if dq.Peek(idx).(int) != a[idx] {
				t.Fatal("dq.Peek(idx).(int) != a[idx]")
			}

			val := rand.Int()
			a[idx] = val
			dq.Replace(idx, val)

			dq.Range(func(i int, v Elem) bool {
				if a[i] != v.(int) {
					t.Fatal("a[i] != v.(int)")
				}
				return true
			})
		}
	}
}
