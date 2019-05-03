package main

import (
	epoller "./epoller"

	"fmt"
	"io"
	"net"
	"time"
)

func main() {
	num := 10
	msgPerConn := 10

	poller, err := epoller.NewPoller()
	if err != nil {
		fmt.Println(err)
	}
	//fmt.Println(poller.Fd)
	// start server
	ln, err := net.Listen("tcp", "127.0.0.1:8000")
	if err != nil {
		fmt.Println(err)
	}
	defer ln.Close()
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			//fmt.Println("get conn ", conn)
			err = poller.Add(conn)
			if err != nil {
				fmt.Println(err)
			}
		}
	}()

	// create num connections and send msgPerConn messages per connection
	for i := 0; i < num; i++ {
		go func() {
			conn, err := net.Dial("tcp", ln.Addr().String())
			if err != nil {
				fmt.Println(err)
				return
			}
			time.Sleep(20 * time.Microsecond)
			for i := 0; i < msgPerConn; i++ {
				conn.Write([]byte("hello world"))
			}
			conn.Close()
		}()
	}

	time.Sleep(100 * time.Millisecond)

	// read those num * msgPerConn messages, and each message (hello world) contains 11 bytes.
	ch := make(chan struct{})
	var total int
	var count int
	var expected = num * msgPerConn * len("hello world")
	go func() {
		for {
			conns, err := poller.Wait(128)
			if err != nil {
				fmt.Println(err)
			}
			count++
			var buf = make([]byte, 11)
			for _, conn := range conns {
				n, err := conn.Read(buf)
				if err != nil {
					if err == io.EOF {

						err = poller.Remove(conn)
						if err != nil {
							fmt.Println(err)
						}
						conn.Close()

					} else {
						fmt.Println(err)
					}
				} else {
					//fmt.Println(" total read buff ", total, buf[0:n])
				}

				total += n
				n = 0
			}

			if total == expected {
				break
			}
		}

		fmt.Println("read all %d bytes, count: %d", total, count)
		close(ch)
	}()

	select {
	case <-ch:
	case <-time.After(200 * time.Second):
	}

	if total != expected {
		fmt.Println("epoller does not work. expect %d bytes but got %d bytes", expected, total)
	}
}
