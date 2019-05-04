package epoller

import (
	"net"
	"reflect"
)

type Poller interface {
	Add(conn net.Conn) error
	Remove(conn net.Conn) error
	Wait(count int) ([]net.Conn, error)
	WaitWithBuffer() ([]net.Conn, error)
	Close() error
}

func socketFD(conn net.Conn) int {
	tcpConn := reflect.Indirect(reflect.ValueOf(conn)).FieldByName("conn")
	fdVal := tcpConn.FieldByName("fd")
	pfdVal := reflect.Indirect(fdVal).FieldByName("pfd")

	return int(pfdVal.FieldByName("Sysfd").Int())
}

func socketFDAsUint(conn net.Conn) uint64 {
	tcpConn := reflect.Indirect(reflect.ValueOf(conn)).FieldByName("conn")
	fdVal := tcpConn.FieldByName("fd")
	pfdVal := reflect.Indirect(fdVal).FieldByName("pfd")

	return pfdVal.FieldByName("Sysfd").Uint()
}
