package main

import (
	"fmt"
	"log"
	"net"

	"github.com/this-is-sandpitturtle/httpfromtcp/internal/request"
)

func main() {
    listener, err := net.Listen("tcp", ":42069")
    if err != nil {
        log.Fatal("listener couldn't be opened")
    }
    defer listener.Close()
    for {
        conn, err := listener.Accept()
        if err != nil {
            log.Fatal("connection error")
        }
        fmt.Println("connection from ",conn.RemoteAddr()," has been established")
        req, err := request.RequestFromReader(conn)
        if err != nil {
            fmt.Println("request couldn't be parsed")
        }
        fmt.Println("- Request line:")
        fmt.Printf("- Method: %s\n", req.RequestLine.Method)
        fmt.Printf("- Target: %s\n", req.RequestLine.RequestTarget)
        fmt.Printf("- Version: %s\n", req.RequestLine.HttpVersion)
        fmt.Println("- Headers:")
        for key, val := range req.Headers {
            fmt.Printf("- %s: %s\n", key, val)
        }
        fmt.Println("- Body:")
        fmt.Println(string(req.Body))
    }
}

//func getLinesChannel(f io.ReadCloser) <-chan string {
//    ch := make(chan string)
//    go func() {
//        defer close(ch)
//        defer f.Close()
//        currentLine :=""
//        for {
//            b := make([]byte, 8)
//            _, err := f.Read(b) 
//            if err != nil {
//                if currentLine != "" {
//                    ch <- currentLine
//                }
//                if errors.Is(err,io.EOF) {
//                    break
//                }
//                return 
//            }
//            splits := strings.Split(string(b), "\n")
//            if len(splits) == 2 {
//                currentLine += string(splits[0])
//                ch <- currentLine
//                currentLine = ""
//                currentLine += string(splits[1])
//            } else {
//                currentLine += string(splits[0])
//            }
//        }
//    }()
//
//    return ch
//}
