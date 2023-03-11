package ring_linked_list

import (
	"fmt"
	"os"
	"testing"
)

func TestRingLinkedList_Read(t *testing.T) {
	r := New(
		WithMaxBufferCapacity(20),
		WithAutoGrow(true),
		WithGrowth(20),
	)

	r.Write([]byte("hello world"))
	r.Write([]byte("111111111"))
	r.Write([]byte("say"))

	p1 := make([]byte, 20)
	fmt.Println(r.Read(p1))
	fmt.Println(string(p1))
	p2 := make([]byte, 20)
	fmt.Println(r.Read(p2))
	fmt.Println(string(p2))
}

func TestRingLinkedList_Write(t *testing.T) {
	r := New(
		WithMaxBufferCapacity(20),
		WithAutoGrow(true),
	)
	fmt.Println(r.Cap())
	fmt.Println(r.Write([]byte("hello world")))
	fmt.Println(r.Write([]byte("111111111")))
	fmt.Println(r.Write([]byte("say")))
	fmt.Println(r.Cap())
}

func TestRingLinkedList_WriteTo(t *testing.T) {
	r := New(
		WithMaxBufferCapacity(20),
		WithAutoGrow(true),
	)
	r.Write([]byte("hello world"))
	r.Write([]byte("111111111"))
	r.Write([]byte("say"))
	r.WriteTo(os.Stdout)
}
