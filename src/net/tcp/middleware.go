package tcp

import (
	"fmt"
	"net"
)

type ConnectionMiddleware interface {
	HandleConnection(error, net.Listener, net.Conn, func(error))
}

type ConnectionHandlerChain struct {
	handlers []ConnectionMiddleware
}

func NewConnectionHandlerChain(handlers ...ConnectionMiddleware) *ConnectionHandlerChain {
	return &ConnectionHandlerChain{
		handlers: handlers,
	}
}

func (c *ConnectionHandlerChain) Use(h ConnectionMiddleware) *ConnectionHandlerChain {
	c.handlers = append(c.handlers, h)

	return c
}

func (c *ConnectionHandlerChain) WithConnectionHandlers(handlers ...ConnectionMiddleware) *ConnectionHandlerChain {
	c.handlers = append(c.handlers, handlers...)

	return c
}

func (c *ConnectionHandlerChain) Execute(ln net.Listener, conn net.Conn) {
	c.execute(0, ln, conn, nil)
}

func (c *ConnectionHandlerChain) execute(index int, ln net.Listener, conn net.Conn, err error) {
	if index < len(c.handlers) {
		handler := c.handlers[index]
		next := func(e error) {
			c.execute(index+1, ln, conn, e)
		}

		handler.HandleConnection(err, ln, conn, next)
	}
}

type MaxConnectionsMiddleware struct {
	count    int
	maxCount int
}

func (m *MaxConnectionsMiddleware) HandleConnection(err error, ln net.Listener, conn net.Conn, next func(error)) {
	if err != nil {
		return
	}

	if m.count >= m.maxCount {
		fmt.Printf("reach max connections limit: %d, %d\n", m.count, m.maxCount)
		conn.Close()
		return
	}

	m.count++
	fmt.Printf("add connection num: %d, %d\n", m.count, m.maxCount)
	next(nil)
}
