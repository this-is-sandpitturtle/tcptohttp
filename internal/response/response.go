package response

import (
	"errors"
	"fmt"
	"io"
	"log"
	"strconv"
	"syscall"

	"github.com/this-is-sandpitturtle/httpfromtcp/internal/headers"
)

type Writer struct {
    StatusCode StatusCode
    Headers headers.Headers
    Body []byte
    W io.Writer
    Status WriterStatus
}

type HandlerError struct {
    StatusCode StatusCode
    Message string
}

type StatusCode int

const (
    OK StatusCode = 200
    BR StatusCode = 400
    ISE StatusCode = 500
)

type WriterStatus int 

const (
    Initialized WriterStatus = iota
    RequestLineDone
    HeadersDone
    BodyDone
)

func (w *Writer) WriteTrailers (trailers headers.Headers) error{
    fmt.Println(trailers)
    for key, val := range trailers {
        _, err := w.W.Write([]byte(fmt.Sprintf("%s: %s\r\n", key, val)))
        if err != nil {
            fmt.Println()
            fmt.Println(err)
            return err
        }
        fmt.Println(key)
        fmt.Println(val)
    }
    w.W.Write([]byte("\r\n"))
    return nil
}

func (w *Writer) WriteChunkedBody(p []byte) (int, error ) {
    len := len(p)
    writtenTotal := 0
    //Just for testing normally you should reduce the number of syscalls
    //TODO
    n, err := w.W.Write([]byte(fmt.Sprintf("%X\r\n",len)))
    if err != nil {
        return 0, err
    }
    writtenTotal += n
    n, err = w.W.Write(p)
    if err != nil {
        return writtenTotal, err
    }
    writtenTotal += n
    n, err = w.W.Write([]byte("\r\n"))
    if err != nil {
        return writtenTotal, err
    }
    writtenTotal += n
    return writtenTotal, err
}

func (w *Writer) WriteChunkedBodyDone() (int, error) {
    defer func() {
        w.Status = BodyDone
    }()
    n, err := w.W.Write([]byte(fmt.Sprintf("%X\r\n",0)))
    if err != nil {
        return 0, err
    }
    return n, err
}

func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
    if w.Status != Initialized {
        return  errors.New("cannot write request line, writer has to be in state initialized")
    }

    defer func() {
        w.Status = RequestLineDone
    }()

    err := WriteStatusLine(w.W, statusCode)
    if err != nil {
        return err
    }
    return nil
}

func (w *Writer) WriteHeaders(Headers headers.Headers) error {
    if w.Status != RequestLineDone{
        return  errors.New("cannot write headers, writer has to be in state RequestLineDone")
    }
    defer func () {
        w.Status = HeadersDone
    }()
    err := WriteHeaders(w.W, Headers)
    if err != nil {
        return err
    }
    return nil
}

func (w *Writer) WriteBody(b []byte) (int, error) {
    if w.Status != HeadersDone{
        return  0, errors.New("cannot write body, writer has to be in state HeadersDone")
    }

    defer func() {
        w.Status = BodyDone
    }()

    n, err := WriteBody(w.W, b)
    if err != nil {
        return 0, err
    }
    return n, nil
}

func (h HandlerError) WriteError(w io.Writer) error {
    mess := []byte(h.Message)
    contentLength := len(mess)
    headers := GetDefaultHeaders(contentLength)
    fmt.Println("headers: ", headers)
    err := WriteStatusLine(w, h.StatusCode)
    if err != nil {
        fmt.Printf("write error - write status line: %v", err)
        return err
    }
    err = WriteHeaders(w, headers)
    if err != nil {
        fmt.Printf("write error - write headers: %v", err)
        return err
    }
    n, err := WriteBody(w, mess)
    if err != nil {
        fmt.Printf("write error - write body: %v", err)
        return err
    }
    if n != contentLength {
        log.Fatalf("content-length didn't match body length")
        return errors.New("content-length didn't match body length")
    }

    return nil
}


func WriteStatusLine(w io.Writer, statusCode StatusCode) error {
    switch statusCode {
    case OK: 
        _,err := w.Write([]byte("HTTP/1.1 200 OK\r\n"))
        if err != nil {
            log.Fatalf("error writing status-line: %v", err)
            return err
        }
        return nil
    case BR: 
        _,err := w.Write([]byte("HTTP/1.1 400 Bad Request\r\n"))
        if err != nil {
            log.Fatalf("error writing status-line: %v", err)
            return err
        }
        return nil
    case ISE: 
        _,err := w.Write([]byte("HTTP/1.1 500 Internal Server Error\r\n"))
        if err != nil {
            log.Fatalf("error writing status-line: %v", err)
            return err
        }
        return nil
    default:
        _,err := w.Write([]byte("HTTP/1.1 200 \r\n"))
        if err != nil {
            log.Fatalf("error writing status-line: %v", err)
            return err
        }
        return nil
    }
}

func GetDefaultHeaders(contentLength int) headers.Headers {
    headers := headers.NewHeaders()
    headers.Set("Content-Length", strconv.Itoa(contentLength))
    headers.Set("Connection", "close")
    headers.Set("Content-Type", "text/plain")
    return headers
}

func WriteHeaders(w io.Writer, headers headers.Headers) error {
    fmt.Println(headers)
    for key, val := range headers {
        n, err := w.Write([]byte(fmt.Sprintf("%s: %s\r\n", key, val)))
        if err != nil {
            fmt.Println("error occured after writing number of bytes: ", n)
            if errors.Is(err, syscall.EPIPE) {
                fmt.Println("EPIPE error")
            }
            return err
        }
    }
    w.Write([]byte("\r\n"))
    return nil
}

func WriteError(w io.Writer, errH HandlerError) error {
    mess := []byte(errH.Message)
    contentLength := len(mess)
    headers := GetDefaultHeaders(contentLength)
    fmt.Println("headers: ", headers)
    err := WriteStatusLine(w, errH.StatusCode)
    if err != nil {
        fmt.Printf("write error - write status line: %v", err)
        return err
    }
    err = WriteHeaders(w, headers)
    if err != nil {
        fmt.Printf("write error - write headers: %v", err)
        return err
    }
    n, err := WriteBody(w, mess)
    if err != nil {
        fmt.Printf("write error - write body: %v", err)
        return err
    }
    if n != contentLength {
        log.Fatalf("content-length didn't match body length")
        return errors.New("content-length didn't match body length")
    }

    return nil
}

func WriteBody(w io.Writer, data []byte) (int, error) {
    n, err := w.Write(data)
    if err != nil {
        log.Fatalf("couldn't write body: %v", err)
        return 0, errors.New("body couldn't be written")
    }
    return n, nil
}

