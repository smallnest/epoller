package main

import (
	. "github.com/smallnest/epoller"
	"log"
	"net"
)

type NetPoller struct {
	Poller   Poller
	WriteReq chan uint64
}

func main() {
	var np NetPoller

	for i := 0; i < 2; i++ {
		poller, err := NewPoller()
		if err != nil {
			log.Fatal(err)
		}
		if err != nil {
			log.Fatal(err)
		}
		// the following line cause goroutine stack grow and copy local variables to new allocated stack and switch to new stack
		// but runtime.adjustpointers will check whether pointers bigger than runtime.minLegalPointer(4096) or throw a panic
		// fatal error: invalid pointer found on stack (runtime/stack.go:599@go1.14.3)
		// since NewEpoller return A pointer created by CreateIoCompletionPort may less than 4096
		np = NetPoller{
			Poller:   poller,
			WriteReq: make(chan uint64, 1000000),
		}

	}
	poller := np.Poller
	// start server
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		log.Fatal(err)
	}
	defer ln.Close()
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}

			poller.Add(conn)
		}
	}()

}
