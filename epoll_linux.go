// +build linux

package epoller

import (
	"net"
	"sync"
	"syscall"

	"golang.org/x/sys/unix"
)

type epoll struct {
	fd          int
	connections map[int]net.Conn
	lock        *sync.RWMutex
	connbuf     []net.Conn
	events      []unix.EpollEvent
}

func NewPoller() (Poller, error) {
	fd, err := unix.EpollCreate1(0)
	if err != nil {
		return nil, err
	}
	return &epoll{
		fd:          fd,
		lock:        &sync.RWMutex{},
		connections: make(map[int]net.Conn),
		connbuf:     make([]net.Conn, 128, 128),
		events:      make([]unix.EpollEvent, 128, 128),
	}, nil
}

func NewPollerWithBuffer(count int) (Poller, error) {
	fd, err := unix.EpollCreate1(0)
	if err != nil {
		return nil, err
	}
	return &epoll{
		fd:          fd,
		lock:        &sync.RWMutex{},
		connections: make(map[int]net.Conn),
		connbuf:     make([]net.Conn, count, count),
		events:      make([]unix.EpollEvent, count, count),
	}, nil
}

func (e *epoll) Close() error {
	e.lock.Lock()
	defer e.lock.Unlock()

	e.connections = nil
	return unix.Close(e.fd)
}

func (e *epoll) Add(conn net.Conn) error {
	// Extract file descriptor associated with the connection
	fd := socketFD(conn)
	err := unix.EpollCtl(e.fd, syscall.EPOLL_CTL_ADD, fd, &unix.EpollEvent{Events: unix.POLLIN | unix.POLLHUP, Fd: int32(fd)})
	if err != nil {
		return err
	}
	e.lock.Lock()
	defer e.lock.Unlock()
	e.connections[fd] = conn
	return nil
}

func (e *epoll) Remove(conn net.Conn) error {
	fd := socketFD(conn)
	err := unix.EpollCtl(e.fd, syscall.EPOLL_CTL_DEL, fd, nil)
	if err != nil {
		return err
	}
	e.lock.Lock()
	defer e.lock.Unlock()
	delete(e.connections, fd)
	return nil
}

func (e *epoll) Wait(count int) ([]net.Conn, error) {
	events := make([]unix.EpollEvent, count, count)

retry:
	n, err := unix.EpollWait(e.fd, events, -1)
	if err != nil {
		if err == unix.EINTR {
			goto retry
		}
		return nil, err
	}

	var connections = make([]net.Conn, 0, n)
	e.lock.RLock()
	for i := 0; i < n; i++ {
		conn := e.connections[int(events[i].Fd)]
		if (events[i].Events & unix.POLLHUP) == unix.POLLHUP {
			conn.Close()
		}

		connections = append(connections, conn)
	}
	e.lock.RUnlock()

	return connections, nil
}

func (e *epoll) WaitWithBuffer() ([]net.Conn, error) {
retry:
	n, err := unix.EpollWait(e.fd, e.events, -1)
	if err != nil {
		if err == unix.EINTR {
			goto retry
		}
		return nil, err
	}

	var connections = e.connbuf[:0]
	e.lock.RLock()
	for i := 0; i < n; i++ {
		conn := e.connections[int(e.events[i].Fd)]
		if (e.events[i].Events & unix.POLLHUP) == unix.POLLHUP {
			conn.Close()
		}
		connections = append(connections, conn)
	}
	e.lock.RUnlock()

	return connections, nil
}

func (e *epoll) WaitChan(count int) <-chan []net.Conn {
	ch := make(chan []net.Conn)
	go func() {
		for {
			conns, err := e.Wait(count)
			if err != nil {
				close(ch)
				return
			}
			if len(conns) == 0 {
				continue
			}

			ch <- conns
		}
	}()
	return ch
}
