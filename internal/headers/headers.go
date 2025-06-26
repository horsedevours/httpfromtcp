package headers

import (
	"bytes"
	"errors"
	"fmt"
)

type Headers map[string]string

func NewHeaders() Headers {
	return Headers{}
}

func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	crlfIdx := bytes.Index(data, []byte{'\r', '\n'})
	if crlfIdx == -1 {
		return
	}
	if crlfIdx == 0 {
		n = 2
		done = true
		return
	}

	header := bytes.TrimSpace(data[:crlfIdx])
	split := bytes.SplitN(header, []byte{':'}, 2)
	if len(split) < 2 {
		err = errors.New("error splitting header by delimiter")
		return
	}

	trimmedName := bytes.ToLower(bytes.TrimRight(split[0], " "))
	if len(split[0]) != len(trimmedName) {
		err = fmt.Errorf("malformed header: %s", string(header))
		return
	}
	for _, b := range trimmedName {
		if !(('a' <= b && b <= 'z') || ('0' <= b && b <= '9') || bytes.Contains([]byte{'!', '#', '$', '%', '&', '\'', '*', '+', '-', '.', '^', '_', '`', '|', '~'}, []byte{b})) {
			err = fmt.Errorf("character not allowed in header name: %s", string(b))
			return
		}
	}
	trimmedValue := bytes.TrimLeft(split[1], " ")

	if val, found := h[string(trimmedName)]; found {
		h[string(trimmedName)] = fmt.Sprintf("%s, %s", val, string(trimmedValue))
	} else {
		h[string(trimmedName)] = string(trimmedValue)
	}

	n = crlfIdx + 2
	return
}
