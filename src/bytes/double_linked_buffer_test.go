package bytes

import (
	"fmt"
	"testing"
)

func TestDoubleLinkedBuffer_Read(t *testing.T) {
	buff := &DoubleLinkedBuffer{}

	s1 := "hello world"
	b1 := make([]byte, len(s1))
	copy(b1, s1)
	s2 := "I love you"
	b2 := make([]byte, len(s2))
	copy(b2, s2)
	buff.Append(b1)
	buff.Append(b2)
	b3 := make([]byte, 50)
	n1, _ := buff.Read(b3)
	fmt.Println(n1, string(b3))
	fmt.Println(string(buff.Bytes()))
}
