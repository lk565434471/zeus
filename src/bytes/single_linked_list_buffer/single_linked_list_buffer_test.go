package single_linked_list_buffer

import (
	"fmt"
	"testing"
)

func TestLinkedListBuffer_Write(t *testing.T) {
	s1 := "Hello World"

	b := &LinkedListBuffer{}
	fmt.Println(b.Write([]byte(s1)))
	fmt.Println(string(b.Bytes()))
	p1 := make([]byte, len(s1)-2)
	fmt.Println(b.Read(p1))
	fmt.Println(string(p1))

	fmt.Println("===============================")

	if p2, err := b.Peek(9); err != nil {
		fmt.Println(string(p2), err)
	} else {
		fmt.Println(string(p2))
	}

	b.Discard(1)
	fmt.Println(string(b.Bytes()))
}
