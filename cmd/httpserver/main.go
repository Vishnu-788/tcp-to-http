package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/vishnu-788/tcp-to-http/internal/request"
	"github.com/vishnu-788/tcp-to-http/internal/response"
	"github.com/vishnu-788/tcp-to-http/internal/server"
)

const port = 42069

func tempHandler(w *response.Writer, req *request.Request) {
	h := response.GetDefaultHeaders(0)

	body := respond200()
	status := response.StatusOK

	if req.RequestLine.RequestTarget == "/yourproblem" {
		body = respond400()
		status = response.StatusBadRequest
	} else 	if req.RequestLine.RequestTarget == "/myproblem" {
		body = respond500()
		status = response.StatusInternalServerError
	}

	h.Replace("Content-length", fmt.Sprintf("%d", len(body)))
	w.WriteStatusLine(status)
	w.WriteHeaders(h)
	w.WriteBody(body)
}


func main() {
	server, err := server.Serve(port, tempHandler)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}

	defer server.Close()
	log.Printf("Server started on PORT: %v", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped...")
}

func respond500() []byte {
	return []byte(`
	<html>
		<head>500 Internal Server Error</head>
		<body>
			<h1>Internal Server Error</h1>
			<p>okay, you know what this one is on me.</p>
		</body>
	</html>
	`)
}

func respond400() []byte {
	return []byte(`
	<html>
		<head>400 Bad Request</head>
		<body>
			<h1>Bad Request</h1>
			<p>What did you even send. This is dumb manh.</p>
		</body>
	</html>
	`)
}

func respond200() []byte {
	return []byte(`
	<html>
		<head>200 Ok</head>
		<body>
			<h1>Ok good!!</h1>
			<p>Good job smh...</p>
		</body>
	</html>
	`)
}
