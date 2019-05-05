package main

import (
	"flag"
	"log"
	"net"
	"strconv"
	"strings"
	"sync"

	"github.com/smallnest/epoller"
	"github.com/smallnest/redcon"
)

var (
	port = flag.Int("port", 6380, "server port")
	pool = flag.Int("pool", 100, "pool size for epoll")
)

var (
	mu      sync.RWMutex
	keys    = make(map[string]string)
	readers = make(map[net.Conn]*redcon.Reader)
	writers = make(map[net.Conn]*redcon.Writer)
)

func main() {
	ln, err := net.Listen("tcp", ":"+strconv.Itoa(*port))
	if err != nil {
		log.Fatalf("failed to listen %d: %v", *port, err)
		return
	}

	poller, err := epoller.NewPollerWithBuffer(128)
	if err != nil {
		log.Fatalf("failed to create a poller for port %d: %v", port, err)
		return
	}

	go poll(poller)
	for {
		select {
		default:
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			poller.Add(conn)

			mu.Lock()
			readers[conn] = redcon.NewReader(conn)
			writers[conn] = redcon.NewWriter(conn)
			mu.Unlock()
		}
	}
}

func poll(poller epoller.Poller) {
	for {

		conns, err := poller.WaitWithBuffer()
		if err != nil {
			if err.Error() != "bad file descriptor" {
				log.Printf("failed to poll: %v", err)
			}

			continue
		}

		for _, conn := range conns {
			mu.RLock()
			r := readers[conn]
			w := writers[conn]
			mu.RUnlock()

			if r == nil || w == nil {
				continue
			}

			cmds, err := r.ReadCommands()
			if err != nil {
				closeConn(conn)
				continue
			}

			for _, cmd := range cmds {
				handleCmd(conn, &cmd, w)
			}
		}
	}
}

func handleCmd(conn net.Conn, cmd *redcon.Command, w *redcon.Writer) {
	args := cmd.GetAllArgs()
	op := strings.ToUpper(string(args[0]))
	switch op {
	default:
		w.WriteError("ERR unknown command '" + string(args[0]) + "'")
	case "PING":
		if len(args) > 2 {
			w.WriteError("ERR wrong number of arguments for '" + string(args[0]) + "' command")
		} else if len(args) == 2 {
			w.WriteBulk(args[1])
		} else {
			w.WriteString("PONG")
		}
	case "ECHO":
		if len(args) != 2 {
			w.WriteError("ERR wrong number of arguments for '" + string(args[0]) + "' command")
		} else {
			w.WriteBulk(args[1])
		}
	case "QUIT":
		w.WriteString("OK")
		w.Flush()
		closeConn(conn)
		return
	case "GET":
		if len(args) != 2 {
			w.WriteError("ERR wrong number of arguments for '" + string(args[0]) + "' command")
		} else {
			key := string(args[1])
			mu.Lock()
			val, ok := keys[key]
			mu.Unlock()
			if !ok {
				w.WriteNull()
			} else {
				w.WriteBulkString(val)
			}
		}
	case "SET":
		if len(args) != 3 {
			w.WriteError("ERR wrong number of arguments for '" + string(args[0]) + "' command")
		} else {
			key, val := string(args[1]), string(args[2])
			mu.Lock()
			keys[key] = val
			mu.Unlock()
			w.WriteString("OK")
		}
	case "DEL":
		if len(args) < 2 {
			w.WriteError("ERR wrong number of arguments for '" + string(args[0]) + "' command")
		} else {
			var n int
			mu.Lock()
			for i := 1; i < len(args); i++ {
				if _, ok := keys[string(args[i])]; ok {
					n++
					delete(keys, string(args[i]))
				}
			}
			mu.Unlock()
			w.WriteInt64(int64(n))
		}
	case "FLUSHDB":
		mu.Lock()
		keys = make(map[string]string)
		mu.Unlock()
		w.WriteString("OK")
	}
	w.Flush()
}

func closeConn(conn net.Conn) {
	conn.Close()
	mu.Lock()
	delete(readers, conn)
	delete(writers, conn)
	mu.Unlock()
}
