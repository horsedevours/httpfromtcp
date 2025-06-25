package request

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"unicode"
)

type Request struct {
	RequestLine RequestLine
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	request, err := io.ReadAll(reader)
	if err != nil {
		return &Request{}, fmt.Errorf("error reading request: %w", err)
	}

	requestLine, err := parseRequestLine(request)
	if err != nil {
		return &Request{}, err
	}

	return &Request{RequestLine: requestLine}, nil
}

func parseRequestLine(request []byte) (RequestLine, error) {
	requestLine, _, found := bytes.Cut(request, []byte{'\r', '\n'})
	if !found {
		return RequestLine{}, fmt.Errorf("malformed request: %s", string(request))
	}

	requestLineParts := strings.Split(string(requestLine), " ")
	if len(requestLineParts) != 3 {
		return RequestLine{}, fmt.Errorf("request line does not contain 3 required parts: %v", requestLineParts)
	}

	method := requestLineParts[0]
	for _, r := range method {
		if !unicode.IsLetter(r) || unicode.IsLower(r) {
			return RequestLine{}, fmt.Errorf("request method must be all caps: %s", method)
		}
	}

	if requestLineParts[2] != "HTTP/1.1" {
		return RequestLine{}, fmt.Errorf("invalid http version: %s", requestLineParts[2])
	}
	version, _ := strings.CutPrefix(requestLineParts[2], "HTTP/")

	return RequestLine{Method: method, RequestTarget: requestLineParts[1], HttpVersion: version}, nil
}
