package headers

import (
	"bytes"
	"fmt"
	"strings"
)

type Headers struct {
	headers map[string]string
}

func (h *Headers) Get(name string) string {
	return h.headers[strings.ToLower(name)]
}

func (h *Headers) Set(name, value string) {
	name = strings.ToLower(name)
	if v, ok := h.headers[name]; ok {
		h.headers[name] = fmt.Sprintf("%s, %s", v, value)
	} else {
		h.headers[name] = value
	}
}

func (h *Headers) ForEach(cb func(n, v string)) {
	for n, v := range h.headers {
		cb(n, v)
	}
}

var RN = []byte("\r\n")
var ERROR_MALFORMED_TOKENNAME = fmt.Errorf("Invalid token name. Only specific characters are allowed")
var ERROR_MALFORMED_FIELDNAME = fmt.Errorf("Malformed Field name")
var ERROR_MALFORMED_FIELDLINE = fmt.Errorf("Malformed Field line")

func NewHeaders() *Headers {
	return &Headers{
		headers: map[string]string{},
	}
}

func isToken(str []byte) bool {

	for _, ch := range str {
		found := false

		if ch >= 'A' && ch <= 'Z' ||
			ch >= 'a' && ch <= 'z' ||
			ch >= '0' && ch <= '9' {
			found = true
		}
		switch ch {
		case '#', '$', '%', '&', '\'', '*', '+', '-', '.', '^', '_', '`', '|', '~':
			found = true
		}

		if !found {
			return false
		}
	}
	return true
}

func parseHeader(fieldLine []byte) (string, string, error) {
	parts := bytes.SplitN(fieldLine, []byte(":"), 2)
	if len(parts) != 2 {
		return "", "", ERROR_MALFORMED_FIELDNAME
	}

	name := parts[0]
	value := bytes.TrimSpace(parts[1])

	if bytes.HasSuffix(name, []byte(" ")) {
		return "", "", ERROR_MALFORMED_FIELDLINE
	}

	return string(name), string(value), nil

}

func (h *Headers) Parse(data []byte) (int, bool, error) {
	read := 0
	done := false
	for {
		idx := bytes.Index(data[read:], RN)
		if idx == -1 {
			break
		}

		if idx == 0 {
			done = true
			read += len(RN)
			break
		}

		name, value, err := parseHeader(data[read : read+idx])
		if err != nil {
			return 0, false, err
		}

		if !isToken([]byte(name)) {
			return 0, false, ERROR_MALFORMED_TOKENNAME
		}

		read += idx + len(RN)

		h.Set(name, value)
	}

	return read, done, nil
}
