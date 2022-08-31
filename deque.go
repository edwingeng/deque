package deque

import (
	"fmt"
	"sync/atomic"
)

const chunkSize = 254

var elemDefValue Elem

type chunk struct {
	s    int
	e    int
	data [chunkSize]Elem
}

func (c *chunk) back() Elem {
	if c.e > c.s {
		return c.data[c.e-1]
	}
	return elemDefValue
}

func (c *chunk) front() Elem {
	if c.e > c.s {
		return c.data[c.s]
	}
	return elemDefValue
}

type deque struct {
	chunks []*chunk

	chunkPitch []*chunk
	sFree      int
	eFree      int
}

var (
	sharedChunkPool = newChunkPool(func() interface{} {
		return &chunk{}
	})
)

// NewDeque creates a new Deque instance.
func NewDeque() Deque {
	dq := &deque{
		chunkPitch: make([]*chunk, 64),
		sFree:      32,
		eFree:      32,
	}
	return dq
}

func (dq *deque) balance() {
	var pitchLen = len(dq.chunkPitch)
	n := len(dq.chunks)
	dq.sFree = pitchLen/2 - n/2
	dq.eFree = pitchLen - dq.sFree - n
	newChunks := dq.chunkPitch[dq.sFree : dq.sFree+n]
	copy(newChunks, dq.chunks)
	dq.chunks = newChunks
	for i := 0; i < dq.sFree; i++ {
		dq.chunkPitch[i] = nil
	}
	for i := pitchLen - dq.eFree; i < pitchLen; i++ {
		dq.chunkPitch[i] = nil
	}
}

func (dq *deque) realloc() {
	if len(dq.chunks) < len(dq.chunkPitch)/2 {
		dq.balance()
		return
	}

	newLen := len(dq.chunkPitch) * 2
	newPitch := make([]*chunk, newLen)
	n := len(dq.chunks)
	dq.sFree = newLen/2 - n/2
	dq.eFree = newLen - dq.sFree - n
	newChunks := newPitch[dq.sFree : dq.sFree+n]
	for i := 0; i < n; i++ {
		newChunks[i] = dq.chunks[i]
	}
	dq.chunkPitch = newPitch
	dq.chunks = newChunks
}

func (dq *deque) expandEnd() {
	if f := dq.eFree; f == 0 {
		dq.realloc()
	}
	c := sharedChunkPool.Get().(*chunk)
	c.s, c.e = 0, 0
	dq.eFree--
	newEnd := len(dq.chunkPitch) - dq.eFree
	dq.chunkPitch[newEnd-1] = c
	dq.chunks = dq.chunkPitch[dq.sFree:newEnd]
}

func (dq *deque) expandStart() {
	if f := dq.sFree; f == 0 {
		dq.realloc()
	}
	c := sharedChunkPool.Get().(*chunk)
	c.s, c.e = chunkSize, chunkSize
	dq.sFree--
	dq.chunkPitch[dq.sFree] = c
	newEnd := len(dq.chunkPitch) - dq.eFree
	dq.chunks = dq.chunkPitch[dq.sFree:newEnd]
}

func (dq *deque) shrinkEnd() {
	n := len(dq.chunkPitch)
	if dq.sFree+dq.eFree == n {
		return
	}
	newEnd := n - dq.eFree - 1
	c := dq.chunkPitch[newEnd]
	dq.chunkPitch[newEnd] = nil
	sharedChunkPool.Put(c)
	dq.eFree++
	dq.chunks = dq.chunkPitch[dq.sFree:newEnd]
	if dq.sFree+dq.eFree < n {
		return
	}
	dq.sFree = n / 2
	dq.eFree = n - dq.sFree
}

func (dq *deque) shrinkStart() {
	n := len(dq.chunkPitch)
	if dq.sFree+dq.eFree == n {
		return
	}
	c := dq.chunkPitch[dq.sFree]
	dq.chunkPitch[dq.sFree] = nil
	sharedChunkPool.Put(c)
	dq.sFree++
	curEnd := len(dq.chunkPitch) - dq.eFree
	dq.chunks = dq.chunkPitch[dq.sFree:curEnd]
	if dq.sFree+dq.eFree < n {
		return
	}
	dq.sFree = n / 2
	dq.eFree = n - dq.sFree
}

func (dq *deque) PushBack(v Elem) {
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

func (dq *deque) PushFront(v Elem) {
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

func (dq *deque) PopBack() Elem {
	n := len(dq.chunks)
	if n == 0 {
		return elemDefValue
	}
	c := dq.chunks[n-1]
	if c.e == c.s {
		return elemDefValue
	}
	c.e--
	r := c.data[c.e]
	c.data[c.e] = elemDefValue
	if c.e == 0 {
		dq.shrinkEnd()
	}
	return r
}

func (dq *deque) PopFront() Elem {
	n := len(dq.chunks)
	if n == 0 {
		return elemDefValue
	}
	c := dq.chunks[0]
	if c.e == c.s {
		return elemDefValue
	}
	r := c.data[c.s]
	c.data[c.s] = elemDefValue
	c.s++
	if c.s == chunkSize {
		dq.shrinkStart()
	}
	return r
}

func (dq *deque) DequeueMany(max int) []Elem {
	return dq.dequeueManyImpl(max, nil)
}

func (dq *deque) dequeueManyImpl(max int, buf []Elem) []Elem {
	n := dq.Len()
	if n == 0 {
		return nil
	}
	if max > 0 && n > max {
		n = max
	}
	if n <= cap(buf) {
		buf = buf[:n]
	} else {
		buf = make([]Elem, n)
	}
	for i := 0; i < n; i++ {
		c := dq.chunks[0]
		buf[i] = c.data[c.s]
		c.data[c.s] = elemDefValue
		c.s++
		if c.s == chunkSize {
			dq.shrinkStart()
		}
	}
	return buf
}

func (dq *deque) DequeueManyWithBuffer(max int, buf []Elem) []Elem {
	return dq.dequeueManyImpl(max, buf)
}

func (dq *deque) Back() Elem {
	n := len(dq.chunks)
	if n == 0 {
		return elemDefValue
	}
	return dq.chunks[n-1].back()
}

func (dq *deque) Front() Elem {
	n := len(dq.chunks)
	if n == 0 {
		return elemDefValue
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

func (dq *deque) Enqueue(v Elem) {
	dq.PushBack(v)
}

func (dq *deque) Dequeue() Elem {
	return dq.PopFront()
}

func (dq *deque) Dump() []Elem {
	n := dq.Len()
	if n == 0 {
		return nil
	}

	vals := make([]Elem, n)
	var idx int
	for _, c := range dq.chunks {
		for i := c.s; i < c.e; i++ {
			vals[idx] = c.data[i]
			idx++
		}
	}
	return vals
}

func (dq *deque) Range(f func(i int, v Elem) bool) {
	n := dq.Len()
	if n == 0 {
		return
	}

	var i int
	for _, c := range dq.chunks {
		for j := c.s; j < c.e; j++ {
			v := c.data[j]
			if !f(i, v) {
				return
			}
			i++
		}
	}
}

func (dq *deque) Peek(idx int) Elem {
	i := idx
	for _, c := range dq.chunks {
		if i < 0 {
			break
		}
		n := c.e - c.s
		if i < n {
			return c.data[c.s+i]
		}
		i -= n
	}
	panic(fmt.Errorf("out of range: %d", idx))
}

func (dq *deque) Replace(idx int, v Elem) {
	i := idx
	for _, c := range dq.chunks {
		if i < 0 {
			break
		}
		n := c.e - c.s
		if i < n {
			c.data[c.s+i] = v
			return
		}
		i -= n
	}
	panic(fmt.Errorf("out of range: %d", idx))
}

// NumChunksAllocated returns the number of chunks allocated by now.
func NumChunksAllocated() int64 {
	return atomic.LoadInt64(&sharedChunkPool.numChunksAllocated)
}
