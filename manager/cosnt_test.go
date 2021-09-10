package manager

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestValid(t *testing.T) {
	assert.Equal(t, DefaultSize1KiB, 1<<10)
	assert.Equal(t, DefaultSize1MiB, 1<<20)
}
