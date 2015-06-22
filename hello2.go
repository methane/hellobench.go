package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"runtime"
	"sync"
	"time"
)

const N = 100000

var message = []byte("Hello")

func serve(l net.Listener) {
	for {
		conn, err := l.Accept()
		if err != nil {
			log.Fatal(err)
		}
		go handler(conn)
	}
}

func handler(conn net.Conn) {
	defer conn.Close()
	buf := make([]byte, 128)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			if err != io.EOF {
				log.Println(err)
			}
			return
		}
		conn.Write(buf[:n])
	}
}

func client(addr string, n int, wg *sync.WaitGroup) {
	defer wg.Done()

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	buf := make([]byte, 64)
	for i := 0; i < n; i++ {
		_, err := conn.Write(message)
		if err != nil {
			log.Fatal(err)
		}

		x, err := conn.Read(buf)
		if err != nil {
			log.Fatal(nil)
		}
		_ = x
	}
}

func main() {
	fmt.Println("Go version: ", runtime.Version())
	p := runtime.GOMAXPROCS(0)
	fmt.Println("MAXPROCS: ", p)

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		log.Fatal(err)
	}
	addr := listener.Addr().String()
	fmt.Println("Listening on: ", addr)
	go serve(listener)

	var wg sync.WaitGroup
	for c := 1; c <= p*4; c *= 2 {
		wg.Add(c)
		t0 := time.Now()
		for i := 0; i < c; i++ {
			go client(addr, N/c, &wg)
		}
		wg.Wait()
		t1 := time.Now()
		fmt.Printf("concurrency: %-3d  %.3f req/sec\n", c, float64(N/c*c)/t1.Sub(t0).Seconds())
		//fmt.Println("concurrency: ", c, "\t", float64(N/c*c)/t1.Sub(t0).Seconds(), " req/sec")
	}
}
