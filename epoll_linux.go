//go:build linux
// +build linux

package epoller

import (
	"errors"
	"net"
	"sync"
	"syscall"

	"golang.org/x/sys/unix"
)

var _ Poller = (*Epoll)(nil)

// Epoll is a epoll based poller.
type Epoll struct {
	fd int

	connBufferSize int
	lock           *sync.RWMutex
	conns          map[int]net.Conn
	connbuf        []net.Conn
}

// NewPoller creates a new epoll poller.
func NewPoller(connBufferSize int) (*Epoll, error) {
	return newPollerWithBuffer(connBufferSize)
}

// newPollerWithBuffer creates a new epoll poller with a buffer.
func newPollerWithBuffer(count int) (*Epoll, error) {
	fd, err := unix.EpollCreate1(0)
	if err != nil {
		return nil, err
	}
	return &Epoll{
		fd:             fd,
		connBufferSize: count,
		lock:           &sync.RWMutex{},
		conns:          make(map[int]net.Conn),
		connbuf:        make([]net.Conn, count, count),
	}, nil
}

// Close closes the poller. If closeConns is true, it will close all the connections.
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

	return unix.Close(e.fd)
}

// Add adds a connection to the poller.
func (e *Epoll) Add(conn net.Conn) error {
	conn = newConnImpl(conn)
	fd := socketFD(conn)
	if e := syscall.SetNonblock(int(fd), true); e != nil {
		return errors.New("udev: unix.SetNonblock failed")
	}

	err := unix.EpollCtl(e.fd, syscall.EPOLL_CTL_ADD, fd, &unix.EpollEvent{Events: unix.POLLIN | unix.POLLHUP, Fd: int32(fd)})
	if err != nil {
		return err
	}

	e.lock.Lock()
	defer e.lock.Unlock()

	e.conns[fd] = conn
	return nil
}

// Remove removes a connection from the poller.
func (e *Epoll) Remove(conn net.Conn) error {
	fd := socketFD(conn)
	err := unix.EpollCtl(e.fd, syscall.EPOLL_CTL_DEL, fd, nil)
	if err != nil {
		return err
	}

	e.lock.Lock()
	defer e.lock.Unlock()
	delete(e.conns, fd)

	return nil
}

// Wait waits for at most count events and returns the connections.
func (e *Epoll) Wait(count int) ([]net.Conn, error) {
	events := make([]unix.EpollEvent, count, count)

retry:
	n, err := unix.EpollWait(e.fd, events, -1)
	if err != nil {
		if err == unix.EINTR {
			goto retry
		}
		return nil, err
	}

	var conns []net.Conn
	if e.connBufferSize == 0 {
		conns = make([]net.Conn, 0, n)
	} else {
		conns = e.connbuf[:0]
	}
	e.lock.RLock()
	for i := 0; i < n; i++ {
		conn := e.conns[int(events[i].Fd)]
		if conn != nil {
			// issue #11: don't close the connection here because maybe data needs to drain
			//
			// if (events[i].Events & unix.POLLHUP) == unix.POLLHUP {
			// 	conn.Close()
			// }

			conns = append(conns, conn)
		}
	}
	e.lock.RUnlock()

	return conns, nil
}

func (e *Epoll) Size() int {
	e.lock.RLock()
	defer e.lock.RUnlock()
	return len(e.conns)
}
