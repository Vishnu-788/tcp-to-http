package request

import (
	"bytes"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/vishnu-788/tcp-to-http/internal/headers"
)

type parserState string

// Enums
const (
	StateInit    parserState = "init"
	StateHeaders parserState = "headers"
	StateBody    parserState = "body"
	StateDone    parserState = "done"
	StateError   parserState = "error"
)

var ERROR_MALFORMED_REQUEST_LINE = fmt.Errorf("Malformed Request-line")
var ERROR_UNSUPPORTED_HTTP_VERSION = fmt.Errorf("Unsupported HTTP varsion. Only HTTP 1/1 is allowed")
var ERROR_REQ_IN_ERROR_STATE = fmt.Errorf("Request is in error state")
var ERROR_STATE = fmt.Errorf("Error occured in the Request state.")

var SEPARATOR = []byte("\r\n")

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

type Request struct {
	RequestLine RequestLine
	state       parserState
	Headers     *headers.Headers
	Body        string
}

func (r *Request) parse(data []byte) (int, error) {
	read := 0
outer:
	for {
		currentData := data[read:]

		if len(currentData) == 0 {
			break outer
		}

		switch r.state {
		case StateInit:
			rl, n, err := parseRequestLine(currentData)

			if err != nil {
				r.state = StateError
				return 0, err
			}

			if n == 0 {
				break outer
			}

			r.RequestLine = *rl
			read += n

			r.state = StateHeaders

		case StateHeaders:
			n, done, err := r.Headers.Parse(currentData)
			if err != nil {
				r.state = StateError
				return 0, err
			}

			if n == 0 {
				break outer
			}

			if done {
				// TODO: If the content doesn't have any body update state to done else to body
				if r.hasBody() {
					r.state = StateBody
				} else {
					r.state = StateDone
				}
			}

			read += n

		case StateBody:
			length := getInt(r.Headers, "content-length", 0)	
			if length == 0 {
				r.state = StateDone
				break
			}
			
			remaining := min(length - len(r.Body), len(currentData))
			r.Body += string(currentData[:remaining])

			read += remaining

			if length == len(r.Body) {
				r.state = StateDone
			}

		case StateDone:
			break outer

		case StateError:
			return 0, ERROR_REQ_IN_ERROR_STATE

		default:
			panic("Error occured in the state machine. [switch case default error throwed]")
		}

	}
	return read, nil
}

func (r *Request) done() bool {
	return r.state == StateDone || r.state == StateError
}

func (r *Request) hasBody() bool {
	return getInt(r.Headers, "content-length", 0) > 0
}

func getInt(headers *headers.Headers, name string, defaultValue int) int {
	vStr, exists := headers.Get(name)
	if !exists {
		return defaultValue
	}
	
	v, err := strconv.Atoi(vStr)
	if err != nil {
		return defaultValue
	}

	return v

}

func newRequest() *Request {
	return &Request{
		state:   StateInit,
		Headers: headers.NewHeaders(),
		Body: "",
	}
}

func parseRequestLine(b []byte) (*RequestLine, int, error) {
	idx := bytes.Index(b, SEPARATOR)
	if idx == -1 {
		return nil, 0, nil
	}

	startLine := b[:idx]
	read := idx + len(SEPARATOR)

	parts := bytes.Split(startLine, []byte(" "))

	if len(parts) != 3 {
		return nil, 0, ERROR_MALFORMED_REQUEST_LINE
	}

	httpParts := strings.Split(string(parts[2]), "/")
	if len(httpParts) != 2 || string(httpParts[0]) != "HTTP" || string(httpParts[1]) != "1.1" {
		return nil, 0, ERROR_UNSUPPORTED_HTTP_VERSION
	}

	rl := RequestLine{
		Method:        string(parts[0]),
		RequestTarget: string(parts[1]),
		HttpVersion:   string(httpParts[1]),
	}

	return &rl, read, nil
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	request := newRequest()

	// TODO: There is case where the request is more that 1k.
	buf := make([]byte, 1024)
	bufLen := 0

	for !request.done() {
		n, err := reader.Read(buf[bufLen:])
		// TODO: what to do with the question marks.
		if err != nil {
			return nil, err
		}

		bufLen += n

		readN, err := request.parse(buf[:bufLen])
		if err != nil {
			return nil, err
		}

		copy(buf, buf[readN:bufLen])
		bufLen -= readN
	}

	return request, nil
}
