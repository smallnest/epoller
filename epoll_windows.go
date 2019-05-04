// +build windows,cgo

package epoller

//#cgo windows LDFLAGS: -lws2_32 -lwsock32
//#include"wepoll.h"
import "C"
import (
	"errors"

	"net"
	"reflect"
	"sync"
	"unsafe"
)

type epoll struct {
	fd          C.HANDLE
	connections map[int]net.Conn
	lock        *sync.RWMutex
	connbuf     []net.Conn
	events      []C.epoll_event
}

func NewPoller() (Poller, error) {
	fd := C.epoll_create1(0)

	if unsafe.Pointer(fd) == C.NULL {
		return nil, errors.New("epoll_create1 error")
	}
	return &epoll{
		fd:          fd,
		lock:        &sync.RWMutex{},
		connections: make(map[int]net.Conn),
		connbuf:     make([]net.Conn, 128, 128),
		events:      make([]C.epoll_event, 128, 128),
	}, nil
}

func NewPollerWithBuffer(count int) (Poller, error) {
	fd := C.epoll_create1(0)
	if unsafe.Pointer(fd) == C.NULL {
		return nil, errors.New("epoll_create1 error")
	}
	return &epoll{
		fd:          fd,
		lock:        &sync.RWMutex{},
		connections: make(map[int]net.Conn),
		connbuf:     make([]net.Conn, count, count),
		events:      make([]C.epoll_event, count, count),
	}, nil
}

func (e *epoll) Close() error {
	e.connections = nil
	i := C.epoll_close(e.fd)
	if i == 0 {
		return nil
	} else {

		return errors.New(" an error occurred on epoll.close ")
	}
}

func (e *epoll) Add(conn net.Conn) error {
	// Extract file descriptor associated with the connection

	tcpConn := reflect.Indirect(reflect.ValueOf(conn)).FieldByName("conn")
	fdVal := tcpConn.FieldByName("fd")
	pfdVal := reflect.Indirect(fdVal).FieldByName("pfd")

	fd := C.SOCKET(pfdVal.FieldByName("Sysfd").Uint())
	var ev C.epoll_event
	ev = C.set_epoll_event(C.EPOLLIN|C.EPOLLHUP, C.SOCKET(fd))

	err := C.epoll_ctl(e.fd, C.EPOLL_CTL_ADD, C.SOCKET(fd), &ev)
	if err == -1 {
		return errors.New("C.EPOLL_CTL_ADD error ")
	}
	e.lock.Lock()
	defer e.lock.Unlock()
	e.connections[int(fd)] = conn
	return nil
}

func (e *epoll) Remove(conn net.Conn) error {
	defer conn.Close()
	tcpConn := reflect.Indirect(reflect.ValueOf(conn)).FieldByName("conn")
	fdVal := tcpConn.FieldByName("fd")
	pfdVal := reflect.Indirect(fdVal).FieldByName("pfd")

	fd := C.SOCKET(pfdVal.FieldByName("Sysfd").Uint())
	var ev C.epoll_event
	err := C.epoll_ctl(e.fd, C.EPOLL_CTL_DEL, C.SOCKET(fd), &ev)
	if err == -1 {
		return errors.New("C.EPOLL_CTL_DEL error ")
	}
	e.lock.Lock()
	defer e.lock.Unlock()
	delete(e.connections, int(fd))
	return nil
}

func (e *epoll) Wait(count int) ([]net.Conn, error) {
	events := make([]C.epoll_event, count, count)

	n := C.epoll_wait(e.fd, &events[0], C.int(count), -1)
	if n == -1 {
		return nil, errors.New("C.epoll_wait error")
	}

	var connections = make([]net.Conn, 0, n)
	e.lock.RLock()
	for i := 0; i < int(n); i++ {
		fd := C.get_epoll_event(events[i])
		//fmt.Println("get_epoll_event i:,fd: ", i, fd)
		conn := e.connections[int(fd)]
		connections = append(connections, conn)
	}
	e.lock.RUnlock()

	return connections, nil
}

func (e *epoll) WaitWithBuffer() ([]net.Conn, error) {
	n := C.epoll_wait(e.fd, &e.events[0], 128, -1)
	if n == -1 {
		return nil, errors.New("WaitWithBuffer err")
	}

	var connections = e.connbuf[:0]
	e.lock.RLock()
	for i := 0; i < int(n); i++ {
		fd := C.get_epoll_event(e.events[i])
		conn := e.connections[int(fd)]

		connections = append(connections, conn)
	}
	e.lock.RUnlock()

	return connections, nil
}
