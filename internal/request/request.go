package request

import (
	"errors"
	"fmt"
	"io"
	"log"
	"regexp"
	"strings"

	"github.com/this-is-sandpitturtle/httpfromtcp/internal/headers"
)

type state int
const supportedVersion = "1.1"
const crlf = "\r\n"

const (
    initialized state = iota
    parsingHeaders
    done 
)

type Request struct {
   RequestLine RequestLine 
   Headers headers.Headers
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
    }
    line := make([]byte, 8)
    window := make([]byte, 8)
    var alreadyParsed int
    var alreadyRead int
    var counter int
    for {
        counter++
        read, err := reader.Read(window)
        if counter >= 100 {
            break
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
            //fmt.Println("=====================")
            //fmt.Println("Parsed Request so far:")
            //fmt.Println(parsedRequest)
            //fmt.Println("=====================")
            //fmt.Println("Vorher:")
            //fmt.Println(string(line))

            newLine := make([]byte, len(line))
            j:=0
            for i:=consumedBytes; i<len(line); i++ {
                if line[i] != '\x00' {
                    newLine[j] = line[i]
                }
                j++
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
            // maybe?
            //fmt.Println("error in parseRequest Line")
            //fmt.Println(err)
            r.state = done
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
        //fmt.Println("=====================")
        //fmt.Println(string(data))
        //fmt.Println("=====================")
        consumedBytes, finished, err := r.Headers.Parse(data) //bool ignored
        if err != nil {
            fmt.Printf("errors parsing header: %s\n", string(data))
            return 0, err
        }
        if finished {
            r.state = done
        }
        //DEBUG
        if consumedBytes != 0 {
            fmt.Println("=====================")
            fmt.Println(r.Headers)
            fmt.Println("=====================")
        }
        return consumedBytes, nil
    case done:
        return 0, errors.New("trying to read into request with state done")
    default:
        return 0, errors.New("unsupported state")
    }


}
