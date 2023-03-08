package single_linked_list_buffer

import (
	"github.com/lk565434471/zeus/src/bytes"
	"io"
)

type node struct {
	buf  []byte
	next *node
}

func (n *node) len() int {
	return len(n.buf)
}

type LinkedListBuffer struct {
	head      *node
	tail      *node
	size      int
	bytesSize int
}

func (b *LinkedListBuffer) Read(p []byte) (int, error) {
	capacity := len(p)

	if capacity == 0 {
		return 0, bytes.ErrBufferIsEmpty
	}

	left := capacity
	totalSize := 0
	l := 0

	for n := b.Pop(); n != nil; n = b.Pop() {
		size := n.len()

		if left >= size {
			l = copy(p[totalSize:], n.buf)

			if l != size {
				return totalSize, bytes.ErrInvalidRead
			}
		} else {
			l = copy(p[totalSize:], n.buf[:left])
			b.PushFront(n.buf[left:])

			if l != left {
				return totalSize, bytes.ErrInvalidRead
			}
		}

		totalSize += l
		left -= l

		if 0 >= left {
			break
		}
	}

	return totalSize, nil
}

func (b *LinkedListBuffer) ReadFrom(r io.Reader) (n int64, err error) {
	buf := make([]byte, bytes.MinimumReadSize)
	l := 0

	for {
		l, err = r.Read(buf)

		if l == 0 && err == io.EOF {
			break
		} else if err != io.EOF {
			return
		}

		n += int64(l)
		b.PushBack(buf[:l])
	}

	return
}

func (b *LinkedListBuffer) Write(p []byte) (int, error) {
	capacity := len(p)

	if capacity == 0 {
		return 0, nil
	}

	b.PushBack(p)

	return capacity, nil
}

func (b *LinkedListBuffer) WriteTo(w io.Writer) (totalSize int64, err error) {
	size := 0

	for n := b.Pop(); n != nil; n = b.Pop() {
		size, err = w.Write(n.buf)

		if ((size == 0) && (err == io.EOF)) ||
			err != nil {
			break
		}

		if size != n.len() {
			return totalSize, bytes.ErrInvalidWrite
		}

		totalSize += int64(size)
	}

	return
}

func (b *LinkedListBuffer) PushFront(p []byte) {
	n := len(p)

	if n == 0 {
		return
	}

	buf := make([]byte, n)
	copy(buf, p)
	b.pushFront(&node{
		buf: buf,
	})
}

func (b *LinkedListBuffer) PushBack(p []byte) {
	n := len(p)

	if n == 0 {
		return
	}

	buf := make([]byte, n)
	copy(buf, p)
	b.pushBack(&node{
		buf: buf,
	})
}

func (b *LinkedListBuffer) Peek(size int) (p []byte, err error) {
	if 0 >= size {
		return
	}

	left := size
	pos := 0

	for n := b.head; n != nil && left > 0; n = n.next {
		if left >= n.len() {
			p = append(p[pos:], n.buf...)
			left -= n.len()
		} else {
			p = append(p[pos:], n.buf[:left]...)
			left -= left
		}
	}

	if left != 0 {
		err = io.ErrShortWrite
	}

	return
}

func (b *LinkedListBuffer) Discard(size int) (totalSize int, err error) {
	if 0 >= size {
		return
	}

	left := size

	for n := b.Pop(); n != nil; n = b.Pop() {
		if left >= n.len() {
			totalSize += n.len()
			left -= n.len()
		} else {
			b.PushFront(n.buf[left:])
			totalSize += left
			left -= left
		}

		if 0 >= left {
			break
		}
	}

	if left != 0 {
		err = io.ErrShortBuffer
	}

	return
}

func (b *LinkedListBuffer) Empty() bool {
	return b.head == nil
}

func (b *LinkedListBuffer) Reset() {
	for n := b.Pop(); n != nil; n = b.Pop() {
	}

	b.head = nil
	b.tail = nil
	b.size = 0
	b.bytesSize = 0
}

func (b *LinkedListBuffer) Pop() *node {
	if b.head == nil {
		return nil
	}

	n := b.head
	b.head = n.next

	if b.head == nil {
		b.tail = nil
	}

	n.next = nil
	b.size--
	b.bytesSize -= n.len()

	return n
}

func (b *LinkedListBuffer) Bytes() []byte {
	buf := make([]byte, b.BytesSize())
	pos := 0

	for n := b.head; n != nil; n = n.next {
		copy(buf[pos:], n.buf)
	}

	return buf
}

func (b *LinkedListBuffer) BytesSize() int {
	return b.bytesSize
}

func (b *LinkedListBuffer) pushFront(n *node) {
	if n == nil {
		return
	}

	if b.head == nil {
		n.next = nil
		b.tail = n
	} else {
		n.next = b.head
	}

	b.head = n
	b.size++
	b.bytesSize += n.len()
}

func (b *LinkedListBuffer) pushBack(n *node) {
	if n == nil {
		return
	}

	if b.tail == nil {
		b.head = n
	} else {
		b.tail.next = n
	}

	n.next = nil
	b.tail = n
	b.size++
	b.bytesSize += n.len()
}
