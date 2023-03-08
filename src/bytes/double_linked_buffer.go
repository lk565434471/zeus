package bytes

import (
	"errors"
)

var (
	errBufferIsEmpty = errors.New("buffer is empty")
)

type node struct {
	buf  []byte
	size int
	prev *node
	next *node
}

type DoubleLinkedBuffer struct {
	head *node
	tail *node
	size int
}

func (dlb *DoubleLinkedBuffer) Append(p []byte) {
	size := len(p)

	if size == 0 {
		return
	}

	n := &node{
		prev: dlb.tail,
		size: size,
	}
	n.buf = append(n.buf, p...)

	if dlb.tail != nil {
		dlb.tail.next = n
	} else {
		dlb.head = n
	}

	dlb.tail = n
	dlb.size += n.size
}

func (dlb *DoubleLinkedBuffer) Prepend(p []byte) {
	size := len(p)

	if size == 0 {
		return
	}

	n := &node{
		next: dlb.head,
		size: size,
	}
	n.buf = append(n.buf, p...)

	if dlb.head != nil {
		dlb.head.prev = n
	} else {
		dlb.tail = n
	}

	dlb.head = n
	dlb.size += n.size
}

func (dlb *DoubleLinkedBuffer) InsertBefore(n *node, p []byte) {
	if n == nil {
		return
	}

	size := len(p)

	if size == 0 {
		return
	}

	if n == dlb.head {
		dlb.Prepend(p)
		return
	}

	newNode := &node{
		prev: n.prev,
		next: n,
		size: size,
	}
	newNode.buf = append(newNode.buf, p...)
	n.prev.next = newNode
	n.prev = newNode
	dlb.size += newNode.size
}

func (dlb *DoubleLinkedBuffer) InsertAfter(n *node, p []byte) {
	if n == nil {
		return
	}

	size := len(p)

	if size == 0 {
		return
	}

	if n == dlb.tail {
		dlb.Append(p)
		return
	}

	newNode := &node{
		prev: n,
		next: n.next,
		size: size,
	}
	newNode.buf = append(newNode.buf, p...)
	n.next.prev = newNode
	n.next = newNode
	dlb.size += newNode.size
}

func (dlb *DoubleLinkedBuffer) Remove(n *node) {
	if n == nil {
		return
	}

	if n == dlb.head {
		dlb.head = n.next
	} else {
		n.prev.next = n.next
	}

	if n == dlb.tail {
		dlb.tail = n.prev
	} else {
		n.next.prev = n.prev
	}

	dlb.size -= n.size
}

// Read 实现 io.Reader 接口
func (dlb *DoubleLinkedBuffer) Read(p []byte) (size int, err error) {
	capacity := len(p)

	if capacity == 0 {
		return size, errBufferIsEmpty
	}

	left := capacity

	for n := dlb.head; n != nil && left > 0; n = n.next {
		if left >= n.size {
			nBytes := copy(p[size:], n.buf)
			size += nBytes
			left -= nBytes
		} else {
			nBytes := copy(p[size:], n.buf[:left])
			size += nBytes
			left -= nBytes
		}
	}

	left = size

	for {
		if 0 >= left {
			break
		}

		n := dlb.head

		if n == nil {
			break
		}

		if left >= n.size {
			left -= n.size
			dlb.Remove(n)
		} else {
			n.buf = n.buf[left:]
			dlb.size -= left
			left -= left
		}
	}

	return
}

func (dlb *DoubleLinkedBuffer) Bytes() []byte {
	buff := make([]byte, dlb.size)
	size := 0

	for n := dlb.head; n != nil; n = n.next {
		size += copy(buff[size:], n.buf)
	}

	return buff
}
