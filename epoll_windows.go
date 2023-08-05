//go:build windows && cgo
// +build windows,cgo

package epoller

import (
	"github.com/smallnest/epoller/wepoll"
)

var _ Poller = (*Epoll)(nil)

type Epoll = wepoll.Epoll

// NewPoller creates a new epoll poller.
func NewPoller() (*Epoll, error) {
	return wepoll.NewPoller()
}

// NewPollerWithBuffer creates a new epoll poller with a buffer.
func NewPollerWithBuffer(count int) (*Epoll, error) {
	return wepoll.NewPollerWithBuffer(count)
}
