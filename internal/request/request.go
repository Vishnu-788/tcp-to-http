package request

import (
	"errors"
	"fmt"
	"io"
	"strings"
)

type Request struct {
	RequestLine RequestLine
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

var ERROR_MALFORMED_REQUEST_LINE = fmt.Errorf("Malformed Request-line")
var ERROR_UNSUPPORTED_HTTP_VERSION = fmt.Errorf("Unsupported HTTP varsion. Only HTTP 1/1 is allowed")
var SEPARATOR = "\r\n"

func parseRequestLine(b string) (*RequestLine, string, error) {
	idx := strings.Index(b, SEPARATOR)
	if idx == -1 {
		return nil, b, nil
	}

	startLine := b[:idx]
	restOfMsg := b[idx+len(SEPARATOR):]

	parts := strings.Split(startLine, " ")

	if len(parts) != 3 {
		return nil, "", ERROR_MALFORMED_REQUEST_LINE
	}

	httpParts := strings.Split(parts[2], "/")
	if len(httpParts) != 2 || httpParts[0] != "HTTP" || httpParts[1] != "1.1" {
		return nil, "", ERROR_UNSUPPORTED_HTTP_VERSION
	}

	rl := RequestLine{
		Method:        parts[0],
		RequestTarget: parts[1],
		HttpVersion:   httpParts[1],
	}

	return &rl, restOfMsg, nil
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, errors.Join(fmt.Errorf("unable to io.ReadAll"), err)
	}

	str := string(data)

	rl, str, err := parseRequestLine(str)

	if err !=  nil {
		return nil, err
	}

	return &Request{
		RequestLine: *rl,
	}, err
}
