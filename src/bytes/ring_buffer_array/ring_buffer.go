package ring_buffer_array

import (
	"github.com/lk565434471/zeus/src/bytes"
	"io"
)

type RingBuffer struct {
	buf      []byte
	cap      int
	writePos int
	readPos  int
	empty    bool
	canGrow  bool
}

func New(cap int, canGrow bool) *RingBuffer {
	return &RingBuffer{
		buf:      make([]byte, cap),
		cap:      cap,
		writePos: 0,
		readPos:  0,
		empty:    true,
		canGrow:  canGrow,
	}
}

func NewDefault() *RingBuffer {
	return New(bytes.MinimumReadSize, false)
}

func (b *RingBuffer) Read(p []byte) (int, error) {
	n := len(p)

	if n == 0 || b.empty {
		return 0, nil
	}

	size := b.Size()

	if n > size {
		n = size
	}

	if b.writePos > b.readPos {
		copy(p, b.buf[b.readPos:b.readPos+n])
		b.readPos += n
		b.tryReset()
		return n, nil
	}

	if (b.readPos + n) <= b.cap {
		copy(p, b.buf[b.readPos:b.readPos+n])
	} else {
		copy(p, b.buf[b.readPos:])
		pos1 := b.cap - b.readPos
		pos2 := n - pos1
		copy(p[pos1:], b.buf[:pos2])
	}

	b.readPos = (b.readPos + n) % b.cap
	b.tryReset()

	return n, nil
}

func (b *RingBuffer) ReadByte() (byte, error) {
	if b.Size() == 0 {
		return 0, io.EOF
	}

	v := b.buf[b.readPos]
	b.readPos++

	if b.readPos == b.cap {
		b.readPos = 0
	}

	b.tryReset()

	return v, nil
}

func (b *RingBuffer) ReadBytes() ([]byte, error) {
	size := b.Size()

	if size == 0 {
		return nil, io.EOF
	}

	buf := make([]byte, size)
	n, err := b.Read(buf)

	if err != nil {
		return nil, err
	}

	if n != size {
		return nil, bytes.ErrInvalidRead
	}

	return buf, nil
}

// ReadFrom implement the ReaderFrom interface
func (b *RingBuffer) ReadFrom(r io.Reader) (n int64, err error) {
	buf := make([]byte, bytes.MinimumReadSize)
	size := 0
	l := 0

	for {
		size, err = r.Read(buf)

		if ((size == 0) && (err == io.EOF)) ||
			(err != nil) {
			break
		}

		l, err = b.Write(buf)

		if (l != size) || (err != nil) {
			break
		}

		n += int64(l)
	}

	return
}

// Write writes len(p) bytes from p to the underlying data stream.
func (b *RingBuffer) Write(p []byte) (int, error) {
	size := len(p)

	if size == 0 {
		return 0, nil
	}

	available := b.Available()

	if size > available {
		if !b.canGrow {
			return 0, bytes.ErrInsufficientSpace
		}

		b.Grow(b.cap * 2)
		available = b.Available()
	}

	if b.writePos >= b.readPos {
		if available >= size {
			copy(b.buf[b.writePos:], p)
			b.writePos += size
		} else {
			copy(b.buf[b.writePos:], p[:available])
			left := size - available
			copy(b.buf, p[left:])
			b.writePos += left
		}
	} else {
		copy(b.buf[b.writePos:], p)
		b.writePos += size
	}

	if b.writePos == b.cap {
		b.writePos = 0
	}

	b.empty = false

	return size, nil
}

func (b *RingBuffer) WriteByte(c byte) {
	if 1 > b.Available() {
		if !b.canGrow {
			return
		}

		b.Grow(b.cap * 2)
	}

	b.buf[b.writePos] = c
	b.writePos++

	if b.writePos == b.cap {
		b.writePos = 0
	}

	b.empty = false
}

func (b *RingBuffer) WriteString(str string) (int, error) {
	return b.Write(bytes.StringToBytes(str))
}

func (b *RingBuffer) WriteTo(w io.Writer) (n int64, err error) {
	if b.empty {
		return
	}

	var l int
	size := b.Size()

	if b.writePos > b.readPos {
		l, err = w.Write(b.buf[b.readPos : b.readPos+size])
		n = int64(l)

		if l != size {
			return n, bytes.ErrInvalidWrite
		}

		b.readPos += l
		b.tryReset()

		return
	}

	if b.cap >= (b.readPos + size) {
		l, err = w.Write(b.buf[b.readPos : b.readPos+size])
		n = int64(l)

		if l != size {
			return n, bytes.ErrInvalidWrite
		}

		return
	}

	l, err = w.Write(b.buf[b.readPos:])
	n = int64(l)

	if err != nil {
		return n, err
	}

	b.readPos = (b.readPos + l) % b.cap

	return
}

func (b *RingBuffer) Grow(newCap int) {
	if b.cap >= newCap {
		return
	}

	buf := make([]byte, newCap)
	copy(buf, b.buf)
	b.buf = buf
	b.cap = newCap
}

func (b *RingBuffer) Discard(n int) (int, error) {
	if 0 >= n {
		return 0, nil
	}

	size := b.Size()

	if size > n {
		b.readPos = (b.readPos + n) % b.cap
		return n, nil
	}

	b.Reset()
	return size, nil
}

func (b *RingBuffer) Peek(n int) ([]byte, error) {
	if b.empty {
		return make([]byte, 0), nil
	}

	readableSize := b.Size()

	if readableSize > n {
		readableSize = n
	}

	buf := make([]byte, readableSize)

	if (b.writePos > b.readPos) ||
		(b.cap >= (b.readPos + readableSize)) {
		copy(buf, b.buf[b.readPos:b.readPos+readableSize])
	} else {
		pos := copy(buf, b.buf[b.readPos:])
		left := b.cap - b.readPos
		copy(buf[pos:], b.buf[:left])
	}

	return buf, nil
}

func (b *RingBuffer) PeekAll() ([]byte, error) {
	return b.Peek(b.Size())
}

// Size returns the number of bytes that can be read from the buffer.
func (b *RingBuffer) Size() int {
	if b.readPos == b.writePos {
		if b.empty {
			return 0
		}

		return b.cap
	}

	if b.writePos > b.readPos {
		return b.writePos - b.readPos
	}

	return b.cap - b.readPos + b.writePos
}

func (b *RingBuffer) Capacity() int {
	return b.cap
}

func (b *RingBuffer) Available() int {
	if b.readPos == b.writePos {
		if b.empty {
			return b.cap
		}

		return 0
	}

	if b.readPos > b.writePos {
		return b.readPos - b.writePos
	}

	return b.cap - b.writePos + b.readPos
}

func (b *RingBuffer) Full() bool {
	return (b.readPos == b.writePos) && !b.empty
}

func (b *RingBuffer) Empty() bool {
	return b.empty
}

func (b *RingBuffer) Reset() {
	b.empty = true
	b.readPos = 0
	b.writePos = 0
}

func (b *RingBuffer) tryReset() {
	if (b.readPos == b.writePos) && !b.empty {
		b.Reset()
	}
}
