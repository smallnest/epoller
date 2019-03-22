# epoller
epoll implementation for connections in Linux and MacOS.

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
