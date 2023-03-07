package net

import (
	"context"
	"fmt"
	"net"
	"time"
)

type ServingFunc func(*Listener) error
type NewConnectionFunc func(*Listener, *Conn) error
type ReceiveDataFunc func(c *Conn) error
type TickerFunc func(*Listener) (delay time.Duration, res error)
type ClosedFunc func(*Conn, error) error

type ListenerOption func(*Listener) error

func WithServingFunc(f ServingFunc) ListenerOption {
	return func(l *Listener) error {
		l.events.Serving = f
		return nil
	}
}

func WithNewConnectionFunc(f NewConnectionFunc) ListenerOption {
	return func(l *Listener) error {
		l.events.NewConnection = f
		return nil
	}
}

func WithReceiveDataFunc(f ReceiveDataFunc) ListenerOption {
	return func(l *Listener) error {
		l.events.ReceiveData = f
		return nil
	}
}

func WithTickerFunc(f TickerFunc) ListenerOption {
	return func(l *Listener) error {
		l.events.Ticker = f
		return nil
	}
}

func WithConnectionClosedFunc(f ClosedFunc) ListenerOption {
	return func(l *Listener) error {
		l.events.ConnectionClosed = f
		return nil
	}
}

type Events struct {
	Serving          ServingFunc
	NewConnection    NewConnectionFunc
	ReceiveData      ReceiveDataFunc
	Ticker           TickerFunc
	ConnectionClosed ClosedFunc
}

type Listener struct {
	ctx         context.Context
	ln          net.Listener
	connections map[*Conn]bool
	events      Events
}

func NewListener(l net.Listener, opts ...ListenerOption) *Listener {
	ln := &Listener{
		ln: l,
	}

	for _, opt := range opts {
		if err := opt(ln); err != nil {
			fmt.Println(err)
		}
	}

	return ln
}

func NewDefaultListener(network, address string) (*Listener, error) {
	defaultOpts := []ListenerOption{
		WithServingFunc(serving),
		WithNewConnectionFunc(newConnection),
		WithReceiveDataFunc(receiveData),
		WithConnectionClosedFunc(connectionClosed),
	}

	return Serve(network, address, defaultOpts...)
}

func Serve(network, address string, opts ...ListenerOption) (*Listener, error) {
	ln, err := net.Listen(network, address)

	if err != nil {
		return nil, err
	}

	listener := NewListener(ln, opts...)
	initDefaultEventHandlers(listener)
	listener.events.Serving(listener)

	go accept(listener)

	return listener, nil
}

func (l *Listener) SetContext(ctx context.Context) {
	l.ctx = ctx
}

func (l *Listener) Context() context.Context {
	return l.ctx
}

func accept(l *Listener) {
	for {
		conn, err := l.ln.Accept()

		if err != nil {
			return
		}

		c := &Conn{
			conn:       conn,
			ln:         l,
			localAddr:  l.ln.Addr(),
			remoteAddr: conn.RemoteAddr(),
		}

		if err = c.ln.events.NewConnection(l, c); err != nil {
			conn.Close()
			l.events.ConnectionClosed(c, err)
			continue
		}

		l.connections[c] = true
		go handleNewConnection(c)
	}
}

func handleNewConnection(c *Conn) {
	for {
	}
}

func serving(l *Listener) error {
	return nil
}

func newConnection(l *Listener, c *Conn) error {
	return nil
}

func receiveData(c *Conn) error {
	return nil
}

func ticker(*Listener) (delay time.Duration, res error) {
	delay, res = 0, nil
	return
}

func connectionClosed(*Conn, error) error {
	return nil
}

func initDefaultEventHandlers(l *Listener) {
	if l.events.Serving == nil {
		l.events.Serving = serving
	}

	if l.events.NewConnection == nil {
		l.events.NewConnection = newConnection
	}

	if l.events.ReceiveData == nil {

	}

	if l.events.ConnectionClosed == nil {

	}
}

type ConnOption func(*Conn) error

func WithTCPKeepAlive(keepAlive time.Duration) ConnOption {
	return func(conn *Conn) error {
		if keepAlive > 0 {
			if c, ok := conn.conn.(*net.TCPConn); ok {
				c.SetKeepAlive(true)
				c.SetKeepAlivePeriod(keepAlive)
			}
		}

		return nil
	}
}

type Conn struct {
	ctx        context.Context
	conn       net.Conn
	ln         *Listener
	localAddr  net.Addr
	remoteAddr net.Addr
}
