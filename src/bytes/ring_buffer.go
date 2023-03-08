package bytes

type RingBuffer struct {
	cap      uint32
	writePos uint32
	readPos  uint32
}
