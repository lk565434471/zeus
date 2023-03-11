package ring_linked_list

import (
	"errors"
	"io"
)

var (
	errBufferFull        = errors.New("buffer full")
	errInvalidWriteCount = errors.New("invalid write count")
	errInvalidReadCount  = errors.New("invalid read count")
)

type node struct {
	buf      []byte
	readPos  int
	writePos int
	prev     *node
	next     *node
}

const smallBufferSize = 512

func (n *node) read(p []byte) (m int, err error) {
	if n.empty() {
		n.reset()

		if len(p) == 0 {
			return 0, nil
		}

		return 0, io.EOF
	}

	l := len(p)

	if l > (n.writePos - n.readPos) {
		l = n.writePos - n.readPos
	}

	m = copy(p, n.buf[n.readPos:n.readPos+l])
	n.readPos += m

	return m, err
}

func (n *node) readByte() (byte, error) {
	if n.empty() {
		n.reset()
		return 0, io.EOF
	}

	c := n.buf[n.readPos]
	n.readPos++
	return c, nil
}

func (n *node) readFrom(r io.Reader) (s int64, err error) {
	if n.available() == 0 {
		return 0, errBufferFull
	}

	var l int
	l, err = r.Read(n.buf[n.writePos:])

	if l < 0 {
		return 0, errInvalidWriteCount
	}

	s = int64(l)
	n.writePos += l

	return
}

func (n *node) write(p []byte) (l int, err error) {
	m := n.available()

	if m == 0 {
		return 0, errBufferFull
	}

	if m > len(p) {
		m = len(p)
	}

	l = copy(n.buf[n.writePos:], p)
	n.writePos += l
	return l, nil
}

func (n *node) writeTo(w io.Writer) (s int64, err error) {
	if nBytes := n.size(); nBytes > 0 {
		l, e := w.Write(n.buf[n.readPos : n.readPos+nBytes])
		n.readPos += l
		s = int64(l)

		if l > nBytes {
			return s, errInvalidWriteCount
		}

		if e != nil {
			return s, e
		}

		if l != nBytes {
			return s, io.ErrShortWrite
		}
	}

	n.reset()
	return s, err
}

func (n *node) writeByte(c byte) error {
	m := n.available()

	if m == 0 {
		return errBufferFull
	}

	n.buf[n.writePos] = c
	n.writePos++
	return nil
}

func (n *node) tryGrowByReSlice(s int) (int, bool) {
	if l := len(n.buf); s <= cap(n.buf)-l {
		n.buf = n.buf[:l+s]
		return l, true
	}

	return 0, false
}

func (n *node) size() int {
	return n.writePos - n.readPos
}

func (n *node) reset() {
	n.readPos = 0
	n.writePos = 0
	n.buf = n.buf[:0]
	n.prev = nil
	n.next = nil
}

func (n *node) empty() bool {
	return n.len() == 0
}

func (n *node) available() int {
	if cap(n.buf) > n.writePos {
		return cap(n.buf) - n.writePos
	}

	return 0
}

// len returns the number of bytes of the unread portion of the buffer;
// n.len() == len(n.bytes()).
func (n *node) len() int {
	return len(n.buf) - n.readPos
}

const defaultMaxCapacity = 4096

type option func(*RingLinkedList)

func WithMaxBufferCapacity(c int) option {
	return func(r *RingLinkedList) {
		r.cap = c
	}
}

func WithAutoGrow(b bool) option {
	return func(r *RingLinkedList) {
		r.autoGrow = b
	}
}

func WithGrowth(n int) option {
	return func(r *RingLinkedList) {
		r.growth = n
	}
}

type RingLinkedList struct {
	cap      int
	autoGrow bool
	growth   int
	empty    bool
	head     *node
	tail     *node
	size     int
}

func New(opts ...option) *RingLinkedList {
	r := &RingLinkedList{}

	for _, opt := range opts {
		opt(r)
	}

	if r.cap == 0 {
		r.cap = defaultMaxCapacity
	}

	n := &node{
		buf:      make([]byte, 0, r.cap),
		readPos:  0,
		writePos: 0,
	}

	r.pushBack(n)

	return r
}

func NewBuffer(buf []byte, opts ...option) *RingLinkedList {
	r := &RingLinkedList{}

	for _, opt := range opts {
		opt(r)
	}

	r.cap = len(buf)

	n := &node{
		buf:      buf,
		readPos:  0,
		writePos: 0,
	}

	r.pushBack(n)

	return r
}

func NewBufferString(s string, opts ...option) *RingLinkedList {
	r := &RingLinkedList{}

	for _, opt := range opts {
		opt(r)
	}

	n := &node{
		buf:      []byte(s),
		readPos:  0,
		writePos: 0,
	}

	r.cap = len(n.buf)

	r.pushBack(n)

	return r
}

// Read reads the next len(p) bytes from the buffer or until the
// buffer is drained. The return value n is the number of bytes
// read. If the buffer has no data to return, err is io.EOF (unless
// len(p) is zero); otherwise it is nil.
func (r *RingLinkedList) Read(p []byte) (n int, err error) {
	if r.Empty() {
		r.Reset()

		if len(p) == 0 {
			return 0, nil
		}

		return 0, io.EOF
	}

	c := r.Size()

	if c > len(p) {
		c = len(p)
	}

	var l int

	for s := r.popFront(); s != nil; s = r.popFront() {
		if l, err = s.read(p[n:]); err != nil {
			return n, err
		}

		n += l
		c -= l

		if s.empty() {
			s.reset()
			r.pushBack(s)
		} else {
			r.pushFront(s)
		}

		if c == 0 {
			break
		}
	}

	return n, nil
}

func (r *RingLinkedList) ReadByte() (byte, error) {
	if r.Empty() {
		r.Reset()
		return 0, io.EOF
	}

	var c byte
	var err error

	for n := r.popFront(); n != nil; n = r.popFront() {
		if c, err = n.readByte(); err != nil {
			return 0, err
		}

		if n.empty() {
			r.pushBack(n)
		} else {
			r.pushFront(n)
		}

		break
	}

	return c, nil
}

const minRead = 512

func (r *RingLinkedList) ReadFrom(reader io.Reader) (n int64, err error) {
	if r.Available() == 0 && !r.autoGrow {
		return 0, errBufferFull
	}

	var l int64

	for s := r.head; s != nil; s = s.next {
		if s.available() == 0 {
			continue
		}

		l, err = s.readFrom(reader)
		n += l

		if err == io.EOF {
			return n, nil
		} else if err != nil {
			return n, err
		}
	}

	c := smallBufferSize

	if r.growth > 0 {
		c = r.growth
	}

	for {
		r.grow(c)
		tail := r.tail
		l, err = tail.readFrom(reader)
		n += l

		if err == io.EOF {
			return n, nil
		} else if err != nil {
			return n, err
		} else if tail.available() != 0 {
			break
		}
	}

	return
}

func (r *RingLinkedList) Write(p []byte) (n int, err error) {
	if r.Available() == 0 && !r.autoGrow {
		return 0, errBufferFull
	}

	if _, ok := r.tryGrowByReSlice(len(p)); !ok {
		r.grow(len(p))
	}

	var l int

	for s := r.head; s != nil; s = s.next {
		available := s.available()

		if 0 >= available {
			continue
		}

		l, err = s.write(p[n:])
		n += l

		if err != nil {
			return n, err
		}

		if n == len(p) {
			return
		}
	}

	if len(p) != n && !r.autoGrow {
		return n, io.ErrShortWrite
	}

	return
}

// WriteTo writes data to w until the buffer is drained or an error occurs.
// The return value n is the number of bytes written; it always fits into
// an int, but it is int64 to match the io.WriterTo interface. Any error
// encountered during to write is also returned.
func (r *RingLinkedList) WriteTo(w io.Writer) (n int64, err error) {
	if nBytes := r.Size(); nBytes > 0 {
		var l int64

		for s := r.popFront(); s != nil; s = r.popFront() {
			l, err = s.writeTo(w)
			n += l

			if err != nil {
				return n, err
			}
		}
	}

	return n, nil
}

func (r *RingLinkedList) WriteByte(c byte) error {
	if r.Available() == 0 && !r.autoGrow {
		return errBufferFull
	}

	if _, ok := r.tryGrowByReSlice(1); !ok {
		r.grow(1)
	}

	for s := r.head; s != nil; s = s.next {
		available := s.available()

		if 0 >= available {
			continue
		}

		if err := s.writeByte(c); err != nil {
			return err
		}
		break
	}

	return nil
}

func (r *RingLinkedList) WriteString(s string) (int, error) {
	return r.Write([]byte(s))
}

func (r *RingLinkedList) pushFront(n *node) {
	if n == nil {
		return
	}

	if r.head == nil {
		n.prev = nil
		n.next = nil
		r.tail = n
	} else {
		n.next = r.head
		r.head.prev = n
	}

	r.head = n
}

func (r *RingLinkedList) pushBack(n *node) {
	if n == nil {
		return
	}

	if r.tail == nil {
		r.head = n
		n.prev = nil
	} else {
		r.tail.next = n
		n.prev = r.tail
	}

	n.next = nil
	r.tail = n
}

func (r *RingLinkedList) popFront() *node {
	if r.head == nil {
		return nil
	}

	n := r.head
	r.head = n.next

	if r.head == nil {
		r.tail = nil
	}

	n.next = nil
	return n
}

func (r *RingLinkedList) popBack() *node {
	if r.tail == nil {
		return nil
	}

	n := r.tail
	r.tail = n.prev

	if r.tail == nil {
		r.head = nil
	}

	n.prev = nil
	return n
}

func (r *RingLinkedList) Foreach(f func(*node)) {
	if r.head == nil {
		return
	}

	for n := r.head; n != nil; n = n.next {
		f(n)
	}
}

// Len returns the number of bytes of the unread portion of the
// buffer; r.Len() == len(r.Bytes()).
func (r *RingLinkedList) Len() int {
	if r.Empty() {
		return 0
	}

	l := 0

	for n := r.head; n != nil; n = n.next {
		l += n.len()
	}

	return r.cap - l
}

func (r *RingLinkedList) Available() int {
	n := 0

	for s := r.head; s != nil; s = s.next {
		n += s.available()
	}

	return n
}

func (r *RingLinkedList) Cap() int {
	return r.cap
}

func (r *RingLinkedList) Size() int {
	n := 0

	for s := r.head; s != nil; s = s.next {
		n += s.size()
	}

	return n
}

func (r *RingLinkedList) Reset() {
	for n := r.head; n != nil; n = n.next {
		n.reset()
	}
}

func (r *RingLinkedList) Empty() bool {
	return r.empty
}

func (r *RingLinkedList) tryGrowByReSlice(n int) (int, bool) {
	available := 0

	for s := r.head; s != nil; s = s.next {
		if n <= s.available() {
			s.buf = s.buf[:len(s.buf)+n]
			available += n
			n -= n
		} else {
			s.buf = s.buf[:len(s.buf)+s.available()]
			available += s.available()
			n -= s.available()
		}
	}

	if n == 0 {
		return available, true
	}

	return available, false
}

// grow grows the buffer to guarantee space for n more bytes.
// If autoGrow is false, it will not do anything.
func (r *RingLinkedList) grow(n int) {
	if !r.autoGrow {
		return
	}

	available := r.Available()

	if available > n {
		return
	}

	if n > available {
		n -= available
	}

	if _, ok := r.tryGrowByReSlice(n); ok {
		return
	}

	c := n

	if r.growth > c {
		c = r.growth
	} else if smallBufferSize > c {
		c = smallBufferSize
	}

	s := &node{
		buf:      make([]byte, n, c),
		readPos:  0,
		writePos: 0,
	}

	r.cap += c
	r.pushBack(s)
}
