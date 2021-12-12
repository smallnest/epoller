package epoller

import (
	"net"
	"syscall"
)

type Poller interface {
	Add(conn net.Conn) error
	Remove(conn net.Conn) error
	Wait(count int) ([]net.Conn, error)
	WaitWithBuffer() ([]net.Conn, error)
	WaitChan(count int) <-chan []net.Conn
	Close() error
}

func socketFD(conn net.Conn) int {
	if con, ok := conn.(syscall.Conn); ok {
		raw, err := con.SyscallConn()
		if err != nil {
			return 0
		}
		sfd := 0
		raw.Control(func(fd uintptr) {
			sfd = int(fd)
		})
		return sfd
	}
	return 0
}