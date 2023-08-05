# epoller
epoll implementation for connections in Linux, MacOS and windows.

[![License](https://img.shields.io/:license-MIT-blue.svg)](https://opensource.org/licenses/MIT) [![GoDoc](https://godoc.org/github.com/smallnest/epoller?status.png)](http://godoc.org/github.com/smallnest/epoller)  [![github actions](https://github.com/smallnest/epoller/actions)](https://github.com/smallnest/epoller/actions/workflows/go.yml/badge.svg) [![Go Report Card](https://goreportcard.com/badge/github.com/smallnest/epoller)](https://goreportcard.com/report/github.com/smallnest/epoller) [![coveralls](https://coveralls.io/repos/smallnest/epoller/badge.svg?branch=master&service=github)](https://coveralls.io/github/smallnest/epoller?branch=master) 


Its target is implementing a simple epoll for nework connections, so you should see it only contains few methods about net.Conn:

```go
type Poller interface {
    Add(conn net.Conn) error
    Remove(conn net.Conn) error
    Wait(count int) ([]net.Conn, error)
    WaitWithBuffer() ([]net.Conn, error)
    WaitChan(count, chanBuffer int) <-chan []net.Conn
    Close(closeConns bool) error
}

```

Welcome any PRs for windows IOCompletePort.

Inspired by [1m-go-websockets](https://github.com/eranyanay/1m-go-websockets).

see [contributors](https://github.com/smallnest/rpcx/graphs/contributors).