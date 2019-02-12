package deque

import (
	"sync/atomic"
)

const chunkSize = 255

type chunk struct {
	data [chunkSize]interface{}
	s    int
	e    int
}

func (c *chunk) back() interface{} {
	if c.e > c.s {
		return c.data[c.e-1]
	}
	return nil
}

func (c *chunk) front() interface{} {
	if c.e > c.s {
		return c.data[c.s]
	}
	return nil
}

type deque struct {
	chunks   []*chunk
	chunkBed []*chunk
	sFree    int
	eFree    int

	chunkPool *chunkPool
}

var (
	sharedChunkPool = newChunkPool(func() interface{} {
		return &chunk{}
	})
)

// NewDeque creates a new Deque.
func NewDeque() Deque {
	dq := &deque{
		chunkBed:  make([]*chunk, 64),
		sFree:     32,
		eFree:     32,
		chunkPool: sharedChunkPool,
	}
	return dq
}

func (dq *deque) realloc() {
	newBedLen := len(dq.chunkBed) * 2
	newBed := make([]*chunk, newBedLen)
	n := len(dq.chunks)
	dq.sFree = newBedLen/2 - n/2
	dq.eFree = newBedLen - dq.sFree - n
	newChunks := newBed[dq.sFree : dq.sFree+n]
	for i := 0; i < n; i++ {
		newChunks[i] = dq.chunks[i]
	}
	dq.chunkBed = newBed
	dq.chunks = newChunks
}

func (dq *deque) expandEnd() {
	if dq.eFree == 0 {
		dq.realloc()
	}
	c := dq.chunkPool.Get().(*chunk)
	c.s, c.e = 0, 0
	dq.eFree--
	newEnd := len(dq.chunkBed) - dq.eFree
	dq.chunkBed[newEnd-1] = c
	dq.chunks = dq.chunkBed[dq.sFree:newEnd]
}

func (dq *deque) expandStart() {
	if dq.sFree == 0 {
		dq.realloc()
	}
	c := dq.chunkPool.Get().(*chunk)
	c.s, c.e = chunkSize, chunkSize
	dq.sFree--
	dq.chunkBed[dq.sFree] = c
	newEnd := len(dq.chunkBed) - dq.eFree
	dq.chunks = dq.chunkBed[dq.sFree:newEnd]
}

func (dq *deque) shrinkEnd() {
	n := len(dq.chunkBed)
	if dq.sFree+dq.eFree == n {
		return
	}
	newEnd := n - dq.eFree - 1
	c := dq.chunkBed[newEnd]
	dq.chunkBed[newEnd] = nil
	dq.chunkPool.Put(c)
	dq.eFree++
	dq.chunks = dq.chunkBed[dq.sFree:newEnd]
	if dq.sFree+dq.eFree == n {
		dq.sFree = n / 2
		dq.eFree = n - dq.sFree
		return
	}
}

func (dq *deque) shrinkStart() {
	n := len(dq.chunkBed)
	if dq.sFree+dq.eFree == n {
		return
	}
	c := dq.chunkBed[dq.sFree]
	dq.chunkBed[dq.sFree] = nil
	dq.chunkPool.Put(c)
	dq.sFree++
	newEnd := len(dq.chunkBed) - dq.eFree
	dq.chunks = dq.chunkBed[dq.sFree:newEnd]
	if dq.sFree+dq.eFree == n {
		dq.sFree = n / 2
		dq.eFree = n - dq.sFree
		return
	}
}

func (dq *deque) PushBack(v interface{}) {
	var c *chunk
	n := len(dq.chunks)
	if n == 0 {
		dq.expandEnd()
		c = dq.chunks[n]
	} else {
		c = dq.chunks[n-1]
		if c.e == chunkSize {
			dq.expandEnd()
			c = dq.chunks[n]
		}
	}
	c.data[c.e] = v
	c.e++
}

func (dq *deque) PushFront(v interface{}) {
	var c *chunk
	n := len(dq.chunks)
	if n == 0 {
		dq.expandStart()
		c = dq.chunks[0]
	} else {
		c = dq.chunks[0]
		if c.s == 0 {
			dq.expandStart()
			c = dq.chunks[0]
		}
	}
	c.s--
	c.data[c.s] = v
}

func (dq *deque) PopBack() interface{} {
	n := len(dq.chunks)
	if n == 0 {
		return nil
	}
	c := dq.chunks[n-1]
	if c.e == c.s {
		return nil
	}
	c.e--
	r := c.data[c.e]
	c.data[c.e] = nil
	if c.e == 0 {
		dq.shrinkEnd()
	}
	return r
}

func (dq *deque) PopManyBack(max int) []interface{} {
	n := dq.Len()
	if n == 0 {
		return nil
	}
	if max > 0 && n > max {
		n = max
	}
	vals := make([]interface{}, n)
	x := len(dq.chunks) - 1
	for i := 0; i < n; i++ {
		c := dq.chunks[x]
		c.e--
		vals[i] = c.data[c.e]
		c.data[c.e] = nil
		if c.e == 0 {
			dq.shrinkEnd()
			x--
		}
	}
	return vals
}

func (dq *deque) PopFront() interface{} {
	n := len(dq.chunks)
	if n == 0 {
		return nil
	}
	c := dq.chunks[0]
	if c.e == c.s {
		return nil
	}
	r := c.data[c.s]
	c.data[c.s] = nil
	c.s++
	if c.s == chunkSize {
		dq.shrinkStart()
	}
	return r
}

func (dq *deque) PopManyFront(max int) []interface{} {
	n := dq.Len()
	if n == 0 {
		return nil
	}
	if max > 0 && n > max {
		n = max
	}
	vals := make([]interface{}, n)
	for i := 0; i < n; i++ {
		c := dq.chunks[0]
		vals[i] = c.data[c.s]
		c.data[c.s] = nil
		c.s++
		if c.s == chunkSize {
			dq.shrinkStart()
		}
	}
	return vals
}

func (dq *deque) Back() interface{} {
	n := len(dq.chunks)
	if n == 0 {
		return nil
	}
	return dq.chunks[n-1].back()
}

func (dq *deque) Front() interface{} {
	n := len(dq.chunks)
	if n == 0 {
		return nil
	}
	return dq.chunks[0].front()
}

func (dq *deque) Empty() bool {
	n := len(dq.chunks)
	return n == 0 || n == 1 && dq.chunks[0].e == dq.chunks[0].s
}

func (dq *deque) Len() int {
	n := len(dq.chunks)
	switch n {
	case 0:
		return 0
	case 1:
		return dq.chunks[0].e - dq.chunks[0].s
	default:
		return chunkSize - dq.chunks[0].s + dq.chunks[n-1].e + (n-2)*chunkSize
	}
}

func (dq *deque) Enqueue(v interface{}) {
	dq.PushBack(v)
}

func (dq *deque) Dequeue() interface{} {
	return dq.PopFront()
}

// NumChunksAllocated returns the number of chunks allocated by now.
func NumChunksAllocated() int64 {
	return atomic.LoadInt64(&sharedChunkPool.numChunksAllocated)
}
