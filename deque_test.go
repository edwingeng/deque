package deque

import (
	"math"
	"math/rand"
	"testing"
	"time"
)

func always(dq Deque, t *testing.T) {
	t.Helper()
	dq1 := dq.(*deque)
	if dq1.sFree < 0 || dq1.sFree > len(dq1.chunkBed) {
		t.Fatal("dq1.sFree < 0 || dq1.sFree > len(dq1.chunkBed)")
	}
	if dq1.eFree < 0 || dq1.eFree > len(dq1.chunkBed) {
		t.Fatal("dq1.eFree < 0 || dq1.eFree > len(dq1.chunkBed)")
	}
	if dq1.sFree+dq1.eFree+len(dq1.chunks) != len(dq1.chunkBed) {
		t.Fatal("dq1.sFree+dq1.eFree+len(dq1.chunks) != len(dq1.chunkBed)")
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
			if len(dq.chunkBed) != n {
				t.Fatal("len(dq.chunkBed) != n")
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
	for i := 0; i < len(dq.chunkBed); i++ {
		dq.chunkBed[i] = &chunk{}
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
	for i := 0; i < len(dq.chunkBed); i++ {
		if dq.chunkBed[i] != nil {
			b++
		} else {
			idx = i
		}
	}
	if b != len(dq.chunkBed)-1 {
		t.Fatal("b != len(dq.chunkBed)-1")
	}
	if idx != len(dq.chunkBed)-ef-1 {
		t.Fatal("idx != len(dq.chunkBed)-ef-1")
	}

	dq.sFree = 60
	dq.eFree = len(dq.chunkBed) - dq.sFree - 1
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
	for i := 0; i < len(dq.chunkBed); i++ {
		dq.chunkBed[i] = &chunk{}
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
	for i := 0; i < len(dq.chunkBed); i++ {
		if dq.chunkBed[i] != nil {
			b++
		} else {
			idx = i
		}
	}
	if b != len(dq.chunkBed)-1 {
		t.Fatal("b != len(dq.chunkBed)-1")
	}
	if idx != sf {
		t.Fatal("idx != sf")
	}

	dq.sFree = 60
	dq.eFree = len(dq.chunkBed) - dq.sFree - 1
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
