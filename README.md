# epoller
epoll implementation for connections in **Linux**, **MacOS** and **Windows**.

[![License](https://img.shields.io/:license-MIT-blue.svg)](https://opensource.org/licenses/MIT) [![GoDoc](https://godoc.org/github.com/smallnest/epoller?status.png)](http://godoc.org/github.com/smallnest/epoller)  [![github actions](https://github.com/smallnest/epoller/actions/workflows/go.yml/badge.svg)](https://github.com/smallnest/epoller/actions) [![Go Report Card](https://goreportcard.com/badge/github.com/smallnest/epoller)](https://goreportcard.com/report/github.com/smallnest/epoller) [![coveralls](https://coveralls.io/repos/smallnest/epoller/badge.svg?branch=master&service=github)](https://coveralls.io/github/smallnest/epoller?branch=master) 


Its target is implementing a simple epoll lib for nework connections, so you should see it only contains few methods about net.Conn:

```go
type Poller interface {
	// Add adds an existed connection to poller.
	Add(conn net.Conn) error
	// Remove removes a connection from poller. Notice it doesn't call the conn.Close method.
	Remove(conn net.Conn) error
	// Wait waits for at most count events and returns the connections.
	Wait(count int) ([]net.Conn, error)
	// Close closes the poller. If closeConns is true, it will close all the connections.
	Close(closeConns bool) error
}
```


Inspired by [1m-go-websockets](https://github.com/eranyanay/1m-go-websockets).

## Contributors
see [contributors](https://github.com/smallnest/rpcx/graphs/contributors).