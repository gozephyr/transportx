// Package buffer provides efficient buffer management using a sync.Pool.
package buffer

import (
	"sync"
)

// Buffer provides efficient buffer management using a sync.Pool
// to reduce memory allocations and garbage collection pressure.
type Buffer struct {
	pools map[int]*sync.Pool // Map of size to pool
	mu    sync.RWMutex       // Protects pools map
}

// NewBuffer creates a new buffer manager.
func NewBuffer() *Buffer {
	return &Buffer{
		pools: make(map[int]*sync.Pool),
	}
}

// Get retrieves a buffer of the specified size from the pool.
// If no buffer is available, a new one is created.
func (b *Buffer) Get(size int) []byte {
	b.mu.RLock()
	pool, exists := b.pools[size]
	b.mu.RUnlock()

	if !exists {
		b.mu.Lock()
		pool, exists = b.pools[size]
		if !exists {
			pool = &sync.Pool{
				New: func() any {
					buf := make([]byte, size)
					return &buf
				},
			}
			b.pools[size] = pool
		}
		b.mu.Unlock()
	}

	return *pool.Get().(*[]byte)
}

// Put returns a buffer to the pool.
func (b *Buffer) Put(buf []byte) {
	if buf == nil {
		return
	}
	size := cap(buf)
	b.mu.RLock()
	pool, exists := b.pools[size]
	b.mu.RUnlock()

	if exists {
		// Reset the buffer
		buf = buf[:size]
		// Zero out the buffer
		for i := range buf {
			buf[i] = 0
		}
		pool.Put(&buf)
	}
}

// Reset clears all buffer pools.
func (b *Buffer) Reset() {
	b.mu.Lock()
	b.pools = make(map[int]*sync.Pool)
	b.mu.Unlock()
}

// DefaultBuffer returns a new buffer manager.
func DefaultBuffer() *Buffer {
	return NewBuffer()
}
