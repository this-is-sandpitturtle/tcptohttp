package request

import (
	"errors"
	"fmt"
	"io"
	"log"
	"regexp"
	"strconv"
	"strings"

	"github.com/this-is-sandpitturtle/httpfromtcp/internal/headers"
)

type state int
const supportedVersion = "1.1"
const crlf = "\r\n"
const contentLengthHeader = "content-length"

const (
    initialized state = iota
    parsingHeaders
    parsingBody
    done 
)

type Request struct {
   RequestLine RequestLine 
   Headers headers.Headers
   Body []byte
   state state
}

type RequestLine struct {
    Method        string
    RequestTarget string
    HttpVersion   string
}

func RequestFromReader(reader io.Reader) (*Request, error) {
    parsedRequest := Request{
        RequestLine: RequestLine{},
        Headers: headers.NewHeaders(),
        Body: make([]byte, 0),
    }
    line := make([]byte, 8)
    window := make([]byte, 8)
    var alreadyParsed int
    var alreadyRead int
    eof := false
    for {
        if eof {
            break
        }
        read, err := reader.Read(window)
        if err != nil {
            if errors.Is(err, io.EOF) {
                eof = true
            }
        }
        //fmt.Printf("read: %d \n", read)
        for i:=alreadyParsed; i<alreadyRead + read; i++ {
            if i >= len(line) {
                line = append(line, window[i - alreadyRead])
            }
            line[i] = window[i - alreadyRead]
        }
        alreadyRead += read
        alreadyParsed += read
        
    
        consumedBytes, err := parsedRequest.parse(line)

        if consumedBytes != 0 {
            newLine := make([]byte, len(line))
            j:=0
            for i:=consumedBytes; i<len(line); i++ {
                if line[i] != '\x00' {
                    newLine[j] = line[i]
                    j++
                }
            }

            line = newLine

            alreadyParsed = j
            alreadyRead = j
        }

        if err != nil {
            return &Request{}, err
        }
        
        if parsedRequest.state == done {
            break
        }
        
    }

    if parsedRequest.state != done {
        fmt.Println(parsedRequest)
        fmt.Println(parsedRequest.state)
        return &parsedRequest, errors.New("invalid request ---------------")
    }

    return &parsedRequest, nil
}

func parseRequestLine(s string) (*RequestLine, int, error) {
    split := strings.Split(s, crlf)
    idxCrlf := strings.Index(s, crlf)
    if len(split) < 2 {
        //need more data to parse request line
        return &RequestLine{}, 0, nil 
    }
    //fmt.Println("===================")
    //fmt.Println("Gesamte Zeile:")
    //fmt.Println(s)
    //fmt.Println("===================")

    requestLine := split[0]
    requestLineSplit := strings.Split(requestLine, " ")
    if len(requestLineSplit) != 3 {
        return &RequestLine{}, 0, errors.New("request line doesn't have right amount of parts")
    }
    methodRegex, err := regexp.Compile("[A-Z]{3,8}")
    if err != nil {
        log.Fatal("methodRegex couldn't be compiled")
    }
    method := requestLineSplit[0]
    target := requestLineSplit[1]
    version := requestLineSplit[2]

    if !methodRegex.Match([]byte(method)) {
        return &RequestLine{}, 0, errors.New("method does contain illegal characters")
    }

    if !strings.Contains(version, supportedVersion) || len(version) < 5 {
        return &RequestLine{}, 0, errors.New("version not supported or to tiny")
    }
    version = version[5:]

    out := RequestLine{
        HttpVersion: version,
        RequestTarget: target,
        Method: method,
    }

    return &out, idxCrlf+2, nil
}

func (r *Request) parse(data []byte) (int, error) {
    switch r.state {
    case initialized:
        s := string(data)
        parsedRequest, consumed, err := parseRequestLine(s)

        if err != nil {
            return 0, err
        }

        if consumed != 0 {
            r.RequestLine.Method = parsedRequest.Method
            r.RequestLine.RequestTarget = parsedRequest.RequestTarget
            r.RequestLine.HttpVersion = parsedRequest.HttpVersion
            r.state = parsingHeaders
        }
        return consumed, nil
    case parsingHeaders:
        consumedBytes, finished, err := r.Headers.Parse(data) 
        if err != nil {
            fmt.Printf("errors parsing header: %s\n", string(data))
            return 0, err
        }
        if finished {
            if n, ok := r.Headers[contentLengthHeader]; !ok || n == "0" {
                r.state = done
            } else {
                r.state = parsingBody
            }
        }
        return consumedBytes, nil
    case parsingBody:
        cl, ok := r.Headers.Get(contentLengthHeader)
        if !ok {
            r.state = done
            return 0, nil
        }
        contentLength, err := strconv.Atoi(cl)
        if err != nil {
            return 0, errors.New("invalid content-length, couldn't convert to number")
        }
        //if idx := bytes.Index(data, []byte(crlf)); idx != 0 {
        //    return 0, errors.New("Invalid start of body")
        //}
        if contentLength == 0 {
            fmt.Println("CONTENT-LENGTH 0")
            r.state = done
            return 0, nil
        }

        nonNullChar := 0
        
        for i:=0; i<len(data); i++ {
            if data[i] != '\x00' {
                nonNullChar++
            }
        }

        if len(r.Body) > contentLength {
            fmt.Printf("%q\n", string(data))
            fmt.Println("Body:")
            fmt.Printf("%q\n", string(r.Body))
            fmt.Println(len(data))
            return 0, errors.New("body length exceeds content-length")
        }

        if nonNullChar == contentLength {
            r.Body = append(r.Body, data[:nonNullChar]...) //bis letzter non null charakter und jedes element
            r.state = done
            return contentLength, nil
        }

        return 0, nil
    case done:
        return 0, errors.New("trying to read into request with state done")
    default:
        return 0, errors.New("unsupported state")
    }


}
