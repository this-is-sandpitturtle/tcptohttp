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
        ch := getLinesChannel(conn)
        for s := range ch {
            fmt.Print(s)
        }

        fmt.Printf("\nconnection from %v is closed\n", conn.RemoteAddr())
    }
}

func getLinesChannel(f io.ReadCloser) <-chan string {
    ch := make(chan string)
    go func() {
        defer close(ch)
        defer f.Close()
        currentLine :=""
        for {
            b := make([]byte, 8)
            _, err := f.Read(b) 
            if err != nil {
                if currentLine != "" {
                    ch <- currentLine
                }
                if errors.Is(err,io.EOF) {
                    break
                }
                return 
            }
            splits := strings.Split(string(b), "\n")
            if len(splits) == 2 {
                ch <- currentLine
                currentLine += string(splits[1])
                currentLine = ""
            }
            currentLine += string(splits[0])
        }
    }()

    return ch
}
