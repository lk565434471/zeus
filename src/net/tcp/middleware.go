package tcp

import (
	"fmt"
	"net"
)

type connectionMiddlewares interface {
	HandleConnection(error, net.Listener, net.Conn, func(error))
}

type connectionMiddlewaresChain struct {
	handlers []connectionMiddlewares
}

func NewConnectionHandlerChain(handlers ...connectionMiddlewares) *connectionMiddlewaresChain {
	return &connectionMiddlewaresChain{
		handlers: handlers,
	}
}

func (c *connectionMiddlewaresChain) Use(h connectionMiddlewares) *connectionMiddlewaresChain {
	c.handlers = append(c.handlers, h)

	return c
}

func (c *connectionMiddlewaresChain) WithConnectionHandlers(handlers ...connectionMiddlewares) *connectionMiddlewaresChain {
	c.handlers = append(c.handlers, handlers...)

	return c
}

func (c *connectionMiddlewaresChain) Execute(ln net.Listener, conn net.Conn) {
	c.execute(0, ln, conn, nil)
}

func (c *connectionMiddlewaresChain) execute(index int, ln net.Listener, conn net.Conn, err error) {
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
