package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
)

func main() {
	addr, err := net.ResolveUDPAddr("udp", ":42069")
	if err != nil {
		log.Fatalf("failed to resolve UPD address: %v", err)
	}

	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		log.Fatalf("failed to create UDP connection: %v", err)
	}
	defer conn.Close()

	rdr := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("> ")

		input, err := rdr.ReadString('\n')
		if err != nil {
			log.Println("failed to read input: ", err)
		}

		_, err = conn.Write([]byte(input))
		if err != nil {
			log.Println("failed to write to UDP connection: ", err)
		}
	}
}
