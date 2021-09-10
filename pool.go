package bufp

import "sync"

var simplePool *sync.Pool

func init() {
	NewPool(_size)
}

// NewPool constructs a new sync.Pool.
func NewPool(size int) *sync.Pool {
	return &sync.Pool{
		New: func() interface{} {
			return &Buffer{bs: make([]byte, 0, size)}
		},
	}
}

func SetSimplePool(pool *sync.Pool) {
	simplePool = pool
}

func Pool() *sync.Pool {
	return simplePool
}

func Put(buf *Buffer) {
	simplePool.Put(buf)
}

func Get() *Buffer {
	buf := simplePool.Get().(*Buffer)
	buf.Reset()
	buf.pool = simplePool
	return buf
}
