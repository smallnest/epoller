// +build darwin netbsd freebsd openbsd dragonfly

package epoller

import (
	"net"
	"sync"
	"syscall"
)

type epoll struct {
	fd          int
	ts          syscall.Timespec
	changes     []syscall.Kevent_t
	connections map[int]net.Conn
	mu          *sync.RWMutex
}

func NewPoller() (Poller, error) {
	p, err := syscall.Kqueue()
	if err != nil {
		panic(err)
	}
	_, err = syscall.Kevent(p, []syscall.Kevent_t{{
		Ident:  0,
		Filter: syscall.EVFILT_USER,
		Flags:  syscall.EV_ADD | syscall.EV_CLEAR,
	}}, nil, nil)
	if err != nil {
		panic(err)
	}

	return &epoll{
		fd:          p,
		ts:          syscall.NsecToTimespec(0),
		mu:          &sync.RWMutex{},
		connections: make(map[int]net.Conn),
	}, nil
}

func (e *epoll) Close() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.connections = nil
	e.changes = nil
	return syscall.Close(e.fd)
}

func (e *epoll) Add(conn net.Conn) error {
	fd := socketFD(conn)

	e.mu.Lock()
	defer e.mu.Unlock()

	e.changes = append(e.changes,
		syscall.Kevent_t{
			Ident: uint64(fd), Flags: syscall.EV_ADD | syscall.EV_EOF, Filter: syscall.EVFILT_READ,
		},
	)

	e.connections[fd] = conn
	return nil
}

func (e *epoll) Remove(conn net.Conn) error {
	fd := socketFD(conn)

	e.mu.Lock()
	defer e.mu.Unlock()

	if len(e.changes) <= 1 {
		e.changes = nil
	} else {
		changes := make([]syscall.Kevent_t, 0, len(e.changes)-1)
		ident := uint64(fd)
		for _, ke := range e.changes {
			if ke.Ident != ident {
				changes = append(changes)
			}
		}
		e.changes = changes
	}

	delete(e.connections, fd)
	return nil
}

func (e *epoll) Wait(count int) ([]net.Conn, error) {
	events := make([]syscall.Kevent_t, count)

	e.mu.RLock()
	changes := e.changes
	e.mu.RUnlock()

	n, err := syscall.Kevent(e.fd, changes, events, &e.ts)
	if err != nil && err != syscall.EINTR {
		return nil, err
	}

	e.mu.RLock()
	defer e.mu.RUnlock()
	var connections []net.Conn
	for i := 0; i < n; i++ {
		conn := e.connections[int(events[i].Ident)]
		connections = append(connections, conn)
	}
	return connections, nil
}
