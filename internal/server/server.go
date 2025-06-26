package server

import (
	"fmt"
	"log"
	"net"
	"sync/atomic"
)

type Server struct {
	Listener net.Listener
	Open     atomic.Bool
}

func Serve(port int) (*Server, error) {
	lstnr, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return &Server{}, err
	}

	server := &Server{}
	server.Listener = lstnr
	server.Open.Store(true)
	go server.listen()

	return server, nil
}

func (s *Server) Close() error {
	err := s.Listener.Close()
	if err != nil {
		return err
	}
	s.Open.Store(false)
	return nil
}

func (s *Server) listen() {
	for {
		conn, err := s.Listener.Accept()
		fmt.Println("Received something")
		if err != nil {
			if !s.Open.Load() {
				log.Println("Server closed")
				return
			}
			log.Fatalf("Failed to accept connection: %v", err)
		}

		go s.handle(conn)
	}
}

func (s *Server) handle(conn net.Conn) {
	defer conn.Close()
	_, err := conn.Write([]byte("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\n\r\nHello World!"))
	if err != nil {
		fmt.Printf("An error: %v", err)
	}
}
