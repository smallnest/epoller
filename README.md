# epoller
epoll implementation for connections in Linux, MacOS and windows.

[![License](https://img.shields.io/:license-MIT-blue.svg)](https://opensource.org/licenses/MIT) [![GoDoc](https://godoc.org/github.com/smallnest/epoller?status.png)](http://godoc.org/github.com/smallnest/epoller)  [![travis](https://travis-ci.org/smallnest/epoller.svg?branch=master)](https://travis-ci.org/smallnest/epoller) [![Go Report Card](https://goreportcard.com/badge/github.com/smallnest/epoller)](https://goreportcard.com/report/github.com/smallnest/epoller) [![coveralls](https://coveralls.io/repos/smallnest/epoller/badge.svg?branch=master&service=github)](https://coveralls.io/github/smallnest/epoller?branch=master) 


Its target is implementing a simple epoll for connection, so you should see it only contains few methods:

```go
type Poller interface {
	Add(conn net.Conn) error
	Remove(conn net.Conn) error
	Wait() ([]net.Conn, error)
	Close() error
}
```

Welcome any PRs for windows IOCompletePort.

Inspired by [1m-go-websockets](https://github.com/eranyanay/1m-go-websockets).

Thanks @sunnyboy00 for providing windows implementation.