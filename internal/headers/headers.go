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
const invalidSigns = "{}[]()@/?<>=;:,\"\t\x20"
//const invalidSigns = "{}[]()@/?<>=;:,\"\t"

func NewHeaders() Headers {
    return Headers{}
}

func (h Headers) Set(key, value string) {
    h[key] = value
}

func (h Headers) Get(key string) (string, bool) {
    key = strings.ToLower(key)
    out, ok := h[key]
    return out, ok
}


func (h Headers) Parse(data []byte) (n int, done bool, err error) {
    idx := bytes.Index(data, []byte(crlf))
    consumedBytes := idx + 2
//    fmt.Printf("%q\n", data)
    if idx == -1 {
        return 0, false, nil
    }
    if idx == 0 {
        fmt.Println("HEADERS FINISHED")
        return 2, true, nil
    }
    
    colonIdx := bytes.IndexAny(data, colon)
    
    if colonIdx == -1 || colonIdx == 0 {
        return 0, false, errors.New("no colon in header line")
    }

    if data[colonIdx - 1] == ' ' {
        return 0, false, errors.New("no whitespace in front of colon allowed")
    }
    
    keyBytes := data[:colonIdx]
    keyBytes = bytes.Trim(keyBytes, string('\x20')) //spaces
    if bytes.ContainsAny(keyBytes, invalidSigns) {
        fmt.Printf("%q", string(data))
        fmt.Println()
        fmt.Printf("%q", string(keyBytes))
        return 0, false, errors.New("header-field contains invalid signs")
    }
    if len(keyBytes) == 0 {
        return 0, false, errors.New("header-field has no key")
    }

    valueBytes := data[colonIdx+1:idx] //von : bis crlf

    if len(valueBytes) == 0 {
        return 0, false, errors.New("header-field-value has no key")
    }

    key := string(keyBytes)
    value := string(valueBytes)

    key = strings.ToLower(strings.TrimSpace(key))
    value = strings.ToLower(strings.TrimSpace(value))
    

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
    return consumedBytes, false, nil
}
