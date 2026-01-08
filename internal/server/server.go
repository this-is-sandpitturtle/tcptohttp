package server

import (
	"fmt"
	"log"
	"net"
	"sync/atomic"
)

type Server struct {
    Running atomic.Bool
    Listener net.Listener
}

type state int
const protocol = "tcp"

const (
    starting state = iota
    running
    stopped
)

func Serve(port int) (*Server, error) {
    serv := Server{Running: atomic.Bool{}}
    listener, err := net.Listen(protocol, fmt.Sprintf(":%d", port))
    if err != nil {
        log.Fatalf("listener could not be created")
    }
    serv.Listener = listener
    serv.Running.Store(true)
    go serv.listen()
    return &serv, nil
}

func (s *Server) Close() error {
    err := s.Listener.Close()
    if err != nil {
        log.Fatalf("listener couldn't be closed")
        return err
    }
    s.Running.Store(false)
    return nil
}

func (s *Server) listen() {
    for {
        conn, err := s.Listener.Accept()
        if err != nil {
            if !s.Running.Load() {
                break //server not running
            } else {
                log.Fatalf("connection couldn't be accepted")
            }
        }
        go s.handle(conn)
    }
}

func (s *Server) handle(conn net.Conn) {
    defer conn.Close()
    response := "HTTP/1.1 200 OK\r\n" +
                "Content-Type: text/plain\r\n" + 
                //"Content-Length: 14\r\n" +
                "\r\n" +    
                "Hello World!\r\n"
    _, err := conn.Write([]byte(response))
    if err != nil {
        log.Fatal("couldn't write response")
        return
    }
    return
}
