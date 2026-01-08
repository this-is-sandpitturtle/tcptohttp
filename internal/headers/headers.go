package headers

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
)

type Headers map[string]string

const crlf = "\r\n"
const colon = ":"
const invalidSigns = "{}[]()@/?<>=;:,\"\t\\\x20"

func NewHeaders() Headers {
    return Headers{}
}

func (h Headers) Set(key, value string) {
    h[key] = value
}


//TODO: look at when the done thing gets set one test fails because en empty line gets read in.
//TODO: maybe a "split" is not working because of one error in a test with whitespaces and doublen crlf
func (h Headers) Parse(data []byte) (n int, done bool, err error) {
    idx := bytes.Index(data, []byte(crlf))
    consumedBytes := idx + 2
    if idx == -1{
        return 0, false, nil
    }
    if idx == 0 {
        fmt.Println("HALLO WIR BRECHEN AB")
        fmt.Printf("%q", string(data))
        return 0, true, nil
    }
    fmt.Println("=================")
    fmt.Println("|||||||||||||||||")
    fmt.Printf("%q", string(data))
    fmt.Println()
    fmt.Println("=================")
    
    colonIdx := bytes.IndexAny(data, colon)
    
    if colonIdx == -1 || colonIdx == 0 {
        return 0, false, errors.New("no colon in header line")
    }

    if data[colonIdx - 1] == ' ' {
        return 0, false, errors.New("no whitespace in front of colon allowed")
    }
    
    keyBytes := data[:colonIdx]
    if bytes.ContainsAny(keyBytes, invalidSigns) {
        return 0, false, errors.New("header-field contains invalid signs")
    }
    if len(keyBytes) == 0 {
        return 0, false, errors.New("header-field has no key")
    }

    valueBytes := data[colonIdx+1:idx] //von : bis crlf

    if len(valueBytes) == 0 {
        return 0, false, errors.New("header-field-value has no key")
    }

//    consumedBytes := 0
    key := string(keyBytes)
    value := string(valueBytes)
//    consumedBytes += len(key) + len(value) + 2

    key = strings.ToLower(strings.TrimSpace(key))
    value = strings.ToLower(strings.TrimSpace(value))

    //fmt.Println("=================")
    //fmt.Printf("%q",key)
    //fmt.Println(len(key))
    //fmt.Println("=================")
    

    //Multiple Headers
    if _,ok := h[key]; ok {
        h.Set(key, h[key] + ", " + value)
    } else {
        h.Set(key, value)
    }

    //len(data)-2 because the \r\n doesn't count as consumed
    //maybe there is a problem with consuming bytes because the last \r\n shouldn't count as consumed
    //however if it only ends on one crlf it should count as consumed
    //alternativ: oben wenn die keys gemacht werden len_key + len_val + 2 fuer \r\n
    fmt.Println("consumed-bytes:")
    fmt.Println(consumedBytes)
    return consumedBytes, false, nil
}
