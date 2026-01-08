package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
)

const port = ":42069"

func main(){
    addr, err := net.ResolveUDPAddr("udp", port)

    if err != nil {
        log.Fatalf("couldn't resolve udp adress on port: %s", port)
    }

    conn, err := net.DialUDP("udp", nil, addr)
    if err != nil {
        log.Fatalf("udp connection couldn't be established %v", err)
    }
    defer conn.Close()

    reader := bufio.NewReader(os.Stdin)
    
    for {
        fmt.Print(">")
        line, err := reader.ReadString('\n')
        if err != nil {
            log.Fatal(err)
        }
        conn.Write([]byte(line))
    }

}
