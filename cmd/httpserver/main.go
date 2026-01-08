package main

import (
	"crypto/sha256"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/this-is-sandpitturtle/httpfromtcp/internal/headers"
	"github.com/this-is-sandpitturtle/httpfromtcp/internal/request"
	"github.com/this-is-sandpitturtle/httpfromtcp/internal/response"
	"github.com/this-is-sandpitturtle/httpfromtcp/internal/server"
)

const port = 42069

//func(w io.Writer, *req) *HandlerError
func main() {
    handler := func(w *response.Writer, req *request.Request) {
        if req.RequestLine.RequestTarget == "/yourproblem" {
            handle400(w)
            return
        } 
        if req.RequestLine.RequestTarget == "/myproblem" {
            handle500(w)
            return
        }
        if strings.HasPrefix(req.RequestLine.RequestTarget, "/httpbin") {
            fmt.Println("BIG CHUNGUS")
            handleHttpBin(w, strings.TrimPrefix(req.RequestLine.RequestTarget,"/httpbin"))
            return
        }
        if req.RequestLine.RequestTarget == "/video" {
            handleVideo(w)
        }
        handle200(w)
    }

    server, err := server.Serve(port, handler)
    if err != nil {
        log.Fatalf("Error starting server: %v", err)
    }
    defer server.Close()
    log.Println("Server started on port:", port)
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
    <-sigChan
    log.Println("Server gracefully stopped")
}

func handle400(w *response.Writer) {
    res := 
    `<html>
    <head>
    <title>400 Bad Request</title>
    </head>
    <body>
    <h1>Bad Request</h1>
    <p>Your request honestly kinda sucked.</p>
    </body>
    </html>`
    w.Body = []byte(res)
    w.WriteStatusLine(400)
    headers := response.GetDefaultHeaders(len(res))
    headers.Set("Content-Type", "text/html")
    w.Headers = headers
    w.WriteHeaders(w.Headers)
    w.WriteBody(w.Body)
}

func handle500(w *response.Writer) {
    res := 
    `<html>
    <head>
    <title>500 Bad Request</title>
    </head>
    <body>
    <h1>Internal Server Error</h1>
    <p>Okay, you know what? This one is on me.</p>
    </body>
    </html>`
    w.Body = []byte(res)
    w.WriteStatusLine(500)
    headers := response.GetDefaultHeaders(len(res))
    headers.Set("Content-Type", "text/html")
    w.Headers = headers
    w.WriteHeaders(w.Headers)
    w.WriteBody(w.Body)
}

func handle200(w *response.Writer) {
    res := 
    `<html>
    <head>
    <title>200 OK</title>
    </head>
    <body>
    <h1>Success!</h1>
    <p>Your request was an absolute banger.</p>
    </body>
    </html>`
    w.Body = []byte(res)
    w.WriteStatusLine(200)
    headers := response.GetDefaultHeaders(len(res))
    headers.Set("Content-Type", "text/html")
    w.Headers = headers
    w.WriteHeaders(w.Headers)
    w.WriteBody(w.Body)
}

func handleHttpBin(w *response.Writer, target string) {
    fmt.Println("-------------")
    fmt.Println("https://httpbin.org" + target)
    fmt.Println("-------------")
    resp, err := http.Get("https://httpbin.org" + target)
    
    fmt.Println(resp)
    if err != nil {
        log.Fatalf("error calling https://httbin.org/%s", target)
        return
    }
    buffer := make([]byte, 1024)
    persBuffer := make([]byte, 0) //for checksum
    dataRead := 0 // X-Content-Length
    noMoreData := false

    w.WriteStatusLine(200)
    heads := response.GetDefaultHeaders(0)
    delete(heads, "Content-Length")
    heads.Set("Transfer-Encoding", "chunked")
    heads.Set("Trailer", "X-Content-SHA256, X-Content-Length")
    w.WriteHeaders(heads)
    
    defer resp.Body.Close()
    for {
        if noMoreData {
            break
        }
        read,_  := resp.Body.Read(buffer)
        dataRead += read
        persBuffer = append(persBuffer, buffer[:read]...)
//        if err != nil {
//            log.Fatalf("error reading response: %v\n", err)
//            break
//        }
        fmt.Printf("amount of bytes read: %d\r\n", read)
        w.WriteChunkedBody(buffer)
        buffer = make([]byte, 1024)

        if read == 0 {
            noMoreData = true
        }
    }
    w.WriteChunkedBodyDone()
    
    trailers := headers.NewHeaders()
    checkSum := sha256.Sum256(persBuffer)
    trailers.Set("X-Content-Sha256", fmt.Sprintf("%x", checkSum))
    trailers.Set("X-Content-Length", strconv.Itoa(len(persBuffer)))
    w.WriteTrailers(trailers)

}

func handleVideo(w *response.Writer) {
    video, err := os.ReadFile("./assets/vim.mp4")
    if err != nil {
        log.Fatalf("error reading video, %v", err)
    }
    w.WriteStatusLine(200)
    heads := response.GetDefaultHeaders(len(video))
    heads.Set("Content-Type", "video/mp4")
    fmt.Println("===================")
    fmt.Println(heads)
    fmt.Println("===================")
    w.WriteHeaders(heads)
    w.WriteBody(video)
}

