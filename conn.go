package epoller

import (
	"net"
)

// newConnImpl returns a net.Conn with GetFD() method.
func newConnImpl(in net.Conn) net.Conn {
	if _, ok := in.(connImpl); ok {
		return in
	}

	return connImpl{
		Conn: in,
		fd:   socketFD(in),
	}
}

type connImpl struct {
	net.Conn
	fd int
}

func (c connImpl) GetFD() int {
	return c.fd
}
