package buffer

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewBuffer(t *testing.T) {
	b := NewBuffer()
	require.NotNil(t, b, "NewBuffer should not return nil")
	require.NotNil(t, b.pools, "Buffer pools map should not be nil")
}

func TestBufferGet(t *testing.T) {
	b := NewBuffer()

	// Test getting a buffer of specific size
	size := 1024
	buf := b.Get(size)

	require.Equal(t, size, len(buf), "Buffer length should match requested size")
	require.Equal(t, size, cap(buf), "Buffer capacity should match requested size")
}

func TestBufferPut(t *testing.T) {
	b := NewBuffer()

	// Test putting a buffer back
	size := 1024
	buf := b.Get(size)

	// Modify the buffer
	for i := 0; i < size; i++ {
		buf[i] = byte(i % 256)
	}

	// Put it back
	b.Put(buf)

	// Get a new buffer of the same size
	newBuf := b.Get(size)

	// The new buffer should be clean (all zeros)
	for i := 0; i < size; i++ {
		require.Equal(t, byte(0), newBuf[i], "Buffer should be clean at index %d", i)
	}
}

func TestBufferPutNil(t *testing.T) {
	b := NewBuffer()

	// Test putting nil buffer
	require.NotPanics(t, func() {
		b.Put(nil)
	}, "Putting nil buffer should not panic")
}

func TestBufferReset(t *testing.T) {
	b := NewBuffer()

	// Create some buffers
	buf1 := b.Get(1024)
	buf2 := b.Get(2048)

	// Put them back
	b.Put(buf1)
	b.Put(buf2)

	// Reset the buffer
	b.Reset()

	// Get a new buffer of the same size
	newBuf := b.Get(1024)

	// The new buffer should be clean
	for i := 0; i < 1024; i++ {
		require.Equal(t, byte(0), newBuf[i], "Buffer should be clean after reset at index %d", i)
	}
}

func TestDefaultBuffer(t *testing.T) {
	b := DefaultBuffer()
	require.NotNil(t, b, "DefaultBuffer should not return nil")
	require.NotNil(t, b.pools, "DefaultBuffer pools map should not be nil")
}

func TestConcurrentBufferOperations(t *testing.T) {
	b := NewBuffer()
	done := make(chan bool)

	// Test concurrent operations
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				size := 1024 * (j + 1)
				buf := b.Get(size)
				require.Equal(t, size, len(buf), "Concurrent buffer length should match requested size")
				require.Equal(t, size, cap(buf), "Concurrent buffer capacity should match requested size")
				b.Put(buf)
			}
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}
