package manager

import (
	"bufp"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
)

func TestNewNode(t *testing.T) {
	p := sync.Pool{New: nil}
	n := newNode(32, &p)
	assert.Equal(t, n.size, 32)
	assert.Equal(t, n.p, &p)
}

func TestManager_SetNode(t *testing.T) {
	m := manager{
		head:   nil,
		tail:   nil,
		middle: nil,
		len:    0,
	}
	m.setNode(1, nil)

	assert.Equal(t, m.len, 1)
	assert.Equal(t, m.head.size, 1)
	assert.Equal(t, m.head, m.tail)
	assert.Nil(t, m.middle)

	m.setNode(DefaultSize1MiB*2, nil)
	assert.Equal(t, m.len, 2)
	assert.Equal(t, m.head.next.size, DefaultSize1MiB*2)
	assert.NotEqual(t, m.tail, m.head)
	assert.NotNil(t, m.middle)

	p := bufp.NewPool(DefaultSize1MiB)
	m.setNode(DefaultSize1MiB, p)
	assert.Equal(t, m.len, 3)
	assert.Equal(t, m.middle, m.head.next)
	assert.NotEqual(t, m.tail, p)
	assert.NotEqual(t, m.head, p)
}

func TestInit(t *testing.T) {
	assert.NotNil(t, m)
	assert.Equal(t, m.len, 1)
	assert.Nil(t, m.head.next)
	assert.Equal(t, m.head.size, DefaultSize1KiB)
}

func TestInitKibPools(t *testing.T) {
	test := []struct {
		data []int
		want []int
	}{
		{data: []int{1, 2, 3}, want: []int{1 * DefaultSize1KiB, 2 * DefaultSize1KiB, 3 * DefaultSize1KiB}},
		{data: []int{2, 3, 1}, want: []int{1 * DefaultSize1KiB, 2 * DefaultSize1KiB, 3 * DefaultSize1KiB}},
		{data: []int{2, 2, 3}, want: []int{2 * DefaultSize1KiB, 3 * DefaultSize1KiB}},
	}

	for _, try := range test {
		m = manager{}
		err := InitKibPools(try.data...)
		assert.Nil(t, err)
		var gets []int
		for traveler := m.head; traveler != nil; traveler = traveler.next {
			gets = append(gets, traveler.size)
		}
		assert.Equal(t, gets, try.want)
	}
}

func TestInitMibPools(t *testing.T) {
	test := []struct {
		data []int
		want []int
	}{
		{data: []int{1, 2, 3}, want: []int{1 * DefaultSize1MiB, 2 * DefaultSize1MiB, 3 * DefaultSize1MiB}},
		{data: []int{2, 3, 1}, want: []int{1 * DefaultSize1MiB, 2 * DefaultSize1MiB, 3 * DefaultSize1MiB}},
		{data: []int{2, 2, 3}, want: []int{2 * DefaultSize1MiB, 3 * DefaultSize1MiB}},
	}

	for _, try := range test {
		m = manager{}
		err := InitMibPools(try.data...)
		assert.Nil(t, err)
		var gets []int
		for traveler := m.head; traveler != nil; traveler = traveler.next {
			gets = append(gets, traveler.size)
		}
		assert.Equal(t, gets, try.want)
	}
}

func TestTrimSize(t *testing.T) {
	test := []struct {
		data []int
		want []int
	}{
		{data: []int{255, DefaultSize1KiB, 5000, 7340033, 1073741824, DefaultSize1MiB, DefaultSize1MiB + 1},
			want: []int{DefaultSize1KiB, DefaultSize1KiB, 5 * DefaultSize1KiB, 8 * DefaultSize1MiB, 1073741824, DefaultSize1MiB, DefaultSize1MiB * 2}},
	}
	for _, try := range test {
		for i := 0; i < len(try.data); i++ {
			get := trimSize(try.data[i])
			assert.Equal(t, get, try.want[i])
		}
	}
}

func TestFind(t *testing.T) {
	m = manager{}
	err := InitMibPools(5, 6, 7, 8)
	assert.Nil(t, err)

	n, ok := m.find(6)
	assert.Nil(t, n)
	assert.False(t, ok)

	n, ok = m.find(6 * DefaultSize1MiB)
	assert.NotNil(t, n)
	assert.True(t, ok)
}

func TestGet(t *testing.T) {
	p, ok := Get(5)
	assert.False(t, ok)
	assert.Nil(t, p)
	_ = InitMibPools(5)
	p, ok = Get(5 * DefaultSize1KiB)
	assert.False(t, ok)
	assert.Nil(t, p)
	p, ok = Get(5 * DefaultSize1MiB)
	assert.True(t, ok)
	assert.NotNil(t, p)
}

func TestSet(t *testing.T) {
	m = manager{}
	p, ok := Get(5 * DefaultSize1MiB)
	assert.Nil(t, p)
	assert.False(t, ok)

	tp := &sync.Pool{New: nil}
	anotherTp := &sync.Pool{}
	Set(5*DefaultSize1MiB, tp)
	p, ok = Get(5 * DefaultSize1MiB)
	assert.True(t, ok)
	assert.Equal(t, p, tp)

	Set(4*DefaultSize1MiB+1, anotherTp)
	p, ok = Get(5 * DefaultSize1MiB)
	assert.True(t, ok)
	assert.Equal(t, p, anotherTp)
}

func TestRightOne(t *testing.T) {
	m = manager{}

	_ = InitMibPools(1, 8)
	_ = InitKibPools(1, 5)
	//pool := &sync.Pool{}
	//Set(7*DefaultSize1MiB+4*DefaultSize1KiB, pool)

	test := []struct {
		data []int
		want []struct {
			w  int
			ok bool
		}
	}{
		{data: []int{255, DefaultSize1KiB, 5000, 7341033, 1073741824, DefaultSize1MiB, DefaultSize1MiB + 1},
			want: []struct {
				w  int
				ok bool
			}{
				{DefaultSize1KiB, true},
				{DefaultSize1KiB, true},
				{5 * DefaultSize1KiB, true},
				{8 * DefaultSize1MiB, true},
				{1073741824, false},
				{DefaultSize1MiB, true},
				{DefaultSize1MiB + 1, false}}}}
	for _, try := range test {
		for i := 0; i < len(try.data); i++ {
			get, ok := RightOne(try.data[i])
			p, _ := Get(try.want[i].w)
			assert.Equal(t, get, p)
			assert.Equal(t, ok, try.want[i].ok)
		}
	}
}

func TestServe(t *testing.T) {
	largeContent := make([]byte, 5555555555, 5555555555)
	Set(5555555555, bufp.NewPool(5555555555))

	writeContent2BufferDoSomething := func(buffer *bufp.Buffer) error {
		c, err := buffer.Write(largeContent)
		assert.Equal(t, 5555555555, c)
		return err
	}

	err := Serve(len(largeContent), writeContent2BufferDoSomething)
	assert.Nil(t, err)

	littleContent := make([]byte, 125, 125)
	err = Serve(len(littleContent), writeContent2BufferDoSomething)
	assert.Nil(t, err)
}

func TestFailAndSetAttemptTimes(t *testing.T) {
	n, ok := m.find(2)
	assert.Nil(t, n)
	assert.False(t, ok)
	assert.Equal(t, 1, m.len)
	SetAttemptTimes(6)
	for i := 0; i < 6; i++ {
		m.fail(2)
	}
	assert.Equal(t, 2, m.len)
	n, ok = m.find(2)
	assert.NotNil(t, n)
	assert.True(t, ok)
	m.fail(2)
	assert.Equal(t, 2, m.len)

}

func BenchmarkServe5MibInPoolWay(b *testing.B) {
	largeContent := make([]byte, 5*DefaultSize1MiB, 5*DefaultSize1MiB)
	Set(5*DefaultSize1MiB, bufp.NewPool(5*DefaultSize1MiB))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		writeContent2BufferDoSomething := func(buffer *bufp.Buffer) error {
			_, err := buffer.Write(largeContent)
			return err
		}

		Serve(len(largeContent), writeContent2BufferDoSomething)
	}
}

func BenchmarkServe5mibInNormalWay(b *testing.B) {
	m.reset()
	largeContent := make([]byte, 5*DefaultSize1MiB, 5*DefaultSize1MiB)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		writeContent2BufferDoSomething := func(buffer *bufp.Buffer) error {
			_, err := buffer.Write(largeContent)
			return err
		}

		Serve(len(largeContent), writeContent2BufferDoSomething)
	}
}

func BenchmarkServe255kibInPoolWay(b *testing.B) {
	largeContent := make([]byte, 255*DefaultSize1KiB, 255*DefaultSize1KiB)
	Set(255*DefaultSize1KiB, bufp.NewPool(255*DefaultSize1KiB))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		writeContent2BufferDoSomething := func(buffer *bufp.Buffer) error {
			_, err := buffer.Write(largeContent)
			return err
		}

		Serve(len(largeContent), writeContent2BufferDoSomething)
	}
}

func BenchmarkServe255kibInNormalWay(b *testing.B) {
	m.reset()
	largeContent := make([]byte, 255*DefaultSize1KiB, 255*DefaultSize1KiB)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		writeContent2BufferDoSomething := func(buffer *bufp.Buffer) error {
			_, err := buffer.Write(largeContent)
			return err
		}
		Serve(len(largeContent), writeContent2BufferDoSomething)
	}
}
