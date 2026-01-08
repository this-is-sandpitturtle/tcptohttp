package request

import (
	//	"bytes"
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
const NULL = "\x00"
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
    req := Request{
        RequestLine: RequestLine{},
        Headers: headers.NewHeaders(),
        Body: make([]byte, 0),
    }
    line := make([]byte, 1024)
    alreadyRead := 0
    for req.state != done {
        //make new buffer if read bytes extend over length of line
        if alreadyRead >= len(line) {
            newLine := make([]byte, len(line) * 2)
            copy(newLine, line[:])
            line = newLine
        }
        readBytes, err := reader.Read(line[alreadyRead:])
        if readBytes == 0 {
            if req.state != done {
                fmt.Printf("incomplete request, in state: %d, read n bytes on EOF: %d\n", req.state, readBytes)
                fmt.Printf("%q\n", line)
                fmt.Println(req)
                return nil, errors.New("BASICALLY EOF ERROR")
            }
            break
        }
        //If err != nil {
        //    return nil, err
        //}
        alreadyRead += readBytes

        alreadyParsed, err := req.parse(line[:alreadyRead]) //other try pass in to index of all read bytes
        if err != nil {
            return nil, err
        }
        
        //reset line and alreadyRead
        copy(line, line[alreadyParsed:])
        alreadyRead -= alreadyParsed 
    }

    return &req, nil
}

func parseRequestLine(data []byte) (*RequestLine, int, error) {
    s := string(data)
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
    bytesParsed := 0
    for r.state != done {
        consumed, err := r.parseSingle(data[bytesParsed:])
        if err != nil {
            return 0, err
        }
        bytesParsed += consumed

        if consumed == 0 {
            break
        }
    }
    
    return bytesParsed, nil
}

func (r *Request) parseSingle(data []byte) (int, error) {
    switch r.state {
    case initialized:
       // idx := bytes.Index(data, []byte(crlf))
       // lastIdx := bytes.LastIndex(data, []byte(crlf))
        parsedRequest, consumed, err := parseRequestLine(data)

        if err != nil {
            return 0, err
        }

        if consumed != 0 {
            r.RequestLine.Method = parsedRequest.Method
            r.RequestLine.RequestTarget = parsedRequest.RequestTarget
            r.RequestLine.HttpVersion = parsedRequest.HttpVersion
            r.state = parsingHeaders
        }
       // if lastIdx > idx {
       //     r.state = done
       // }
        return consumed, nil
    case parsingHeaders:
        fmt.Printf("%q\n", string(data))
        consumedBytes, finished, err := r.Headers.Parse(data) 
        if err != nil {
            fmt.Printf("errors parsing header: %s\n", string(data))
            return 0, err
        }
        if finished {
            if n, ok := r.Headers[contentLengthHeader]; !ok || n == "0" {
                r.state = done
            } else {
                fmt.Println("After Header Parsing")
                fmt.Printf("%q\n", data)
                r.state = parsingBody
            }
        }
        return consumedBytes, nil
    case parsingBody:
        fmt.Println("Parsing Body")
        fmt.Printf("%q\n", data)
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
            fmt.Println("das hier geht schief richtig?")
            r.Body = append(r.Body, data[:nonNullChar]...) //bis letzter non null charakter und jedes element
            r.state = done
            return contentLength, nil
        }

        if len(r.Body) > contentLength {
            fmt.Println("das hier geht schief richtig?")
            return 0, errors.New("body length exceeds content-length")
        }

        return 0, nil
    case done:
        return 0, errors.New("trying to read into request with state done")
    default:
        return 0, errors.New("unsupported state")
    }


}
