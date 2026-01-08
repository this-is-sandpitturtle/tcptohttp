package request

import (
	"errors"
	"io"
	"log"
	"regexp"
	"strings"
)

const supportedVersion = "1.1"

type Request struct {
   RequestLine RequestLine 
}

type RequestLine struct {
    HttpVersion   string
    RequestTarget string
    Method        string
}

func RequestFromReader(reader io.Reader) (*Request, error) {
    requestString, err := io.ReadAll(reader)
    if err != nil {
       log.Fatal("request could not be slurped") 
    }
    requestLine, err := parseRequestLine(string(requestString))
    if err != nil {
        return nil, err
    }
    return &Request{RequestLine: *requestLine},nil
}

func parseRequestLine(s string) (*RequestLine, error) {
    split := strings.Split(s, "\r\n")
    requestLine := split[0]
    requestLineSplit := strings.Split(requestLine, " ")
    if len(requestLineSplit) != 3 {
        return &RequestLine{}, errors.New("request line doesn't have right amount of parts")
    }
    methodRegex, err := regexp.Compile("[A-Z]{3}")
    if err != nil {
        log.Fatal("methodRegex couldn't be compiled")
    }
    method := requestLineSplit[0]
    target := requestLineSplit[1]
    version := requestLineSplit[2]

    if !methodRegex.Match([]byte(method)) {
        return &RequestLine{}, errors.New("method does contain illegal characters")
    }

    if !strings.Contains(version, supportedVersion) || len(version) < 5 {
        return &RequestLine{}, errors.New("version not supported or to tiny")
    }
    version = version[5:]

    out := RequestLine{
        HttpVersion: version,
        RequestTarget: target,
        Method: method,
    }
    
    return &out, nil
}
