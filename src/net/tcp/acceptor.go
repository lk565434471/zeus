package tcp

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"time"
)

type ListenerConfig struct {
	Protocol     string
	Timeout      time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	Address      string
	opts         []ListenerOption
}

type ListenerOption interface {
	apply(*Acceptor)
}

type timeoutOption time.Duration

func (o timeoutOption) apply(a *Acceptor) {
	a.Config.Timeout = time.Duration(o)
}

type addressOption string

func (o addressOption) apply(a *Acceptor) {
	a.Config.Address = string(o)
}

type Acceptor struct {
	listener             net.Listener
	connectionMiddleware *ConnectionHandlerChain
	Config               ListenerConfig
	opts                 []ListenerOption
}

func NewTCPAcceptor(config ListenerConfig, opts ...ConnectionMiddleware) (*Acceptor, error) {
	return &Acceptor{
		Config:               config,
		opts:                 config.opts,
		connectionMiddleware: NewConnectionHandlerChain(opts...),
	}, nil
}

func (a *Acceptor) ListenAndServe() {
	if err := a.Listen(); err != nil {
		return
	}

	defer a.Close()
	fmt.Printf("start to listen on: %s\n", a.Config.Address)

	for {
		conn, err := a.Accept()

		if err != nil {
			// the listener has closed
			if errors.Is(err, net.ErrClosed) {
				// log a record for listener error
				break
			}

			conn.Close()
			// log a record for connection error
			continue
		}

		// create a goroutine to handle new connection
		go a.handleNewConnection(conn)
	}
}

func (a *Acceptor) Listen() error {
	var ln net.Listener
	var err error

	if a.Config.Protocol == "tcp" {
		ln, err = net.Listen("tcp", a.Config.Address)
	} else if a.Config.Protocol == "tls" {
		ln, err = tls.Listen("tls", a.Config.Address, nil)
	} else {
		return fmt.Errorf("invalid protocol: %s", a.Config.Protocol)
	}

	if err != nil {
		return err
	}

	a.listener = ln

	for _, opt := range a.opts {
		opt.apply(a)
	}

	return nil
}

func (a *Acceptor) Accept() (net.Conn, error) {
	return a.listener.Accept()
}

func (a *Acceptor) Addr() net.Addr {
	return a.listener.Addr()
}

func (a *Acceptor) Close() error {
	return a.listener.Close()
}

func (a *Acceptor) handleNewConnection(conn net.Conn) {
	// Here, we will run a series of connection middlewares that you have registered.
	// For example: MaxConnectionsMiddleware
	a.connectionMiddleware.Execute(a, conn)
}
