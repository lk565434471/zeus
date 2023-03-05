package conn

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync"
	"time"
)

var (
	errInvalidSession   = errors.New("invalid session")
	errExistsSession    = errors.New("exists session")
	errNotExistsSession = errors.New("not exists session")
)

type Session struct {
	id   uint64
	conn net.Conn
	ctx  context.Context
}

func NewSession(id uint64, conn net.Conn, ctx context.Context) *Session {
	return &Session{
		id:   id,
		conn: conn,
		ctx:  ctx,
	}
}

func (s *Session) GetId() uint64 {
	return s.id
}

func (s *Session) GetFd() (uintptr, error) {
	sock, ok := s.conn.(*net.TCPConn)

	if !ok {
		return ^(uintptr(0)), nil
	}

	f, err := sock.File()

	if err != nil {
		return ^(uintptr(0)), err
	}

	return f.Fd(), nil
}

func (s *Session) Read(b []byte) (n int, err error) {
	n, err = s.conn.Read(b)

	return n, err
}

func (s *Session) Write(b []byte) (n int, err error) {
	n, err = s.conn.Write(b)

	return n, err
}

func (s *Session) Close() error {
	return s.conn.Close()
}

func (s *Session) LocalAddr() net.Addr {
	return s.conn.LocalAddr()
}

func (s *Session) RemoteAddr() net.Addr {
	return s.conn.RemoteAddr()
}

func (s *Session) SetDeadline(t time.Time) error {
	return s.conn.SetDeadline(t)
}

func (s *Session) SetReadDeadline(t time.Time) error {
	return s.conn.SetReadDeadline(t)
}

func (s *Session) SetWriteDeadline(t time.Time) error {
	return s.conn.SetWriteDeadline(t)
}

type SessionManager struct {
	ctx            context.Context
	sessions       map[uint64]*Session
	sessionsByConn map[net.Conn]*Session
	enableMutex    bool
	mutex          sync.RWMutex
	sessionPool    sync.Pool
}

func NewSessionManager(ctx context.Context) *SessionManager {
	return &SessionManager{
		ctx:            ctx,
		sessions:       make(map[uint64]*Session),
		sessionsByConn: make(map[net.Conn]*Session),
		sessionPool: sync.Pool{
			New: func() any {
				return &Session{}
			},
		},
	}
}

func (m *SessionManager) NewSession(ctx context.Context, id uint64, conn net.Conn) *Session {
	m.lock()
	defer m.unlock()

	s := m.sessionPool.Get().(*Session)

	if s == nil {
		s = NewSession(id, conn, ctx)
	}

	return s
}

func (m *SessionManager) AddSession(s *Session) error {
	if s == nil {
		return errInvalidSession
	}

	if m.ExistsSessionById(s.id) || m.ExistsSessionByConn(s.conn) {
		return errExistsSession
	}

	if err := m.addSessionById(s); err != nil {
		return err
	}

	if err := m.addSessionByConn(s); err != nil {
		return err
	}

	return nil
}

func (m *SessionManager) RemoveSession(s *Session) (bool, error) {
	if s == nil {
		return false, errInvalidSession
	}

	if !m.ExistsSessionById(s.id) && !m.ExistsSessionByConn(s.conn) {
		return false, errNotExistsSession
	}

	if ok, err := m.removeSessionById(s.id); !ok || err != nil {
		return ok, err
	}

	if ok, err := m.removeSessionByConn(s.conn); !ok || err != nil {
		return ok, err
	}

	m.sessionPool.Put(s)

	return true, nil
}

func (m *SessionManager) removeSessionById(id uint64) (bool, error) {
	if !m.ExistsSessionById(id) {
		return false, errNotExistsSession
	}

	m.lock()
	defer m.unlock()
	delete(m.sessions, id)
	fmt.Printf("removeSessionById: %d", id)

	return true, nil
}

func (m *SessionManager) removeSessionByConn(conn net.Conn) (bool, error) {
	if !m.ExistsSessionByConn(conn) {
		return false, errNotExistsSession
	}

	m.lock()
	defer m.unlock()
	delete(m.sessionsByConn, conn)
	fmt.Printf("removeSessionByConn: %v", conn)

	return true, nil
}

func (m *SessionManager) ExistsSessionById(id uint64) bool {
	m.lock()
	defer m.unlock()

	if _, ok := m.sessions[id]; ok {
		return true
	}

	return false
}

func (m *SessionManager) ExistsSessionByConn(conn net.Conn) bool {
	m.lock()
	defer m.unlock()

	if _, ok := m.sessionsByConn[conn]; ok {
		return true
	}

	return false
}

func (m *SessionManager) addSessionById(s *Session) error {
	m.lock()
	defer m.unlock()

	if s == nil {
		return errInvalidSession
	}

	if _, ok := m.sessions[s.id]; ok {
		return errExistsSession
	}

	m.sessions[s.id] = s
	fmt.Printf("addSessionById: %d", s.id)

	return nil
}

func (m *SessionManager) addSessionByConn(s *Session) error {
	m.lock()
	defer m.unlock()

	if s == nil {
		return errInvalidSession
	}

	if _, ok := m.sessionsByConn[s.conn]; ok {
		return errExistsSession
	}

	m.sessionsByConn[s.conn] = s
	fmt.Printf("addSessionByConn: %v", s.conn)

	return nil
}

func (m *SessionManager) GetSessionById(id uint64) *Session {
	m.lock()
	defer m.unlock()

	if s, ok := m.sessions[id]; ok {
		return s
	}

	return nil
}

func (m *SessionManager) GetSessionByConn(conn net.Conn) *Session {
	m.lock()
	defer m.unlock()

	if s, ok := m.sessionsByConn[conn]; ok {
		return s
	}

	return nil
}

func (m *SessionManager) lock() {
	if !m.enableMutex {
		return
	}

	m.mutex.Lock()
}

func (m *SessionManager) unlock() {
	if !m.enableMutex {
		return
	}

	m.mutex.Unlock()
}
