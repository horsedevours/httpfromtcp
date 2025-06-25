package request

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"
	"unicode"
)

type ParseState int

const (
	Initialized ParseState = iota
	Done
)

type Request struct {
	ParseState  ParseState
	RequestLine RequestLine
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

const bufferSize = 8

func RequestFromReader(reader io.Reader) (*Request, error) {
	request := &Request{ParseState: Initialized}
	buf := make([]byte, bufferSize)
	readToIndex := 0
	bytesParsed := 0
	for request.ParseState != Done {
		if readToIndex == len(buf) {
			temp := make([]byte, 2*len(buf))
			copy(temp, buf)
			buf = temp
		}
		n, err := reader.Read(buf[readToIndex:])
		if err != nil {
			if err == io.EOF {
				request.ParseState = Done
				break
			}
			return &Request{}, fmt.Errorf("error reading request: %w", err)
		}
		readToIndex += n

		m, err := request.parse(buf[:readToIndex])
		if err != nil {
			return &Request{}, fmt.Errorf("error parsing bytes: %w", err)
		}
		if m != 0 {
			bytesParsed += m
			temp := make([]byte, len(buf))
			copy(temp, buf[bytesParsed:])
			buf = temp
			readToIndex -= bytesParsed
		}
	}

	return request, nil
}

func (r *Request) parse(data []byte) (int, error) {
	if r.ParseState == Initialized {
		n, err := r.parseRequestLine(data)
		if err != nil {
			return 0, err
		}
		if n == 0 {
			return 0, nil
		}
		r.ParseState = Done
		return n, nil
	}
	if r.ParseState == Done {
		return 0, errors.New("error trying to read data in a done state")
	}
	return 0, errors.New("error unknown state")
}

func (r *Request) parseRequestLine(request []byte) (int, error) {
	requestLine, _, found := bytes.Cut(request, []byte{'\r', '\n'})
	if !found {
		return 0, nil
	}

	requestLineParts := strings.Split(string(requestLine), " ")
	if len(requestLineParts) != 3 {
		return 0, fmt.Errorf("request line does not contain 3 required parts: %v", requestLineParts)
	}

	method := requestLineParts[0]
	for _, r := range method {
		if !unicode.IsLetter(r) || unicode.IsLower(r) {
			return 0, fmt.Errorf("request method must be all caps: %s", method)
		}
	}
	r.RequestLine.Method = method
	r.RequestLine.RequestTarget = requestLineParts[1]

	if requestLineParts[2] != "HTTP/1.1" {
		return 0, fmt.Errorf("invalid http version: %s", requestLineParts[2])
	}
	version, _ := strings.CutPrefix(requestLineParts[2], "HTTP/")
	r.RequestLine.HttpVersion = version

	return len(requestLine), nil
}
