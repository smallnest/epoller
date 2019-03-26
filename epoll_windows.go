// +build windows

package epoller

import (
	"net"
	"sync"

	"golang.org/x/sys/unix"
)

const _DWORD_MAX = 0xffffffff
const _INVALID_HANDLE_VALUE = ^uintptr(0)

var iocpOnce sync.Once

var iocphandle uintptr = _INVALID_HANDLE_VALUE // completion port io handle

func netpollinit() {
	iocphandle = stdcall4(_CreateIoCompletionPort, _INVALID_HANDLE_VALUE, 0, 0, _DWORD_MAX)
	if iocphandle == 0 {
		println("runtime: CreateIoCompletionPort failed (errno=", getlasterror(), ")")
		throw("runtime: netpollinit failed")
	}
}

type epoll struct {
	fd          int
	connections map[int]net.Conn
	lock        *sync.RWMutex
}

func NewPoller() (Poller, error) {
	iocpOnce.Do(netpollinit)

	return &epoll{
		fd:          fd,
		lock:        &sync.RWMutex{},
		connections: make(map[int]net.Conn),
	}, nil
}

func (e *epoll) Close() error {
	e.connections = nil
	return unix.Close(e.fd)
}

func (e *epoll) Add(conn net.Conn) error {
	fd := socketFD(conn)

}

func (e *epoll) Remove(conn net.Conn) error {
	fd := socketFD(conn)

	return nil
}

func (e *epoll) Wait() ([]net.Conn, error) {

	return nil, nil
}
