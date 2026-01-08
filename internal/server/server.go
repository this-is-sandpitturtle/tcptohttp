package server

import (
	"fmt"
	"log"
	"net"
	"sync/atomic"

	"github.com/this-is-sandpitturtle/httpfromtcp/internal/request"
	"github.com/this-is-sandpitturtle/httpfromtcp/internal/response"
)

type Server struct {
    Listener net.Listener
    Handler Handler
    Running atomic.Bool
}

type Handler func(w *response.Writer, req *request.Request) 

type state int
const protocol = "tcp"

const (
    starting state = iota
    running
    stopped
)

func Serve(port int, h Handler) (*Server, error) {
    serv := Server{Running: atomic.Bool{}}
    listener, err := net.Listen(protocol, fmt.Sprintf(":%d", port))
    if err != nil {
        log.Fatalf("listener could not be created")
    }
    serv.Listener = listener
    serv.Running.Store(true)
    serv.Handler = h
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

//Don't know maybe this needs a handler
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
    req, err := request.RequestFromReader(conn)
    fmt.Println(req)
    if err != nil {
        fmt.Println("Generic error response without body: ", err)
        response.WriteError(conn, response.HandlerError{StatusCode: 400, Message: err.Error()})
        return
    }
    buffer := &response.Writer{
        W: conn,
        Status: response.Initialized,
        }
    s.Handler(buffer, req)
//    if handlerError != nil {
//        fmt.Println("handler error - response write error")
//        fmt.Println(handlerError)
//        handlerError.WriteError(conn)
//        return
//    }
   // buffer.WriteStatusLine(buffer.StatusCode)
   // buffer.WriteHeaders(buffer.Headers)
   // buffer.WriteBody(buffer.Body)
    //body := buffer.String()
    //head := response.GetDefaultHeaders(len(body))
    //err = response.WriteStatusLine(conn, 200)
    //if err != nil {
    //    fmt.Println("status line writing problem: ", err)
    //    return
    //}

    //err = response.WriteHeaders(conn, head)
    //if err != nil {
    //    fmt.Println("header writing went wrong: ", err)
    //    return
    //}

    //_,err = response.WriteBody(conn, []byte(body))
    //if err != nil {
    //    fmt.Println("body writing went wrong: ", err)
    //}
}


