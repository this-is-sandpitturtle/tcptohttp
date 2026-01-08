package request

import (
	"bytes"
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
    line := make([]byte, 1024)
    window := make([]byte, 1024)
    var alreadyParsed int
    var alreadyRead int
    for {
        read, err := reader.Read(window)
        if read == 0 {
            //MAYBE THIS IS FEASIBLE TO DETERMINE THE END OF PARTIAL CONTENT
            //break when no new data is read
            //in tcp we send ACK and NACK to acknowledge receiving of data
            //so when we read no new data the request should definitely be over.
            //and should have been already parsed.
            //TODO
            //Except for cases where we read the entire request in one go...
            // ---------------------------------------------------------------
            //in cases where we read the entire request in one line (realistic buffer size of 1024 bytes for example)
            //i think we have to check if we already have more than one part --> determined by crlf
            //and we have to loop through the parts and parse.
            //then it should work.
            //parse and resize until the request is completely consumed i guess...
            break
        }

        for i:=alreadyParsed; i<alreadyRead + read; i++ {
            if i >= len(line) {
                line = append(line, window[i - alreadyRead])
            }
            line[i] = window[i - alreadyRead]
        }
        alreadyRead += read
        alreadyParsed += read
    
        consumedBytes, err := parsedRequest.parse(line)

        if err != nil {
            return &Request{}, err
        }

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
        
        if parsedRequest.state == done {
            break
        }
        
    }

    if parsedRequest.state != done {
        fmt.Println(parsedRequest)
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
        idx := bytes.Index(data, []byte(crlf))
        lastIdx := bytes.LastIndex(data, []byte(crlf))
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
        if lastIdx > idx {
            r.state = done
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
            r.state = done
            return 0, nil
        }

        nonNullChar := 0
        for i:=0; i<len(data); i++ {
            if data[i] != '\x00' {
                nonNullChar++
            }
        }
        //if len(data) >= contentLength {
        //    if nonNullChar >= contentLength && data[contentLength-1] == '\x00' { //letzter eintrag von darf nicht leer sein
        //        fmt.Println("VERSUCH TERMINIERUNG DURCH CHECK")
        //        fmt.Printf("%q\n", string(data))
        //        fmt.Println(strconv.Itoa(nonNullChar))
        //        fmt.Println(strconv.Itoa(contentLength))
        //        return 0, errors.New("body-length < content-length")
        //    }
        //}

        if nonNullChar == contentLength {
            r.Body = append(r.Body, data[:nonNullChar]...) //bis letzter non null charakter und jedes element
            r.state = done
            return contentLength, nil
        }

        if len(r.Body) > contentLength {
            return 0, errors.New("body length exceeds content-length")
        }
        return 0, nil
    case done:
        return 0, errors.New("trying to read into request with state done")
    default:
        return 0, errors.New("unsupported state")
    }


}
