package main

import (
	"fmt"
	"log"
	"net"

	"github.com/horsedevours/httpfromtcp/internal/request"
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
		request, err := request.RequestFromReader(conn)
		if err != nil {
			log.Println("error reading request: ", err)
		}
		fmt.Println("Request line:")
		fmt.Printf("- Method: %s\n", request.RequestLine.Method)
		fmt.Printf("- Target: %s\n", request.RequestLine.RequestTarget)
		fmt.Printf("- Version: %s\n", request.RequestLine.HttpVersion)
		fmt.Println("Headers:")
		for k, v := range request.Headers {
			fmt.Printf("- %s: %s\n", k, v)
		}
	}
}
