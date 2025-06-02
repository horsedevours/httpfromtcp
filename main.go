package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
)

func main() {
	lstnr, err := net.Listen("tcp", ":42069")
	if err != nil {
		log.Fatalf("error creating listener: %s", err)
	}
	defer lstnr.Close()

	for {
		conn, err := lstnr.Accept()
		if err != nil {
			log.Fatalf("error creating connection: %s", err)
		}
		log.Println("message has been accepted")
		strChan := getLinesChannel(conn)
		for {
			str, ok := <-strChan
			if !ok {
				fmt.Println("connection has been closed")
				break
			}
			fmt.Printf("%s\n", str)
		}
	}
}

func getLinesChannel(f io.ReadCloser) <-chan string {
	chunk := make([]byte, 8)
	strChan := make(chan string)

	go func() {
		defer f.Close()
		defer close(strChan)
		line := ""
		for {
			n, err := f.Read(chunk)
			if errors.Is(err, io.EOF) {
				if line != "" {
					strChan <- line
				}
				break
			} else if err != nil {
				log.Fatalf("error reading input: %s", err)
			}

			parts := strings.Split(string(chunk[:n]), "\n")
			for _, part := range parts[:len(parts)-1] {
				line += part
				strChan <- line
				line = ""
			}

			line += parts[len(parts)-1]
		}
	}()

	return strChan
}
