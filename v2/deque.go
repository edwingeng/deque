// Package deque implements a highly optimized double-ended queue, which is
// much efficient compared with list.List when adding or removing elements from
// the beginning or the end.
package deque

import (
	"errors"
	"fmt"
	"sync"
	"unsafe"
)

const (
	defaultPitchSize = 64
)

var (
	errEmpty = errors.New("deque is empty")
)

type chunk[T any] struct {
	s    int
	e    int // not included
	data []T
}

func (c *chunk[T]) back() (T, bool) {
	if c.e > c.s {
		return c.data[c.e-1], true
	}
	return *new(T), false
}

func (c *chunk[T]) front() (T, bool) {
	if c.e > c.s {
		return c.data[c.s], true
	}
	return *new(T), false
}

// Deque is a highly optimized double-ended queue.
type Deque[T any] struct {
	chunks []*chunk[T]
	count  int

	chunkPitch []*chunk[T]
	sFree      int
	eFree      int
	chunkSize  int
	chunkPool  sync.Pool
}

func minInt(a, b int) int {
	if a <= b {
		return a
	} else {
		return b
	}
}

func maxInt(a, b int) int {
	if a >= b {
		return a
	} else {
		return b
	}
}

type optionHolder struct {
	chunkSize int
}

// NewDeque creates a new Deque instance.
func NewDeque[T any](opts ...Option) *Deque[T] {
	dq := &Deque[T]{
		chunkPitch: make([]*chunk[T], defaultPitchSize),
		sFree:      32,
		eFree:      32,
	}

	var holder optionHolder
	holder.chunkSize = maxInt(1024/int(unsafe.Sizeof(*new(T))), 16)
	for _, opt := range opts {
		opt(&holder)
	}

	dq.chunkSize = holder.chunkSize
	dq.chunkPool = sync.Pool{
		New: func() any {
			return &chunk[T]{
				data: make([]T, dq.chunkSize, dq.chunkSize),
			}
		},
	}

	return dq
}

func (dq *Deque[T]) balance() {
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

func (dq *Deque[T]) realloc() {
	if len(dq.chunks) < len(dq.chunkPitch)/2 {
		dq.balance()
		return
	}

	newLen := len(dq.chunkPitch) * 2
	newPitch := make([]*chunk[T], newLen, newLen)
	n := len(dq.chunks)
	dq.sFree = (newLen - n) / 2
	dq.eFree = (newLen - n) - dq.sFree
	chunks := newPitch[dq.sFree : dq.sFree+n]
	copy(chunks, dq.chunks)
	dq.chunkPitch = newPitch
	dq.chunks = chunks
}

func (dq *Deque[T]) expandEnd() {
	if f := dq.eFree; f == 0 {
		dq.realloc()
	}
	c := dq.chunkPool.Get().(*chunk[T])
	c.s, c.e = 0, 0
	dq.eFree--
	newEnd := len(dq.chunkPitch) - dq.eFree
	dq.chunkPitch[newEnd-1] = c
	dq.chunks = dq.chunkPitch[dq.sFree:newEnd]
}

func (dq *Deque[T]) expandStart() {
	if f := dq.sFree; f == 0 {
		dq.realloc()
	}
	c := dq.chunkPool.Get().(*chunk[T])
	c.s, c.e = dq.chunkSize, dq.chunkSize
	dq.sFree--
	dq.chunkPitch[dq.sFree] = c
	pitchLen := len(dq.chunkPitch)
	dq.chunks = dq.chunkPitch[dq.sFree : pitchLen-dq.eFree]
}

func (dq *Deque[T]) shrinkEnd() {
	pitchLen := len(dq.chunkPitch)
	dq.eFree++
	newEnd := pitchLen - dq.eFree
	c := dq.chunkPitch[newEnd]
	dq.chunkPitch[newEnd] = nil
	dq.chunkPool.Put(c)
	dq.chunks = dq.chunkPitch[dq.sFree:newEnd]

	if dq.sFree+dq.eFree < pitchLen {
		return
	}
	dq.sFree = pitchLen / 2
	dq.eFree = pitchLen - dq.sFree
}

func (dq *Deque[T]) shrinkStart() {
	c := dq.chunkPitch[dq.sFree]
	dq.chunkPitch[dq.sFree] = nil
	dq.chunkPool.Put(c)
	dq.sFree++
	pitchLen := len(dq.chunkPitch)
	dq.chunks = dq.chunkPitch[dq.sFree : pitchLen-dq.eFree]

	if dq.sFree+dq.eFree < pitchLen {
		return
	}
	dq.sFree = pitchLen / 2
	dq.eFree = pitchLen - dq.sFree
}

// PushBack adds a new value at the back of dq.
func (dq *Deque[T]) PushBack(v T) {
	n := len(dq.chunks)
	if n == 0 || dq.chunks[n-1].e == dq.chunkSize {
		dq.expandEnd()
		n++
	}
	c := dq.chunks[n-1]
	c.data[c.e] = v
	c.e++
	dq.count++
}

// PushFront adds a new value at the front of dq.
func (dq *Deque[T]) PushFront(v T) {
	n := len(dq.chunks)
	if n == 0 || dq.chunks[0].s == 0 {
		dq.expandStart()
	}
	c := dq.chunks[0]
	c.s--
	c.data[c.s] = v
	dq.count++
}

// TryPopBack tries to remove a value from the back of dq and returns the removed value if any.
// The return value ok indicates whether it succeeded.
func (dq *Deque[T]) TryPopBack() (_ T, ok bool) {
	n := len(dq.chunks)
	if n == 0 {
		return *new(T), false
	}
	c := dq.chunks[n-1]
	if c.e == c.s {
		return *new(T), false
	}
	c.e--
	r := c.data[c.e]
	var defVal T
	c.data[c.e] = defVal
	if n > 1 {
		if c.e == c.s {
			dq.shrinkEnd()
		}
	} else {
		if c.e == 0 {
			dq.shrinkEnd()
		}
	}
	dq.count--
	return r, true
}

// PopBack removes a value from the back of dq and returns the removed value.
// It panics if dq is empty.
func (dq *Deque[T]) PopBack() T {
	n := len(dq.chunks)
	if n == 0 {
		panic(errEmpty)
	}
	c := dq.chunks[n-1]
	if c.e == c.s {
		panic(errEmpty)
	}
	c.e--
	r := c.data[c.e]
	var defVal T
	c.data[c.e] = defVal
	if n > 1 {
		if c.e == c.s {
			dq.shrinkEnd()
		}
	} else {
		if c.e == 0 {
			dq.shrinkEnd()
		}
	}
	dq.count--
	return r
}

// TryPopFront tries to remove a value from the front of dq and returns the removed value
// if any. The return value ok indicates whether it succeeded.
func (dq *Deque[T]) TryPopFront() (_ T, ok bool) {
	n := len(dq.chunks)
	if n == 0 {
		return *new(T), false
	}
	c := dq.chunks[0]
	if c.e == c.s {
		return *new(T), false
	}
	r := c.data[c.s]
	var defVal T
	c.data[c.s] = defVal
	c.s++
	if n > 1 {
		if c.s == c.e {
			dq.shrinkStart()
		}
	} else {
		if c.s == dq.chunkSize {
			dq.shrinkStart()
		}
	}
	dq.count--
	return r, true
}

// PopFront removes a value from the front of dq and returns the removed value.
// It panics if dq is empty.
func (dq *Deque[T]) PopFront() T {
	n := len(dq.chunks)
	if n == 0 {
		panic(errEmpty)
	}
	c := dq.chunks[0]
	if c.e == c.s {
		panic(errEmpty)
	}
	r := c.data[c.s]
	var defVal T
	c.data[c.s] = defVal
	c.s++
	if n > 1 {
		if c.s == c.e {
			dq.shrinkStart()
		}
	} else {
		if c.s == dq.chunkSize {
			dq.shrinkStart()
		}
	}
	dq.count--
	return r
}

// DequeueMany removes a number of values from the front of dq and returns
// the removed values or nil if dq is empty.
//
// If max <= 0, DequeueMany removes and returns all the values in dq.
func (dq *Deque[T]) DequeueMany(max int) []T {
	return dq.DequeueManyWithBuffer(max, nil)
}

// DequeueManyWithBuffer is similar to DequeueMany except that it uses
// buf to store the removed values as long as it has enough space.
func (dq *Deque[T]) DequeueManyWithBuffer(max int, buf []T) []T {
	n := dq.count
	if n == 0 {
		return nil
	}
	if max > 0 && n > max {
		n = max
	}
	if n <= cap(buf) {
		buf = buf[:n]
	} else {
		buf = make([]T, n)
	}

	var defVal T
	var i int
	for n > 0 {
		c := dq.chunks[0]
		num := minInt(n, c.e-c.s)
		copy(buf[i:], c.data[c.s:c.s+num])
		i += num
		for j := c.s; j < c.s+num; j++ {
			c.data[j] = defVal
		}
		c.s += num
		if c.s == c.e {
			dq.shrinkStart()
		}
		n -= num
	}

	dq.count -= len(buf)
	return buf
}

// Back returns the last value of dq if any. The return value ok
// indicates whether it succeeded.
func (dq *Deque[T]) Back() (_ T, ok bool) {
	n := len(dq.chunks)
	if n == 0 {
		return *new(T), false
	}
	return dq.chunks[n-1].back()
}

// Front returns the first value of dq if any. The return value ok
// indicates whether it succeeded.
func (dq *Deque[T]) Front() (_ T, ok bool) {
	n := len(dq.chunks)
	if n == 0 {
		return *new(T), false
	}
	return dq.chunks[0].front()
}

// IsEmpty returns whether dq is empty.
func (dq *Deque[T]) IsEmpty() bool {
	return dq.count == 0
}

// Len returns the number of values in dq.
func (dq *Deque[T]) Len() int {
	return dq.count
}

// Enqueue is an alias of PushBack.
func (dq *Deque[T]) Enqueue(v T) {
	dq.PushBack(v)
}

// TryDequeue is an alias of TryPopFront.
func (dq *Deque[T]) TryDequeue() (_ T, ok bool) {
	return dq.TryPopFront()
}

// Dequeue is an alias of PopFront.
func (dq *Deque[T]) Dequeue() T {
	return dq.PopFront()
}

// Dump returns all the values in dq.
func (dq *Deque[T]) Dump() []T {
	n := dq.Len()
	if n == 0 {
		return nil
	}

	vals := make([]T, n)
	var i int
	for _, c := range dq.chunks {
		for j := c.s; j < c.e; j++ {
			vals[i] = c.data[j]
			i++
		}
	}
	return vals
}

// Range iterates all the values in dq. Do NOT add values to dq or remove values from dq during Range.
func (dq *Deque[T]) Range(f func(i int, v T) bool) {
	var i int
	for _, c := range dq.chunks {
		for j := c.s; j < c.e; j++ {
			if !f(i, c.data[j]) {
				return
			}
			i++
		}
	}
}

// Peek returns the value at idx. It panics if idx is out of range.
func (dq *Deque[T]) Peek(idx int) T {
	if idx < 0 || idx >= dq.count {
		panic(fmt.Errorf("out of range: %d", idx))
	}

	i := idx
	for _, c := range dq.chunks {
		n := c.e - c.s
		if i < n {
			return c.data[c.s+i]
		}
		i -= n
	}

	panic("impossible")
}

// Replace replaces the value at idx with v. It panics if idx is out of range.
func (dq *Deque[T]) Replace(idx int, v T) {
	if idx < 0 || idx >= dq.count {
		panic(fmt.Errorf("out of range: %d", idx))
	}

	i := idx
	for _, c := range dq.chunks {
		n := c.e - c.s
		if i < n {
			c.data[c.s+i] = v
			return
		}
		i -= n
	}

	panic("impossible")
}

// Swap exchanges the two values at idx1 and idx2. It panics if idx1 or idx2 is out of range.
func (dq *Deque[T]) Swap(idx1, idx2 int) {
	if idx1 < 0 || idx1 >= dq.count {
		panic(fmt.Errorf("out of range: %d", idx1))
	}
	if idx2 < 0 || idx2 >= dq.count {
		panic(fmt.Errorf("out of range: %d", idx2))
	}

	i1, i2 := idx1, idx2
	var p1, p2 *T
	for _, c := range dq.chunks {
		n := c.e - c.s
		if p1 == nil {
			if i1 < n {
				p1 = &c.data[c.s+i1]
			} else {
				i1 -= n
			}
		}
		if p2 == nil {
			if i2 < n {
				p2 = &c.data[c.s+i2]
			} else {
				i2 -= n
			}
		}
		if p1 != nil && p2 != nil {
			*p1, *p2 = *p2, *p1
			return
		}
	}

	panic("impossible")
}

// Insert inserts a new value v before the value at idx.
//
// Insert may cause the split of a chunk inside dq. Because the size of a chunk is fixed,
// the amount of time taken by Insert has a reasonable limit.
func (dq *Deque[T]) Insert(idx int, v T) {
	if idx <= 0 {
		dq.PushFront(v)
		return
	}
	if idx >= dq.Len() {
		dq.PushBack(v)
		return
	}

	i := idx
	for j, c := range dq.chunks {
		n := c.e - c.s
		if i < n {
			dq.insertImpl(i, v, j, c)
			dq.count++
			return
		}
		i -= n
	}

	panic("impossible")
}

func (dq *Deque[T]) insertImpl(i int, v T, j int, c *chunk[T]) {
	sf0 := c.s > 0
	sf1 := j > 0 && dq.chunks[j-1].e < dq.chunkSize
	ef0 := c.e < dq.chunkSize
	ef1 := j+1 < len(dq.chunks) && dq.chunks[j+1].s > 0
	sfx := sf0 || sf1
	efx := ef0 || ef1
	if sfx {
		if efx {
			if i < c.e-c.s-i {
				dq.insertHead(i, v, j, c, sf0)
				return
			} else {
				dq.insertTail(i, v, j, c, ef0)
				return
			}
		} else {
			dq.insertHead(i, v, j, c, sf0)
			return
		}
	} else if efx {
		dq.insertTail(i, v, j, c, ef0)
		return
	}

	// Now c.s is zero
	if i < c.e-c.s-i {
		if i == 0 {
			cc := dq.insertNewChunk(j, true)
			cc.s, cc.e = 0, 1
			cc.data[0] = v
		} else {
			cc := dq.insertNewChunk(j, true)
			cc.s, cc.e = 0, i
			copy(cc.data, c.data[:i])
			var defVal T
			for k := 0; k < i-1; k++ {
				c.data[k] = defVal
			}
			c.s = i - 1
			c.data[c.s] = v
		}
	} else {
		cc := dq.insertNewChunk(j, false)
		start := i
		cc.s, cc.e = start, c.e
		copy(cc.data[start:], c.data[i:])
		var defVal T
		for k := i + 1; k < c.e; k++ {
			c.data[k] = defVal
		}
		c.data[i] = v
		c.e = i + 1
	}
}

func (dq *Deque[T]) insertHead(i int, v T, j int, c *chunk[T], sf0 bool) {
	if sf0 {
		if i == 0 {
			c.s--
			c.data[c.s] = v
		} else {
			old := c.s
			c.s--
			copy(c.data[c.s:], c.data[old:old+i])
			c.data[c.s+i] = v
		}
	} else {
		// Now c.s is zero
		if i == 0 {
			cc := dq.chunks[j-1]
			cc.data[cc.e] = v
			cc.e++
		} else {
			cc := dq.chunks[j-1]
			cc.data[cc.e] = c.data[0]
			cc.e++
			copy(c.data, c.data[1:i])
			c.data[i-1] = v
		}
	}
}

func (dq *Deque[T]) insertTail(i int, v T, j int, c *chunk[T], ef0 bool) {
	if ef0 {
		copy(c.data[c.s+i+1:], c.data[c.s+i:c.e])
		c.data[c.s+i] = v
		c.e++
	} else {
		cc := dq.chunks[j+1]
		cc.s--
		cc.data[cc.s] = c.data[c.e-1]
		copy(c.data[c.s+i+1:], c.data[c.s+i:c.e-1])
		c.data[c.s+i] = v
	}
}

func (dq *Deque[T]) insertNewChunk(j int, before bool) *chunk[T] {
	cc := dq.chunkPool.Get().(*chunk[T])
	if before {
		if f := dq.sFree; f == 0 {
			dq.realloc()
		}
		dq.sFree--
		pitchLen := len(dq.chunkPitch)
		a := dq.chunkPitch[dq.sFree : pitchLen-dq.eFree]
		copy(a, dq.chunks[:j])
		a[j] = cc
		dq.chunks = a
	} else {
		if f := dq.eFree; f == 0 {
			dq.realloc()
		}
		dq.eFree--
		pitchLen := len(dq.chunkPitch)
		a := dq.chunkPitch[dq.sFree : pitchLen-dq.eFree]
		copy(a[j+2:], dq.chunks[j+1:])
		a[j+1] = cc
		dq.chunks = a
	}
	return cc
}

// Remove removes the value at idx. It panics if idx is out of range.
func (dq *Deque[T]) Remove(idx int) {
	if idx < 0 || idx >= dq.count {
		panic(fmt.Errorf("out of range: %d", idx))
	}

	i := idx
	for j, c := range dq.chunks {
		n := c.e - c.s
		if i < n {
			dq.removeElement(i, j, c)
			dq.count--
			return
		}
		i -= n
	}

	panic("impossible")
}

func (dq *Deque[T]) removeElement(i, j int, c *chunk[T]) {
	n := c.e - c.s
	if n == 1 {
		c.data[c.s+i] = *new(T)
		dq.removeChunk(j, c)
		return
	}

	if i < n-i-1 {
		copy(c.data[c.s+1:], c.data[c.s:c.s+i])
		c.data[c.s] = *new(T)
		c.s++
		dq.mergeChunks(j - 1)
	} else {
		copy(c.data[c.s+i:], c.data[c.s+i+1:c.e])
		c.data[c.e-1] = *new(T)
		c.e--
		dq.mergeChunks(j)
	}
}

func (dq *Deque[T]) removeChunk(j int, c *chunk[T]) {
	dq.chunkPool.Put(c)
	if j < len(dq.chunks)-j-1 {
		copy(dq.chunks[1:], dq.chunks[:j])
		dq.chunks[0] = nil
		dq.chunks = dq.chunks[1:]
		dq.sFree++
	} else {
		copy(dq.chunks[j:], dq.chunks[j+1:])
		newLen := len(dq.chunks) - 1
		dq.chunks[newLen] = nil
		dq.chunks = dq.chunks[:newLen]
		dq.eFree++
	}
}

func (dq *Deque[T]) mergeChunks(j int) {
	if j < 0 {
		return
	}
	if j+1 >= len(dq.chunks) {
		return
	}

	c0 := dq.chunks[j]
	c1 := dq.chunks[j+1]
	n0 := c0.e - c0.s
	x0 := n0 + (dq.chunkSize >> 2)
	if x0 <= c1.s {
		c1.s -= n0
		copy(c1.data[c1.s:], c0.data[c0.s:c0.e])
		var defVal T
		for i := c0.s; i < c0.e; i++ {
			c0.data[i] = defVal
		}
		dq.removeChunk(j, c0)
		return
	}

	n1 := c1.e - c1.s
	x1 := n1 + (dq.chunkSize >> 2)
	if x1 <= dq.chunkSize-c0.e {
		copy(c0.data[c0.e:], c1.data[c1.s:c1.e])
		c0.e += n1
		var defVal T
		for i := c1.s; i < c1.e; i++ {
			c1.data[i] = defVal
		}
		dq.removeChunk(j+1, c1)
		return
	}
}

// Clear removes all the values from dq.
func (dq *Deque[T]) Clear() {
	var defVal T
	for _, c := range dq.chunks {
		for j := c.s; j < c.e; j++ {
			c.data[j] = defVal
		}
	}
	for i, c := range dq.chunks {
		dq.chunkPool.Put(c)
		dq.chunks[i] = nil
	}

	dq.chunks = nil
	dq.count = 0

	dq.sFree = len(dq.chunkPitch) / 2
	dq.eFree = len(dq.chunkPitch) - dq.sFree
}

// Option represents the option of Deque.
type Option func(*optionHolder)

// WithChunkSize sets the chunk size of a Deque.
func WithChunkSize(n int) Option {
	return func(holder *optionHolder) {
		if n >= 8 {
			holder.chunkSize = n
		}
	}
}
