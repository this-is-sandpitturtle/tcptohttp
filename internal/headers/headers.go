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
const invalidSigns = "{}[]()@/?<>=;:,\""

func NewHeaders() Headers {
    return Headers{}
}

func (h Headers) Set(key, value string) {
    h[key] = value
}

func (h Headers) Parse(data []byte) (n int, done bool, err error) {
    idx := bytes.IndexAny(data, crlf)
    if idx == -1{
        return 0, false, nil
    }
    if idx == 0 {
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
    if bytes.ContainsAny(keyBytes, invalidSigns) {
        return 0, false, errors.New("header-field contains invalid signs")
    }
    if len(keyBytes) == 0 {
        return 0, false, errors.New("header-field has no key")
    }

    valueBytes := data[colonIdx+1:]

    if len(valueBytes) == 0 {
        return 0, false, errors.New("header-field-value has no key")
    }

    key := string(keyBytes)
    value := string(valueBytes)

    key = strings.ToLower(strings.TrimSpace(key))
    value = strings.ToLower(strings.TrimSpace(value))
    fmt.Println("=================")
    fmt.Println(key)
    fmt.Println(value)
    fmt.Println("=================")

    h.Set(key, value)
    //len(data)-2 because the \r\n doesn't count as consumed
    //maybe there is a problem with consuming bytes because the last \r\n shouldn't count as consumed
    //however if it only ends on one crlf it should count as consumed
    return len(data) - 2, false, nil
}
