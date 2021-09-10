package manager

import (
	"bufp"
	"sync"
)

var m manager

type manager struct {
	head        *node
	tail        *node
	mbFirstNode *node
	len         int
}

func (m *manager) setNode(size int, pool *sync.Pool) {
	defer func() { m.len++ }()
	if pool == nil {
		pool = bufp.NewPool(size)
	}
	n := newNode(size, pool)

	if size>>20 != 0 && m.mbFirstNode == nil {
		m.mbFirstNode = &n
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
			if n.size>>20 != 0 && traveler == m.mbFirstNode {
				m.mbFirstNode = &n
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
	endCondition := m.mbFirstNode
	traveler := m.head
	if size>>20 != 0 {
		traveler = m.mbFirstNode
		endCondition = nil
	}

	for ; traveler != endCondition; traveler = traveler.next {
		if traveler.size == size {
			return traveler, true
		}
	}
	return nil, false
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

func init() {
	m = manager{
		head: nil,
		tail: nil,
		len:  0,
	}
	m.setNode(DefaultSize1KiB, nil)
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
	size = trimSize(size)
	m.setNode(size, pool)
}

func RightOne(size int) (*sync.Pool, bool) {
	size = trimSize(size)
	endCondition := m.mbFirstNode
	traveler := m.head
	if size>>20 != 0 {
		traveler = m.mbFirstNode
		endCondition = nil
	}

	for ; traveler != endCondition; traveler = traveler.next {
		if traveler.size >= size && (traveler.size-size) < (1<<20) {
			return traveler.p, true
		}
	}
	return nil, false
}

func Serve(size int, fn func(bufp.Buffer) error) error {
	if pool, ok := RightOne(size); ok {
		buffer := pool.Get().(bufp.Buffer)
		err := fn(buffer)
		buffer.Free()
		return err
	}
	buffer := bufp.Buffer{}
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
