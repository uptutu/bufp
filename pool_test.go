package bufp

import (
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
)

func TestBuffers(t *testing.T) {
	const dummyData = "dummy data"

	var wg sync.WaitGroup
	for g := 0; g < 10; g++ {
		wg.Add(1)
		go func() {
			for i := 0; i < 100; i++ {
				buf := Get()
				assert.Zero(t, buf.Len(), "Expected truncated buffer")
				assert.NotZero(t, buf.Cap(), "Expected non-zero capacity")

				buf.AppendString(dummyData)
				assert.Equal(t, buf.Len(), len(dummyData), "Expected buffer to contain dummy data")

				Put(buf)
			}
			wg.Done()
		}()
	}
	wg.Wait()
}
