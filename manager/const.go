package manager

import "errors"

const (
	DefaultSize1KiB = 1 << ((iota + 1) * 10) // 1 KiB
	DefaultSize1MiB

	MaxSizeUnit = 1<<10
)

var (
	InvalidSizeErr = errors.New("invalid size: Data size exceeds unit maximum")
)
