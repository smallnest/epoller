//go:build windows && cgo
// +build windows,cgo

package wepoll

//#cgo windows LDFLAGS: -lws2_32 -lwsock32
//#include"wepoll.h"
import "C"

import (
	"errors"
	"net"
	"sync"
	"syscall"
)

type Epoll struct {
	fd C.uintptr_t

	connBufferSize int
	lock           *sync.RWMutex
	conns          map[int]net.Conn
	connbuf        []net.Conn
}

// NewPoller creates a new epoll poller.
func NewPoller(connBufferSize int) (*Epoll, error) {
	fd := C.epoll_create1(0)

	if fd == 0 {
		return nil, errors.New("epoll_create1 error")
	}
	return &Epoll{
		fd:             fd,
		connBufferSize: connBufferSize,
		lock:           &sync.RWMutex{},
		conns:          make(map[int]net.Conn),
		connbuf:        make([]net.Conn, connBufferSize, connBufferSize),
	}, nil
}

func (e *Epoll) Close(closeConns bool) error {
	e.lock.Lock()
	defer e.lock.Unlock()

	if closeConns {
		for _, conn := range e.conns {
			conn.Close()
		}
	}

	e.conns = nil
	e.connbuf = e.connbuf[:0]

	i := C.epoll_close(e.fd)
	if i == 0 {
		return nil
	} else {
		return errors.New(" an error occurred on epoll.close ")
	}
}

// Add adds a connection to the poller.
func (e *Epoll) Add(conn net.Conn) error {
	// Extract file descriptor associated with the connection
	fd := C.SOCKET(socketFDAsUint(conn))
	var ev C.epoll_event
	ev = C.set_epoll_event(C.EPOLLIN|C.EPOLLHUP, C.SOCKET(fd))
	e.lock.Lock()
	defer e.lock.Unlock()
	err := C.epoll_ctl(e.fd, C.EPOLL_CTL_ADD, C.SOCKET(fd), &ev)
	if err == -1 {
		return errors.New("C.EPOLL_CTL_ADD error ")
	}
	e.conns[int(fd)] = conn
	return nil
}

// Remove removes a connection from the poller.
func (e *Epoll) Remove(conn net.Conn) error {
	fd := C.SOCKET(socketFDAsUint(conn))
	var ev C.epoll_event
	err := C.epoll_ctl(e.fd, C.EPOLL_CTL_DEL, C.SOCKET(fd), &ev)
	if err == -1 {
		return errors.New("C.EPOLL_CTL_DEL error ")
	}
	e.lock.Lock()
	defer e.lock.Unlock()
	delete(e.conns, int(fd))
	return nil
}

// Wait waits for events on the connections the poller is managing.
func (e *Epoll) Wait(count int) ([]net.Conn, error) {
	events := make([]C.epoll_event, count)

	n := C.epoll_wait(e.fd, &events[0], C.int(count), -1)
	if n == -1 {
		return nil, errors.New("C.epoll_wait error")
	}

	var conns []net.Conn
	if e.connBufferSize > 0 {
		conns = e.connbuf[:0]
	} else {
		conns = make([]net.Conn, 0, n)
	}

	e.lock.RLock()
	for i := 0; i < int(n); i++ {
		fd := C.get_epoll_event(events[i])
		conn := e.conns[int(fd)]
		if conn != nil {
			conns = append(conns, conn)
		}

	}
	e.lock.RUnlock()

	return conns, nil
}

func socketFDAsUint(conn net.Conn) uint64 {
	if con, ok := conn.(syscall.Conn); ok {
		raw, err := con.SyscallConn()
		if err != nil {
			return 0
		}
		sfd := uint64(0)
		raw.Control(func(fd uintptr) {
			sfd = uint64(fd)
		})
		return sfd
	}
	return 0
}

func (e *Epoll) Size() int {
	e.lock.RLock()
	defer e.lock.RUnlock()

	return len(e.conns)
}
