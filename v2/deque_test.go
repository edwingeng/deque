package deque

import (
	"fmt"
	"math/rand"
	"reflect"
	"sort"
	"testing"
	"time"
	"unsafe"
)

type invariantParams struct {
	skipChunkMerge bool
}

type invariantOption func(params *invariantParams)

func skipChunkMerge() invariantOption {
	return func(params *invariantParams) {
		params.skipChunkMerge = true
	}
}

//gocyclo:ignore
func invariant(t *testing.T, dq *Deque[int], opts ...invariantOption) {
	t.Helper()
	var params invariantParams
	for _, opt := range opts {
		opt(&params)
	}

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
		x1 := (*reflect.SliceHeader)(unsafe.Pointer(&dq.chunks))
		x2 := (*reflect.SliceHeader)(unsafe.Pointer(&dq.chunkPitch))
		beyond := x2.Data + uintptr(x2.Cap)*unsafe.Sizeof(*new(int))
		if x1.Data < x2.Data || x1.Data >= beyond {
			t.Fatal(`x1.Data < x2.Data || x1.Data >= beyond`)
		}
	}

	if t.Name() != "TestDeque_shrinkEnd" && t.Name() != "TestDeque_shrinkStart" {
		for i := 0; i < len(dq.chunkPitch); i++ {
			if i < dq.sFree {
				if dq.chunkPitch[i] != nil {
					t.Fatal(`dq.chunkPitch[i] != nil [1]`)
				}
			} else if i >= len(dq.chunkPitch)-dq.eFree {
				if dq.chunkPitch[i] != nil {
					t.Fatal(`dq.chunkPitch[i] != nil [2]`)
				}
			} else {
				if dq.chunkPitch[i] == nil {
					t.Fatal(`dq.chunkPitch[i] == nil`)
				}
			}
		}
	}

	var count int
	for i, c := range dq.chunks {
		count += c.e - c.s
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
		if c.e < c.s {
			t.Fatal(`c.e < c.s`)
		}

		if !params.skipChunkMerge {
			if i+1 < len(dq.chunks) {
				if (c.e-c.s)+(dq.chunkSize>>2) <= dq.chunks[i+1].s {
					t.Fatal(`(c.e-c.s)+(dq.chunkSize>>2) <= dq.chunks[i+1].s`)
				}
				if (dq.chunks[i+1].e-dq.chunks[i+1].s)+(dq.chunkSize>>2) <= dq.chunkSize-c.e {
					t.Fatal(`(dq.chunks[i+1].e-dq.chunks[i+1].s)+(dq.chunkSize>>2) <= dq.chunkSize-c.e`)
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

	if NewDeque[int64](WithChunkSize(-100)).chunkSize != 128 {
		t.Fatal(`NewDeque[int64](WithChunkSize(-100)).chunkSize != 128`)
	}
	if NewDeque[int64](WithChunkSize(0)).chunkSize != 128 {
		t.Fatal(`NewDeque[int64](WithChunkSize(0)).chunkSize != 128`)
	}
	if NewDeque[int64](WithChunkSize(20)).chunkSize != 20 {
		t.Fatal(`NewDeque[int64](WithChunkSize(20)).chunkSize != 20`)
	}
}

//gocyclo:ignore
func TestChunk(t *testing.T) {
	dq1 := NewDeque[int]()
	total1 := dq1.chunkSize * 3
	for i := 0; i < total1; i++ {
		dq1.PushBack(i)
		invariant(t, dq1)
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
	invariant(t, dq1)

	dq2 := NewDeque[int]()
	total2 := dq2.chunkSize * 3
	for i := 0; i < total2; i++ {
		dq2.PushFront(i)
		invariant(t, dq2)
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
	invariant(t, dq1)
}

func TestDeque_realloc(t *testing.T) {
	dq1 := NewDeque[int]()
	dq1.PushBack(1)
	for i := 0; i < 3; i++ {
		count := dq1.chunkSize * (dq1.eFree + 1)
		for j := 0; j < count; j++ {
			dq1.PushBack(j)
		}
		pitchLen := defaultPitchSize
		for j := 0; j < i+1; j++ {
			pitchLen *= 2
		}
		if len(dq1.chunkPitch) != pitchLen {
			t.Fatal(`len(dq1.chunkPitch) != pitchLen`)
		}
		if dq1.sFree != (pitchLen-(len(dq1.chunks)-1))/2 {
			t.Fatal(`dq1.sFree != (pitchLen-(len(dq1.chunks)-1))/2`)
		}
	}

	dq2 := NewDeque[int]()
	dq2.PushFront(1)
	for i := 0; i < 3; i++ {
		count := dq2.chunkSize * (dq2.sFree + 1)
		for j := 0; j < count; j++ {
			dq2.PushFront(j)
		}
		pitchLen := defaultPitchSize
		for j := 0; j < i+1; j++ {
			pitchLen *= 2
		}
		if len(dq2.chunkPitch) != pitchLen {
			t.Fatal(`len(dq2.chunkPitch) != pitchLen`)
		}
		if dq2.sFree != (pitchLen-(len(dq2.chunks)-1))/2-1 {
			t.Fatal(`dq2.sFree != (pitchLen-(len(dq2.chunks)-1))/2-1`)
		}
	}

	dq3 := NewDeque[int]()
	for i := 0; i < dq3.chunkSize*3; i++ {
		dq3.PushBack(i)
	}
	if dq3.sFree != 32 {
		t.Fatal(`dq3.sFree != 32`)
	}
	if dq3.eFree != 29 {
		t.Fatal(`dq3.eFree != 29`)
	}
	dq3.realloc()
	if dq3.sFree != 31 {
		t.Fatal(`dq3.sFree != 31`)
	}
	if dq3.eFree != 30 {
		t.Fatal(`dq3.eFree != 30`)
	}
}

func TestDeque_expandEnd(t *testing.T) {
	dq := NewDeque[int]()
	dq.sFree += 10
	dq.eFree -= 10
	sf := dq.sFree
	ef := dq.eFree
	dq.expandEnd()
	invariant(t, dq)
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
	invariant(t, dq)
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
	dq.count = dq.chunkSize * (len(dq.chunkPitch) - dq.sFree - dq.eFree)
	invariant(t, dq)
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
	dq.count = dq.chunkSize * (len(dq.chunkPitch) - dq.sFree - dq.eFree)
	invariant(t, dq)
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
	dq.count = dq.chunkSize * (len(dq.chunkPitch) - dq.sFree - dq.eFree)
	invariant(t, dq)
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
	dq.count = dq.chunkSize * (len(dq.chunkPitch) - dq.sFree - dq.eFree)
	invariant(t, dq)
	if dq.sFree != 32 {
		t.Fatal("dq.sFree != 32")
	}
	if dq.eFree != 32 {
		t.Fatal("dq.eFree != 32")
	}
}

//gocyclo:ignore
func TestDeque_PushBack(t *testing.T) {
	dq := NewDeque[int]()
	total := dq.chunkSize * 10
	for i := 0; i < total; i++ {
		if dq.Len() != i {
			t.Fatal("dq.Len() != i")
		}
		dq.PushBack(i)
	}
	for i := 0; i < total; i++ {
		if v := dq.PopFront(); v != i {
			invariant(t, dq)
			t.Fatal("v != i")
		}
	}
	for i := 0; i < total; i++ {
		dq.PushBack(i)
	}
	for i := 0; i < total; i++ {
		if v, ok := dq.TryPopFront(); !ok || v != i {
			invariant(t, dq)
			t.Fatal("!ok || v != i")
		}
	}

	for i := 0; i < total; i++ {
		dq.PushBack(i)
	}
	for i := 0; i < total; i++ {
		if v := dq.PopBack(); v != total-i-1 {
			invariant(t, dq)
			t.Fatal("v != total-i-1")
		}
	}
	for i := 0; i < total; i++ {
		dq.PushBack(i)
	}
	for i := 0; i < total; i++ {
		if v, ok := dq.TryPopBack(); !ok || v != total-i-1 {
			invariant(t, dq)
			t.Fatal("!ok || v != total-i-1")
		}
	}
}

//gocyclo:ignore
func TestDeque_PushFront(t *testing.T) {
	dq := NewDeque[int]()
	total := dq.chunkSize * 10
	for i := 0; i < total; i++ {
		if dq.Len() != i {
			t.Fatal("dq.Len() != i")
		}
		dq.PushFront(i)
	}
	for i := 0; i < total; i++ {
		if v := dq.PopFront(); v != total-i-1 {
			invariant(t, dq)
			t.Fatal("v != total-i-1")
		}
	}
	for i := 0; i < total; i++ {
		dq.PushFront(i)
	}
	for i := 0; i < total; i++ {
		if v, ok := dq.TryPopFront(); !ok || v != total-i-1 {
			invariant(t, dq)
			t.Fatal("!ok || v != total-i-1")
		}
	}

	for i := 0; i < total; i++ {
		dq.PushFront(i)
	}
	for i := 0; i < total; i++ {
		if v := dq.PopBack(); v != i {
			invariant(t, dq)
			t.Fatal("v != i")
		}
	}
	for i := 0; i < total; i++ {
		dq.PushFront(i)
	}
	for i := 0; i < total; i++ {
		if v, ok := dq.TryPopBack(); !ok || v != i {
			invariant(t, dq)
			t.Fatal("!ok || v != i")
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
		invariant(t, dq1)
	}

	dq2 := NewDeque[int]()
	for i := 0; i < 500; i++ {
		expected := make([]int, i)
		for j := 0; j < i; j++ {
			dq2.PushBack(j)
			expected[j] = j
		}
		invariant(t, dq2)
		ret := dq2.DequeueMany(0)
		if len(ret) != i {
			t.Fatalf("dq2.DequeueMany(0) should return %d values", i)
		}
		invariant(t, dq2)
		checkBufs(nil, ret, expected, fmt.Sprintf("i: %d", i), t)
	}

	const total = 100
	whole := make([]int, total)
	for i := 0; i < total; i++ {
		whole[i] = i
	}

	dq3 := NewDeque[int](WithChunkSize(16))
	steps := []int{1, 2, 3, 5, 7, 8, 15, 16, 17, 31, 32, 33, 47, 48, 49}
	for i := 0; i < total; i++ {
		if dq3.Len() != 0 {
			t.Fatal(`dq3.Len() != 0`)
		}
		expected := whole[:i]
		for _, step := range steps {
			for j := 0; j < i; j++ {
				dq3.PushBack(j)
			}
			invariant(t, dq3)
			remaining := i
			for remaining > 0 {
				c := step
				if remaining < step {
					c = remaining
				}
				ret := dq3.DequeueMany(step)
				if len(ret) != c {
					t.Fatalf("len(ret) != c. len: %d, c: %d, i: %d, step: %d", len(ret), c, i, step)
				}
				invariant(t, dq3)
				start := i - remaining
				checkBufs(nil, ret, expected[start:start+c], fmt.Sprintf("i: %d, step: %d", i, step), t)
				remaining -= c
			}
			if dq3.DequeueMany(0) != nil {
				t.Fatalf("dq3.DequeueMany(0) != nil")
			}
		}
	}
}

func checkBufs(bufA, bufB, expected []int, suffix string, t *testing.T) {
	t.Helper()
	if len(bufB) != len(expected) {
		t.Fatal(`len(bufB) != len(expected). ` + suffix)
	}
	for i, v := range expected {
		if bufB[i] != v {
			t.Fatal(`bufB[i] != v. ` + suffix)
		}
	}
	if bufA == nil || bufB == nil {
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
		invariant(t, dq1)
	}

	dq2 := NewDeque[int]()
	for i := 0; i < 500; i++ {
		expected := make([]int, i)
		for j := 0; j < i; j++ {
			dq2.PushBack(j)
			expected[j] = j
		}
		invariant(t, dq2)
		size := dq2.chunkSize * 2 / (i%3 + 1)
		bufA := make([]int, size, size)
		bufB := dq2.DequeueManyWithBuffer(0, bufA)
		if len(bufB) != i {
			t.Fatalf("dq2.DequeueManyWithBuffer(0, bufA) should return %d values", i)
		}
		invariant(t, dq2)
		checkBufs(bufA, bufB, expected, fmt.Sprintf("i: %d", i), t)
	}

	const total = 100
	whole := make([]int, total)
	for i := 0; i < total; i++ {
		whole[i] = i
	}

	dq3 := NewDeque[int](WithChunkSize(16))
	steps := []int{1, 2, 3, 5, 7, 8, 15, 16, 17, 31, 32, 33, 47, 48, 49}
	for i := 0; i < total; i++ {
		if dq3.Len() != 0 {
			t.Fatal(`dq3.Len() != 0`)
		}
		expected := whole[:i]
		for _, step := range steps {
			for j := 0; j < i; j++ {
				dq3.PushBack(j)
			}
			invariant(t, dq3)
			remaining := i
			for remaining > 0 {
				c := step
				if remaining < step {
					c = remaining
				}
				size := dq2.chunkSize * 2 / (i%3 + 1)
				bufA := make([]int, size, size)
				bufB := dq3.DequeueManyWithBuffer(step, bufA)
				if len(bufB) != c {
					t.Fatalf("len(bufB) != c. len: %d, c: %d, i: %d, step: %d", len(bufB), c, i, step)
				}
				invariant(t, dq3)
				start := i - remaining
				str := fmt.Sprintf("len: %d, c: %d, i: %d, j: %d", len(bufB), c, i, step)
				checkBufs(bufA, bufB, expected[start:start+c], str, t)
				remaining -= c
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

	var n2 int
	dq.Range(func(i int, v int) bool {
		n2++
		return false
	})
	if n2 != 1 {
		t.Fatal(`n2 != 1`)
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

func TestDeque_Dump(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	var a []int
	dq := NewDeque[int](WithChunkSize(16))
	if dq.Dump() != nil {
		t.Fatal(`dq.Dump() != nil`)
	}
	dq.PushBack(1)
	dq.PopFront()
	if dq.Dump() != nil {
		t.Fatal(`dq.Dump() != nil`)
	}

	for i := 0; i < 5000; i++ {
		invariant(t, dq)
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

func TestDeque_Swap(t *testing.T) {
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
			idx1 := rand.Intn(n)
			if dq.Peek(idx1) != a[idx1] {
				t.Fatal("dq.Peek(idx1) != a[idx1]")
			}
			idx2 := rand.Intn(n)
			if dq.Peek(idx2) != a[idx2] {
				t.Fatal("dq.Peek(idx2) != a[idx2]")
			}

			a[idx1], a[idx2] = a[idx2], a[idx1]
			dq.Swap(idx1, idx2)

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
		}

		for _, idx := range []int{-1, dq.Len()} {
			func() {
				defer func() {
					_ = recover()
				}()
				dq.Swap(idx, 0)
				t.Fatal("Swap should panic")
			}()
			func() {
				defer func() {
					_ = recover()
				}()
				dq.Swap(0, idx)
				t.Fatal("Swap should panic")
			}()
		}
	}
}

func checkValues(t *testing.T, dq *Deque[int], expected ...int) {
	t.Helper()
	a := dq.Dump()
	for i, v := range a {
		if v != expected[i] {
			t.Fatal(`v != expected[i]`)
		}
	}
}

func TestDeque_Insert(t *testing.T) {
	dq := NewDeque[int]()
	dq.Insert(0, 100)
	invariant(t, dq)
	if v, ok := dq.Back(); !ok || v != 100 {
		t.Fatal(`v, ok := dq.Back(); !ok || v != 100`)
	}
	if v, ok := dq.Front(); !ok || v != 100 {
		t.Fatal(`v, ok := dq.Front(); !ok || v != 100`)
	}

	dq.Insert(-1, 99)
	invariant(t, dq)
	checkValues(t, dq, 99, 100)

	dq.Insert(0, 98)
	invariant(t, dq)
	checkValues(t, dq, 98, 99, 100)

	dq.Insert(3, 101)
	invariant(t, dq)
	checkValues(t, dq, 98, 99, 100, 101)

	dq.Insert(3, 102)
	invariant(t, dq)
	checkValues(t, dq, 98, 99, 100, 102, 101)

	expected := dq.Dump()
	for i := 0; i < dq.chunkSize; i++ {
		v := -1000 + i
		dq.Insert(0, v)
		invariant(t, dq)
		expected = append([]int{v}, expected...)
		checkValues(t, dq, expected...)
	}
	for i := 0; i < dq.chunkSize; i++ {
		v := 1000 + i
		dq.Insert(9999, v)
		invariant(t, dq)
		expected = append(expected, v)
		checkValues(t, dq, expected...)
	}
	for i := 0; i < dq.chunkSize*3; i++ {
		v := 3000 + i
		where := dq.Len() / 2
		dq.Insert(where, v)
		invariant(t, dq)
		expected = append(expected[:where], append([]int{v}, expected[where:]...)...)
		checkValues(t, dq, expected...)
	}
}

func freeSlots(dq *Deque[int], idx int, sFree, eFree int) {
	c := dq.chunks[idx]
	for i := 0; i < sFree; i++ {
		c.data[i] = 0
	}
	for i := dq.chunkSize - eFree; i < dq.chunkSize; i++ {
		c.data[i] = 0
	}
	dq.count -= sFree - c.s
	dq.count -= eFree - (dq.chunkSize - c.e)
	c.s = sFree
	c.e = dq.chunkSize - eFree
}

//gocyclo:ignore
func TestDeque_insertImpl(t *testing.T) {
	dq := NewDeque[int]()
	for i := 0; i < dq.chunkSize*3; i++ {
		dq.PushBack(i)
	}
	if len(dq.chunks) != 3 {
		t.Fatal(`len(dq.chunks) != 3`)
	}
	freeSlots(dq, 1, 10, 10)
	invariant(t, dq)

	dq.Insert(dq.chunkSize, 888)
	if v := dq.Peek(dq.chunkSize); v != 888 {
		t.Fatal(`v := dq.Peek(dq.chunkSize); v != 888`)
	}
	if dq.chunks[1].s != 9 {
		t.Fatal(`dq.chunks[1].s != 9`)
	}
	if dq.chunks[1].e != dq.chunkSize-10 {
		t.Fatal(dq.chunks[1].e != dq.chunkSize-10)
	}
	invariant(t, dq)

	var idx int
	idx = dq.Len() - dq.chunkSize - 1
	dq.Insert(idx, 888)
	if v := dq.Peek(idx); v != 888 {
		t.Fatal(`v := dq.Peek(idx); v != 888`)
	}
	if dq.chunks[1].s != 9 {
		t.Fatal(`dq.chunks[1].s != 9`)
	}
	if dq.chunks[1].e != dq.chunkSize-9 {
		t.Fatal(`dq.chunks[1].e != dq.chunkSize-9`)
	}
	invariant(t, dq)

	idx = dq.chunkSize - 1
	dq.Insert(idx, 999)
	if v := dq.Peek(idx); v != 999 {
		t.Fatal(`v := dq.Peek(idx); v != 999`)
	}
	if dq.chunks[1].s != 8 {
		t.Fatal(`dq.chunks[1].s != 8`)
	}
	if dq.chunks[1].e != dq.chunkSize-9 {
		t.Fatal(`dq.chunks[1].e != dq.chunkSize-9`)
	}
	invariant(t, dq)

	idx = dq.Len() - dq.chunkSize
	dq.Insert(idx, 999)
	if v := dq.Peek(idx); v != 999 {
		t.Fatal(`v := dq.Peek(idx); v != 999`)
	}
	if dq.chunks[1].s != 8 {
		t.Fatal(`dq.chunks[1].s != 8`)
	}
	if dq.chunks[1].e != dq.chunkSize-8 {
		t.Fatal(`dq.chunks[1].e != dq.chunkSize-8`)
	}
	invariant(t, dq)

	for i := 0; i < dq.chunkSize*2; i++ {
		dq.PushFront(i)
	}
	invariant(t, dq)

	dq.Insert(dq.chunkSize, 777)
	if v := dq.Peek(dq.chunkSize); v != 777 {
		t.Fatal(`v := dq.Peek(dq.chunkSize); v != 777`)
	}
	if dq.chunks[1].s != 0 || dq.chunks[1].e != 1 {
		t.Fatal(`dq.chunks[1].s != 0 || dq.chunks[1].e != 1`)
	}
	invariant(t, dq)

	dq.Insert(10, 666)
	if v := dq.Peek(10); v != 666 {
		t.Fatal(`v := dq.Peek(10); v != 666`)
	}
	if dq.chunks[0].s != 0 || dq.chunks[0].e != 10 {
		t.Fatal(`dq.chunks[0].s != 0 || dq.chunks[0].e != 10`)
	}
	if dq.chunks[1].s != 9 {
		t.Fatal(`dq.chunks[1].s != 9`)
	}
	invariant(t, dq)

	for i := 0; i < dq.chunkSize*2; i++ {
		dq.PushBack(i)
	}
	invariant(t, dq)

	idx = dq.Len() - 10
	dq.Insert(idx, 666)
	if v := dq.Peek(idx); v != 666 {
		t.Fatal(`v := dq.Peek(idx); v != 666`)
	}
	if dq.chunks[len(dq.chunks)-1].s != dq.chunkSize-10 {
		t.Fatal(`dq.chunks[len(dq.chunks)-1].s != dq.chunkSize-10`)
	}
	if dq.chunks[len(dq.chunks)-1].e != dq.chunkSize {
		t.Fatal(`dq.chunks[len(dq.chunks)-1].e != dq.chunkSize`)
	}
	if dq.chunks[len(dq.chunks)-2].e != dq.chunkSize-9 {
		t.Fatal(`dq.chunks[len(dq.chunks)-2].e != dq.chunkSize-9`)
	}
	invariant(t, dq)
}

func TestDeque_insertNewChunk(t *testing.T) {
	dq := NewDeque[int]()
	for i := 0; i < dq.chunkSize*3; i++ {
		dq.PushBack(i)
	}

	dq.chunkPitch = dq.chunkPitch[dq.sFree : dq.sFree+len(dq.chunks)]
	dq.sFree, dq.eFree = 0, 0
	invariant(t, dq)

	dq.insertNewChunk(0, true)
	invariant(t, dq)

	dq.chunkPitch = dq.chunkPitch[dq.sFree : dq.sFree+len(dq.chunks)]
	dq.sFree, dq.eFree = 0, 0

	dq.insertNewChunk(len(dq.chunks)-1, false)
	invariant(t, dq)
}

//gocyclo:ignore
func TestDeque_Remove(t *testing.T) {
	dq := NewDeque[int]()
	var a []int
	for i := 0; i < dq.chunkSize*3; i++ {
		dq.PushBack(i)
		a = append(a, i)
	}

	for _, idx := range []int{-1, dq.Len()} {
		func() {
			defer func() {
				_ = recover()
			}()
			dq.Remove(idx)
			t.Fatal("Remove should panic")
		}()
	}

	for i := 0; i < dq.chunkSize; i++ {
		idx := dq.chunkSize + (dq.chunkSize-i)/2
		dq.Remove(idx)
		a = append(a[:idx], a[idx+1:]...)
	}
	if len(dq.chunks) != 2 {
		t.Fatal(`len(dq.chunks) != 2`)
	}
	invariant(t, dq)
	checkValues(t, dq, a...)

	for i := 0; i < dq.chunkSize; i++ {
		idx := (dq.chunkSize - i) / 2
		dq.Remove(idx)
		a = append(a[:idx], a[idx+1:]...)
	}
	if len(dq.chunks) != 1 {
		t.Fatal(`len(dq.chunks) != 1`)
	}
	invariant(t, dq)
	checkValues(t, dq, a...)

	for i := 0; i < 10; i++ {
		idx := dq.chunkSize - i - 1
		dq.Remove(idx)
		a = append(a[:idx], a[idx+1:]...)
		if dq.chunks[0].s != 0 || dq.chunks[0].e != idx {
			t.Fatal(`dq.chunks[0].s != 0 || dq.chunks[0].e != idx`)
		}
	}
	invariant(t, dq)
	checkValues(t, dq, a...)

	n := dq.Len()
	for i := 0; i < n; i++ {
		dq.Remove(0)
	}
	if len(dq.chunks) != 0 {
		t.Fatal(`len(dq.chunks) != 0`)
	}
	invariant(t, dq)

	var idx int
	for i := 0; i < dq.chunkSize; i++ {
		dq.PushBack(i)
	}
	idx = dq.chunkSize / 2
	dq.Insert(idx, 999)
	dq.Remove(idx)
	if len(dq.chunks) != 2 {
		t.Fatal(`len(dq.chunks) != 2`)
	}
	if dq.chunks[0].e != dq.chunks[1].s {
		t.Fatal(`dq.chunks[0].e != dq.chunks[1].s`)
	}
	for i := 0; i < dq.chunkSize/4-1; i++ {
		dq.Remove(idx)
	}
	if len(dq.chunks) != 2 {
		t.Fatal(`len(dq.chunks) != 2`)
	}
	dq.Remove(idx)
	if len(dq.chunks) != 1 {
		t.Fatal(`len(dq.chunks) != 1`)
	}
	if dq.chunks[0].s != dq.chunkSize/4 {
		t.Fatal(`dq.chunks[0].s != dq.chunkSize/4`)
	}
	invariant(t, dq)

	dq.Clear()
	invariant(t, dq)

	for i := 0; i < dq.chunkSize*3/2; i++ {
		dq.PushBack(i)
	}
	idx = dq.chunkSize / 2
	for i := 0; i < dq.chunkSize/2; i++ {
		dq.Remove(idx)
	}
	if len(dq.chunks) != 2 {
		t.Fatal(`len(dq.chunks) != 2`)
	}
	if dq.chunks[0].e != dq.chunkSize/2 {
		t.Fatal(`dq.chunks[0].e != dq.chunkSize/2`)
	}
	idx = dq.chunkSize / 4
	for i := 0; i < dq.chunkSize/4-1; i++ {
		dq.Remove(idx)
	}
	if len(dq.chunks) != 2 {
		t.Fatal(`len(dq.chunks) != 2`)
	}
	dq.Remove(idx)
	if len(dq.chunks) != 1 {
		t.Fatal(`len(dq.chunks) != 1`)
	}
	if dq.chunks[0].e != dq.chunkSize*3/4 {
		t.Fatal(`dq.chunks[0].e != dq.chunkSize*3/4`)
	}
	invariant(t, dq)

	for i := 0; i < 1000; i++ {
		dq.PushFront(i)
	}
	dq.Clear()
	invariant(t, dq)
	dq.Clear()
	invariant(t, dq)
}

//gocyclo:ignore
func TestDeque_Random(t *testing.T) {
	cfg1 := map[string]int{
		"Clear":       1,
		"Dequeue":     100,
		"DequeueMany": 10,
		"Enqueue":     300,
		"Insert":      300,
		"PopBack":     100,
		"PopFront":    100,
		"PushBack":    300,
		"PushFront":   300,
		"Remove":      100,
		"Replace":     10,
		"Swap":        10,
		"TryDequeue":  100,
		"TryPopBack":  100,
		"TryPopFront": 100,
	}

	var keys []string
	for k := range cfg1 {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	cfg2 := make(map[int]string)
	var sum int
	for _, k := range keys {
		tmp := sum
		sum += cfg1[k]
		for i := tmp; i < sum; i++ {
			cfg2[i] = k
		}
	}

	const total = 1000000
	seed := time.Now().UnixMilli()
	t.Logf("seed: %d", seed)
	r := rand.New(rand.NewSource(seed))
	dq := NewDeque[int](WithChunkSize(8))
	var a []int
	var i int

	checkValues := func() {
		t.Helper()
		if dq.Len() != len(a) {
			t.Fatal(`dq.Len() != len(a)`)
		}
		var idx int
		for _, c := range dq.chunks {
			for j := c.s; j < c.e; j++ {
				if c.data[j] != a[idx] {
					t.Fatalf(`c.data[j] != a[idx]. i: %d`, i)
				}
				idx++
			}
		}
	}

	for i = 1; i <= total; i++ {
		action := cfg2[r.Intn(sum)]
		switch action {
		case "Clear":
			dq.Clear()
			a = nil
		case "Dequeue":
			if len(a) > 0 {
				dq.Dequeue()
				a = a[1:]
			}
		case "DequeueMany":
			count := r.Intn(20) + 1
			dq.DequeueMany(count)
			if len(a) > count {
				a = a[count:]
			} else {
				a = a[:0]
			}
		case "Enqueue":
			dq.Enqueue(i)
			a = append(a, i)
		case "Insert":
			var idx int
			if len(a) > 0 {
				idx = r.Intn(len(a))
			}
			dq.Insert(idx, i)
			a = append(a, 0)
			copy(a[idx+1:], a[idx:len(a)-1])
			a[idx] = i
		case "PopBack":
			if len(a) > 0 {
				dq.PopBack()
				a = a[:len(a)-1]
			}
		case "PopFront":
			if len(a) > 0 {
				dq.PopFront()
				a = a[1:]
			}
		case "PushBack":
			dq.PushBack(i)
			a = append(a, i)
		case "PushFront":
			dq.PushFront(i)
			a = append(a, i)
			copy(a[1:], a[:len(a)-1])
			a[0] = i
		case "Remove":
			if len(a) > 0 {
				idx := r.Intn(len(a))
				dq.Remove(idx)
				a = append(a[:idx], a[idx+1:]...)
			}
		case "Replace":
			if len(a) > 0 {
				idx := r.Intn(len(a))
				dq.Replace(idx, i)
				a[idx] = i
			}
		case "Swap":
			if len(a) > 0 {
				idx1 := r.Intn(len(a))
				idx2 := r.Intn(len(a))
				dq.Swap(idx1, idx2)
				a[idx1], a[idx2] = a[idx2], a[idx1]
			}
		case "TryDequeue":
			dq.TryDequeue()
			if len(a) > 0 {
				a = a[1:]
			}
		case "TryPopBack":
			dq.TryPopBack()
			if len(a) > 0 {
				a = a[:len(a)-1]
			}
		case "TryPopFront":
			dq.TryPopFront()
			if len(a) > 0 {
				a = a[1:]
			}
		default:
			panic("impossible")
		}

		checkValues()
		invariant(t, dq, skipChunkMerge())

		if testing.Verbose() {
			if i%10000 == 0 {
				fmt.Printf("Progress: %d\n", i)
			}
		}
	}
}
