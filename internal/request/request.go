package request

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"
	"unicode"

	"github.com/horsedevours/httpfromtcp/internal/headers"
)

type ParseState int

const (
	Initialized ParseState = iota
	ParsingHeaders
	ParsingBody
	Done
)

type Request struct {
	ParseState  ParseState
	RequestLine RequestLine
	Headers     headers.Headers
	Body        []byte
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

const bufferSize = 8

func RequestFromReader(reader io.Reader) (*Request, error) {
	request := &Request{ParseState: Initialized, Headers: headers.NewHeaders()}
	buf := make([]byte, bufferSize)
	readToIndex := 0
	for request.ParseState != Done {
		if readToIndex == len(buf) {
			temp := make([]byte, 2*len(buf))
			copy(temp, buf)
			buf = temp
		}
		n, err := reader.Read(buf[readToIndex:])
		if err != nil {
			if err == io.EOF {
				if request.ParseState != Done {
					return &Request{}, fmt.Errorf("incomplete request")
				}
				break
			} else {
				return &Request{}, fmt.Errorf("error reading request: %w", err)
			}
		}
		readToIndex += n

		bytesParsed, err := request.parse(buf[:readToIndex])
		if err != nil {
			return &Request{}, fmt.Errorf("error parsing bytes: %w", err)
		}
		if bytesParsed != 0 {
			temp := make([]byte, len(buf))
			copy(temp, buf[bytesParsed:])
			buf = temp
			readToIndex -= bytesParsed
			bytesParsed = 0
		}
	}

	return request, nil
}

func (r *Request) parse(data []byte) (int, error) {
	totalBytesParsed := 0
	for r.ParseState != Done {
		switch r.ParseState {
		case Initialized:
			n, err := r.parseRequestLine(data[totalBytesParsed:])
			if err != nil {
				return 0, err
			}
			if n == 0 {
				return totalBytesParsed, nil
			}
			r.ParseState = ParsingHeaders
			totalBytesParsed += n + 2
		case ParsingHeaders:
			n, done, err := r.Headers.Parse(data[totalBytesParsed:])
			if err != nil {
				return 0, err
			}
			if n == 0 {
				return totalBytesParsed, nil
			}
			totalBytesParsed += n
			if done {
				contentLength := r.Headers.Get("Content-Length")
				if contentLength == "" || contentLength == "0" {
					r.ParseState = Done
					return totalBytesParsed, nil
				}
				r.ParseState = ParsingBody
			}
		case ParsingBody:
			contentLength := r.Headers.Get("Content-Length")
			numContentLength, err := strconv.Atoi(contentLength)
			if err != nil {
				return 0, fmt.Errorf("cannot convert content length to number: %s", contentLength)
			}
			r.Body = append(r.Body, data[totalBytesParsed:]...)
			totalBytesParsed += len(data[totalBytesParsed:])

			if len(r.Body) > numContentLength {
				return 0, fmt.Errorf("length of body (%d) exceeds content length (%d)", len(r.Body), numContentLength)
			}
			if len(r.Body) == numContentLength {
				r.ParseState = Done
				log.Println("All content consumed")
			}
		case Done:
			return 0, errors.New("error trying to read data in a done state")
		default:
			return 0, errors.New("error unknown state")
		}
		if totalBytesParsed == len(data) {
			break
		}
	}
	return totalBytesParsed, nil
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
