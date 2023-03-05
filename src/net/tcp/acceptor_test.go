package tcp

import (
	"fmt"
	"net"
	"testing"
	"time"
)

func TestAcceptor_Accept(t *testing.T) {
	a, err := NewTCPAcceptor(ListenerConfig{
		Protocol: "tcp",
		Address:  ":11001",
	})

	if err != nil {
		t.Fatalf("TestAcceptor_Accept error: %s", err.Error())
	}

	go a.ListenAndServe()
}

func TestAcceptor_ConnectionsMiddleware(t *testing.T) {
	a, err := NewTCPAcceptor(ListenerConfig{
		Protocol: "tcp",
		Address:  ":11002",
	}, &MaxConnectionsMiddleware{
		maxCount: 5,
	})

	if err != nil {
		t.Fatalf("TestAcceptor_ConnectionsMiddleware error: %s", err.Error())
	}

	go a.ListenAndServe()
	time.Sleep(time.Second * time.Duration(2))
	createTCPConnection("127.0.0.1:11002", 10)
}

func createTCPConnection(address string, count int) {
	for i := 0; i < count; i++ {
		go tcpConnect(address)
		time.Sleep(time.Second)
	}
}

func tcpConnect(address string) {
	conn, err := net.Dial("tcp", address)

	if err != nil {
		fmt.Println("Dial error:", err)
		return
	}

	defer func() {
		time.Sleep(time.Second * time.Duration(10))
		conn.Close()
	}()
}
