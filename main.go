package main

import (
	"fmt"
	"io"
	"os"
)

func main() {
	file, err := os.Open("messages.txt")
	if err != nil {
		fmt.Printf("couldn't open the file: %s", err)
		os.Exit(1)
	}
	defer file.Close()

	chunk := make([]byte, 8)
	for n := 0; err != io.EOF; n, err = file.Read(chunk) {
		if err != nil {
			fmt.Printf("error reading from file: %s", err)
			os.Exit(1)
		}
		fmt.Printf("read: %s\n", string(chunk[0:n]))

	}

	os.Exit(0)
}
