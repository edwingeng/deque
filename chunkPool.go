package deque

import (
	"sync"
	"sync/atomic"
)

type chunkPool struct {
	sync.Pool
	numChunksAllocated int64
}

func newChunkPool(newChunk func() interface{}) *chunkPool {
	var x chunkPool
	x.New = func() interface{} {
		atomic.AddInt64(&x.numChunksAllocated, 1)
		return newChunk()
	}
	return &x
}
