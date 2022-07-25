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

var errEmpty = errors.New("deque is empty")

type chunk[T any] struct {
	s    int
	e    int // not included
	data []T
}

func newChunk[T any](size int) *chunk[T] {
	return &chunk[T]{
		data: make([]T, size, size),
	}
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
	chunks     []*chunk[T]
	chunkPitch []*chunk[T]
	sFree      int
	eFree      int

	chunkSize int
	chunkPool sync.Pool
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
		chunkPitch: make([]*chunk[T], 64),
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
			return newChunk[T](dq.chunkSize)
		},
	}

	return dq
}

func (dq *Deque[T]) realloc() {
	newPL := len(dq.chunkPitch) * 2
	newPitch := make([]*chunk[T], newPL)
	n := len(dq.chunks)
	var sf, ef int
	if dq.sFree >= dq.eFree {
		sf = 32
		ef = newPL - sf - n
	} else {
		ef = 32
		sf = newPL - ef - n
	}
	dq.sFree = sf
	dq.eFree = ef
	chunks := newPitch[sf : sf+n]
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
	newEnd := len(dq.chunkPitch) - dq.eFree
	dq.chunks = dq.chunkPitch[dq.sFree:newEnd]
}

func (dq *Deque[T]) shrinkEnd() {
	n := len(dq.chunkPitch)
	newEnd := n - dq.eFree - 1
	c := dq.chunkPitch[newEnd]
	dq.chunkPitch[newEnd] = nil
	dq.chunkPool.Put(c)
	dq.eFree++
	dq.chunks = dq.chunkPitch[dq.sFree:newEnd]

	if dq.sFree+dq.eFree < n {
		return
	}
	dq.sFree = n / 2
	dq.eFree = n - dq.sFree
}

func (dq *Deque[T]) shrinkStart() {
	c := dq.chunkPitch[dq.sFree]
	dq.chunkPitch[dq.sFree] = nil
	dq.chunkPool.Put(c)
	dq.sFree++
	curEnd := len(dq.chunkPitch) - dq.eFree
	dq.chunks = dq.chunkPitch[dq.sFree:curEnd]

	n := len(dq.chunkPitch)
	if dq.sFree+dq.eFree < n {
		return
	}
	dq.sFree = n / 2
	dq.eFree = n - dq.sFree
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
}

// TryPopBack tries to remove a value from the back of dq and returns the removed value
// and true if dq is not empty, otherwise it returns false.
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
	if c.e == 0 {
		dq.shrinkEnd()
	}
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
	if c.e == 0 {
		dq.shrinkEnd()
	}
	return r
}

// TryPopFront tries to remove a value from the front of dq and returns the removed value
// and true if dq is not empty, otherwise it returns false.
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
	if c.s == dq.chunkSize {
		dq.shrinkStart()
	}
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
	if c.s == dq.chunkSize {
		dq.shrinkStart()
	}
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
		if c.s == dq.chunkSize {
			dq.shrinkStart()
		}
		n -= num
	}
	return buf
}

// Back returns the last value of dq
// and true if dq is not empty, otherwise it returns false.
func (dq *Deque[T]) Back() (_ T, ok bool) {
	n := len(dq.chunks)
	if n == 0 {
		return *new(T), false
	}
	return dq.chunks[n-1].back()
}

// Front returns the first value of dq
// and true if dq is not empty, otherwise it returns false.
func (dq *Deque[T]) Front() (_ T, ok bool) {
	n := len(dq.chunks)
	if n == 0 {
		return *new(T), false
	}
	return dq.chunks[0].front()
}

// IsEmpty returns whether dq is empty.
func (dq *Deque[T]) IsEmpty() bool {
	n := len(dq.chunks)
	return n == 0 || n == 1 && dq.chunks[0].e == dq.chunks[0].s
}

// Len returns the number of values in dq.
func (dq *Deque[T]) Len() int {
	n := len(dq.chunks)
	switch n {
	case 0:
		return 0
	case 1:
		return dq.chunks[0].e - dq.chunks[0].s
	default:
		return dq.chunks[0].e - dq.chunks[0].s + (n-2)*dq.chunkSize + dq.chunks[n-1].e
	}
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
	var idx int
	for _, c := range dq.chunks {
		for i := c.s; i < c.e; i++ {
			vals[idx] = c.data[i]
			idx++
		}
	}
	return vals
}

// Range iterates all the values in dq.
func (dq *Deque[T]) Range(f func(i int, v T) bool) {
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

// Peek returns the value at idx.
func (dq *Deque[T]) Peek(idx int) T {
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

// Replace replaces the value at idx.
func (dq *Deque[T]) Replace(idx int, v T) {
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

// Option represents the option of the Deque.
type Option func(*optionHolder)

// WithChunkSize sets the chunk size of a Deque.
func WithChunkSize(n int) Option {
	return func(holder *optionHolder) {
		if n > 0 {
			holder.chunkSize = n
		}
	}
}
