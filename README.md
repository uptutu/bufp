# Bufp

![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/uptutu/bufp?style=flat-square)
[![GitHub tag (latest SemVer)](https://img.shields.io/github/tag/uptutu/bufp)](https://github.com/uptutu/bufp)
[![GoDoc](https://godoc.org/github.com/uptutu/bufp?status.svg)](https://pkg.go.dev/github.com/uptutu/bufp)
[![Go Report Card](https://goreportcard.com/badge/github.com/uptutu/bufp)](https://goreportcard.com/report/github.com/uptutu/bufp)
[![Unit-Tests](https://github.com/uptutu/bufp/workflows/Unit-Tests/badge.svg)](https://github.com/uptutu/bufp/actions)
[![Coverage Status](https://coveralls.io/repos/github/uptutu/bufp/badge.svg?branch=master)](https://coveralls.io/github/uptutu/bufp?branch=master)

💪 Useful utils Buffer Pool for decrease Go GC stress.

## Usage

The principle of this package is based on `sync.Pool`. A layer of `manager` is wrapped around it for buffer size
management.

### Install

```bash
$ go get github.com/uptutu/bufp
```

### Use

#### Manager

The manager is a `deque` at the Underlying logic , with each node managing a `sync.pool` of one size

package has a default global manager managing a 1kib Pool node.

```go
package main

import (
	"github.com/uptutu/bufp/manager"
	"sync"
)

func main() {
	// Initialize pools of different sizes in Kib unit or Mib unit
	// Of cares, you don't have to initialize
	manager.InitKibPools(1, 2, 3, 4)
	manager.InitMibPools(1, 2, 3, 4)

	// get pool of the size you want from this manager
	// _, ok := pool.(*sync.Pool); ok == true
	size := manager.DefaultSize1KiB
	pool, ok := manager.Get(size)

	// here is 2 way to Set diff size pool to manager
	p := &sync.Pool{New: func() interface{} { return &bufp.Buffer{bs: make([]byte, 0, 1023)} }}
	manager.Set(1023, p)
	// if *sync.Pool is nil then manager will accord your size new a Pool.
	// this size will trim to 1024
	manager.Set(1023, nil)
	// ok == true
	pool, ok := manager.Get(1024)

	// this is a magic func.
	// accord your passing on references size and trim it to find a suit size pool in manager to you
	// if no suit pool this will return a nil pointer and false flag
	// if you try multi times, manager will new this size pool for you
	p, ok = manager.RightOne(sizeYouWant)

	// this func serve you Service Logic in the lambda func
	manager.Serve(len("content"), func(buf *bufp.Buffer) err {
		buf.WriteString("content")
		return nil
	})
}

```

#### Buffer

```go
size := 1024
buffer := bufp.NewBuffer(size)

// append
buffer.AppebdByte(byte(oneByteUnit))
buffer.AppendString(string(s))
buffer.AppendInt(int64(10))
buffer.AppentTime(time.Now(), "2021-2-21 15:30:23")
buffer.AppendUint(uint64(10))
buffer.AppendBool(false)
buffer.AppendFloat(3.1415, 10)
buffer.Write([]byte("content"))
buffer.WriteByte('c')
buffer.WriteString("string")

buffer.TrimNewline()

// get buffer content
buffer.Len()
buffer.Cap()
buffer.Bytes()
buffer.String()

```

### Pool

this package has a default 1kib size buffer global pool named simplePool.

```go
size := 2014
// New: return *bufp.Buffer
pool := bufp.NewPool(size) // *sync.Pool

simplePool := bufp.Pool()
// this buffer default was 1kib size buffer
buffer := simplePool.Get()
// easy way get this default 1kib buffer
buffer := Get()
// easy way put buffer to simplePool
bufp.Put(buffer)

// you can change the default 1 kib buffer size by this func
bufp.Set(&sync.Pool{New:func (){return &bufp.Buffer{bs: make([]byte, 0, sizeYouWant)}}})
```

## License

[MIT](LICENSE)
