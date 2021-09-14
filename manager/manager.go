package manager

import (
	"bufp"
	"errors"
	"sync"
)

func init() {
	m = manager{
		failCase: make(map[int]int),
		len:      0,
	}
	m.setNode(DefaultSize1KiB, nil)
}

var (
	m             manager
	attemptTimes  = 10
	attachableGap = 5 * DefaultSize1KiB
)

func SetAttemptTimes(size int) {
	attemptTimes = size
}

type manager struct {
	head     *node
	tail     *node
	middle   *node
	failCase map[int]int
	len      int
}

func (m *manager) setNode(size int, pool *sync.Pool) {
	defer func() { m.len++ }()
	if pool == nil {
		pool = bufp.NewPool(size)
	}
	n := newNode(size, pool)

	if size>>20 != 0 && m.middle == nil {
		m.middle = &n
	}
	if m.head == nil {
		m.head = &n
		m.tail = &n
		return
	}

	for traveler := m.head; traveler != nil; traveler = traveler.next {
		if traveler.size >= size {
			if traveler.size == size {
				traveler.p = pool
				m.len--
				return
			}
			n.previous = traveler.previous
			n.next = traveler
			if n.size>>20 != 0 && traveler == m.middle {
				m.middle = &n
			}
			if traveler.previous == nil {
				m.head = &n
				traveler.previous = &n
				return
			}
			traveler.previous.next = &n
			traveler.previous = &n
			return
		}
	}

	m.tail.next = &n
	n.previous = m.tail
	m.tail = &n

	return
}

func (m manager) find(size int) (*node, bool) {
	endCondition := m.middle
	traveler := m.head
	if size>>20 != 0 {
		traveler = m.middle
		endCondition = nil
	}

	for ; traveler != endCondition; traveler = traveler.next {
		if traveler.size == size {
			return traveler, true
		}
	}
	return nil, false
}

func (m *manager) reset() {
	m.head = nil
	m.tail = nil
	m.middle = nil
	m.len = 0
}

func (m *manager) fail(size int) {
	if val, ok := m.failCase[size]; ok {
		m.failCase[size]++
		if val+1 >= attemptTimes {
			m.setNode(size, nil)
		}
		return
	}
	m.failCase[size] = 1
}

type node struct {
	size     int
	p        *sync.Pool
	next     *node
	previous *node
}

func newNode(size int, pool *sync.Pool) node {
	return node{
		size:     size,
		p:        pool,
		next:     nil,
		previous: nil,
	}
}

func InitKibPools(sizes ...int) error {
	for _, s := range sizes {
		if s >= MaxSizeUnit {
			return InvalidSizeErr
		}
		originSize := DefaultSize1KiB * s
		m.setNode(originSize, nil)
	}
	return nil
}

func InitMibPools(sizes ...int) error {
	for _, s := range sizes {
		if s >= MaxSizeUnit {
			return InvalidSizeErr
		}
		originSize := DefaultSize1MiB * s
		m.setNode(originSize, nil)
	}
	return nil
}

func Get(size int) (*sync.Pool, bool) {
	if node, ok := m.find(size); ok {
		return node.p, ok
	}
	return nil, false
}

func Set(size int, pool *sync.Pool) {
	if pool == nil {
		size = trimSize(size)
	}
	m.setNode(size, pool)
}

func RightOne(size int) (*sync.Pool, bool) {
	size = trimSize(size)
	endCondition := m.middle
	traveler := m.head
	if size>>20 != 0 {
		traveler = m.middle
		endCondition = nil
	}

	for ; traveler != endCondition; traveler = traveler.next {
		if traveler.size >= size && (traveler.size-size) < attachableGap {
			return traveler.p, true
		}
	}
	m.fail(size)
	return nil, false
}

func Serve(size int, fn func(*bufp.Buffer) error) error {
	if pool, ok := RightOne(size); ok {
		buffer, ok := pool.Get().(*bufp.Buffer)
		if !ok {
			return errors.New("pool get obj err")
		}
		err := fn(buffer)
		buffer.Reset()
		pool.Put(buffer)
		return err
	}
	buffer := bufp.NewBuffer(size)
	return fn(buffer)
}

func trimSize(size int) int {
	flag := 1
	o := size
	addNeed := false
	for {
		if (size >> 10) < MaxSizeUnit {
			if ((1<<10)-1)&size == 0 {
				size >>= 10
				if addNeed {
					size++
				}
				break
			}
			size = (size >> 10) + 1
			break
		}
		flag <<= 1
		if (((1 << 10) - 1) & size) != 0 {
			size = (size >> 10) + 1
			addNeed = true
			continue
		}
		size >>= 10
	}

	switch flag {
	case 1:
		return size * DefaultSize1KiB
	case 1 << 1:
		return size * DefaultSize1MiB
	default:
		return o
	}
}
