package utils

import "bytes"

// BufferPool maintains a pool of reusable byte buffers.
type BufferPool struct {
	BufferSize int
	MaxSize    int

	c chan *bytes.Buffer
}

// NewBufferPool constructs a new buffer pool and initializes half the pool.
func NewBufferPool(initialSize, maxSize, poolSize int) *BufferPool {
	p := &BufferPool{
		BufferSize: initialSize,
		MaxSize:    maxSize,
		c:          make(chan *bytes.Buffer, poolSize),
	}

	for i := 0; i < poolSize/2; i++ {
		p.c <- bytes.NewBuffer(make([]byte, 0, p.BufferSize))
	}

	return p
}

// Get a buffer from the pool. Consumers should Give() the buffer back after
// use.
func (p *BufferPool) Get() *bytes.Buffer {
	select {
	case b := <-p.c:
		return b
	default:
		return bytes.NewBuffer(make([]byte, 0, p.BufferSize))
	}
}

// Give a buffer back to the pool. Consumers should not use the buffer after
// calling Give.
func (p *BufferPool) Give(b *bytes.Buffer) {
	if b.Cap() <= p.MaxSize {
		b.Reset()
		select {
		case p.c <- b:
		default:
			// Discard
		}
	}
}
