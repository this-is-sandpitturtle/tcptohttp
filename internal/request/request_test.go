package request

import (
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type chunkReader struct {
	data            string
	numBytesPerRead int
	pos             int
}

// Read reads up to len(p) or numBytesPerRead bytes from the string per call
// its useful for simulating reading a variable number of bytes per chunk from a network connection
func (cr *chunkReader) Read(p []byte) (n int, err error) {
	if cr.pos >= len(cr.data) {
		return 0, io.EOF
	}
	endIndex := cr.pos + cr.numBytesPerRead
	if endIndex > len(cr.data) {
		endIndex = len(cr.data)
	}
	n = copy(p, cr.data[cr.pos:endIndex])
	cr.pos += n

	return n, nil
}

func TestRequestLineParse(t *testing.T) {
    reader := &chunkReader{
        data: "GET / HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
        numBytesPerRead: 3,
    }

    // Happy GET-Request | RequestLine
    fmt.Println(reader.data)
    r, err := RequestFromReader(reader)
    require.NoError(t, err)
    require.NotNil(t, r)
    assert.Equal(t, "GET", r.RequestLine.Method)
    assert.Equal(t, "/", r.RequestLine.RequestTarget)
    assert.Equal(t, "1.1", r.RequestLine.HttpVersion)

    fmt.Println(r)

    reader = &chunkReader{
        data: "GET /coffee HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
        numBytesPerRead: 1,
    }
    fmt.Println(reader.data)
    // Happy GET-Request | RequestLine with path
    r, err = RequestFromReader(reader)
    require.NoError(t, err)
    require.NotNil(t, r)
    assert.Equal(t, "GET", r.RequestLine.Method)
    assert.Equal(t, "/coffee", r.RequestLine.RequestTarget)
    assert.Equal(t, "1.1", r.RequestLine.HttpVersion)

    reader = &chunkReader{
        data: "POST /coffee HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
        numBytesPerRead: len(reader.data),
    }
    fmt.Println(reader.data)
    // Happy POST-Request | RequestLine with path
    r, err = RequestFromReader(reader)
    require.NoError(t, err)
    require.NotNil(t, r)
    assert.Equal(t, "POST", r.RequestLine.Method)
    assert.Equal(t, "/coffee", r.RequestLine.RequestTarget)
    assert.Equal(t, "1.1", r.RequestLine.HttpVersion)

    // Unhappy GET-Requst | Invalid number of parts in RequestLine
    _, err = RequestFromReader(strings.NewReader("/coffee HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n"))
    require.Error(t, err)

    // Unhappy GET-Requst | Invalid HTTP-Method
    _, err = RequestFromReader(strings.NewReader("get /coffee HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n"))
    require.Error(t, err)

    // Unhappy GET-Requst | Invalid HTTP-Version
    _, err = RequestFromReader(strings.NewReader("GET /coffee HTTP/1.3\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n"))
    require.Error(t, err)

}

func TestParseHeaders(t *testing.T) {
    // Test: Standard Headers
    reader := &chunkReader{
        data: "GET / HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
        numBytesPerRead: 3,
    }

    r, err := RequestFromReader(reader)
    fmt.Println(r)
    require.NoError(t, err)
    require.NotNil(t, r)
    assert.Equal(t, "localhost:42069", r.Headers["host"])
    assert.Equal(t, "curl/7.81.0", r.Headers["user-agent"])
    assert.Equal(t, "*/*", r.Headers["accept"])

    // Test: Malformed Header
    reader = &chunkReader{
        data: "GET / HTTP/1.1\r\nHost localhost:42069\r\n\r\n",
        numBytesPerRead: 3,
    }

    r, err = RequestFromReader(reader)
    require.Error(t, err)

}

