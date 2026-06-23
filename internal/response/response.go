package response

import (
	"fmt"
	"io"

	"github.com/vishnu-788/tcp-to-http/internal/headers"
)

type Response struct {
}

type StatusCode int

const (
	StatusOK                  StatusCode = 200
	StatusBadRequest          StatusCode = 400
	StatusInternalServerError StatusCode = 500
)

func GetDefaultHeaders(contenLen int) *headers.Headers {
	h := headers.NewHeaders()

	h.Set("Content-Length", fmt.Sprintf("%d", contenLen))
	h.Set("Connection", "close")
	h.Set("Content-Type", "text/plain")

	return h
}

type Writer struct {
	writer io.Writer
}

func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
	statusLine := []byte{}
	switch statusCode {
	case StatusOK:
		statusLine = []byte("HTTP/1.1 200 OK\r\n")
	case StatusBadRequest:
		statusLine = []byte("HTTP/1.1 400 BAD REQUEST\r\n")
	case StatusInternalServerError:
		statusLine = []byte("HTTP/1.1 500 INTERNAL SERVER ERROR\r\n")
	default:
		return fmt.Errorf("Unrecognized Error code")
	}

	_, err := w.writer.Write(statusLine)
	return err
}

func (w *Writer) WriteHeaders(h *headers.Headers) error {
	b := []byte{}

	h.ForEach(func(n, v string) {
		b = fmt.Appendf(b, "%s: %s\r\n", n, v)

	})
	b = fmt.Append(b, "\r\n")
	_, err := w.writer.Write(b)
	return err
}

func (w *Writer) WriteBody(p []byte) (int, error) {
	n, err := w.writer.Write(p)
	return n, err
}

func NewWriter(w io.Writer) *Writer {
	return &Writer {
		writer: w,
	}
}
