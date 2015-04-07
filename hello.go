package main

import (
	"encoding/json"
	"flag"
	"log"
	"net"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/exec"
)

type Message struct {
	Message string `json:"message"`
}

const (
	helloWorldString = "Hello, World!"
)

var (
	helloWorldBytes = []byte(helloWorldString)
)

var prefork int
var child = flag.Bool("child", false, "is child proc")

func main() {
	var listener net.Listener
	flag.IntVar(&prefork, "prefork", 0, "Number of worker process.")
	flag.Parse()
	if prefork != 0 {
		listener = doPrefork()
	}

	http.HandleFunc("/json", jsonHandler)
	http.HandleFunc("/plaintext", plaintextHandler)
	http.HandleFunc("/", plaintextHandler)

	if prefork == 0 {
		http.ListenAndServe(":8080", nil)
	} else {
		http.Serve(listener, nil)
	}
}

func doPrefork() (listener net.Listener) {
	var err error
	var fl *os.File
	var tcplistener *net.TCPListener
	if !*child {
		var addr *net.TCPAddr
		addr, err = net.ResolveTCPAddr("tcp", ":8080")
		if err != nil {
			log.Fatal(err)
		}
		tcplistener, err = net.ListenTCP("tcp", addr)
		if err != nil {
			log.Fatal(err)
		}
		fl, err = tcplistener.File()
		if err != nil {
			log.Fatal(err)
		}
		children := make([]*exec.Cmd, prefork)
		for i := range children {
			children[i] = exec.Command(os.Args[0], "-prefork=1", "-child")
			children[i].Stdout = os.Stdout
			children[i].Stderr = os.Stderr
			children[i].ExtraFiles = []*os.File{fl}
			err = children[i].Start()
			if err != nil {
				log.Fatal(err)
			}
		}
		for _, ch := range children {
			var err error = ch.Wait()
			if err != nil {
				log.Print(err)
			}
		}
		os.Exit(0)
	} else {
		fl = os.NewFile(3, "")
		listener, err = net.FileListener(fl)
		if err != nil {
			log.Fatal(err)
		}
	}
	return listener
}

// Test 1: JSON serialization
func jsonHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(&Message{helloWorldString})
}

// Test 6: Plaintext
func plaintextHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.Write(helloWorldBytes)
}
