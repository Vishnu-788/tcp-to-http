package main

import (
	"fmt"
	"log"
	"net"

	"github.com/vishnu-788/tcp-to-http/internal/request"
)

func main() {
	listener, err := net.Listen("tcp", ":42069")
	if err != nil {
		log.Fatal("error opening file", err)
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal("error opening file", err)
		}

		r, err := request.RequestFromReader(conn)
		if err != nil {
			log.Fatal("error", err)
		}

		fmt.Printf("Request Line: \n")
		fmt.Printf("- Method: %s\n", r.RequestLine.Method)
		fmt.Printf("- Target: %s\n", r.RequestLine.RequestTarget)
		fmt.Printf("- Version: %s\n", r.RequestLine.HttpVersion)

	}
}
