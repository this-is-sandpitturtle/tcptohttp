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
    r, err := RequestFromReader(reader)
    fmt.Println("Parsed Request")
    fmt.Println(r)
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
        //numBytesPerRead: len(reader.data),
        numBytesPerRead: len(reader.data),
    }
    //fmt.Println(reader.data)
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

    //Empty Headers
    reader = &chunkReader{
        data: "GET / HTTP/1.1\r\n\r\n",
        numBytesPerRead: 3,
    }
    r, err = RequestFromReader(reader)
    fmt.Println(r)
    require.NoError(t, err)
    require.NotNil(t, r)
    assert.Equal(t, 0, len(r.Headers))

    //Duplicate Headers
    reader = &chunkReader{
        data: "GET / HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\nAccept: application/json\r\n\r\n",
        numBytesPerRead: 10,
    }

    r, err = RequestFromReader(reader)
    fmt.Println(r)
    require.NoError(t, err)
    require.NotNil(t, r)
    assert.Equal(t, "localhost:42069", r.Headers["host"])
    assert.Equal(t, "curl/7.81.0", r.Headers["user-agent"])
    assert.Equal(t, "*/*, application/json", r.Headers["accept"])

    //Case Insensitive Headers:
    reader = &chunkReader{
        data: "GET / HTTP/1.1\r\nHOST: locALHost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\nAccept: application/json\r\n\r\n",
        numBytesPerRead: 10,
    }

    r, err = RequestFromReader(reader)
    fmt.Println(r)
    require.NoError(t, err)
    require.NotNil(t, r)
    assert.Equal(t, "localhost:42069", r.Headers["host"])
    assert.Equal(t, "curl/7.81.0", r.Headers["user-agent"])
    assert.Equal(t, "*/*, application/json", r.Headers["accept"])
    assert.Equal(t, "", r.Headers["HOST"])


    // Test: Malformed Header
    reader = &chunkReader{
        data: "GET / HTTP/1.1\r\nHost localhost:42069\r\n\r\n",
        numBytesPerRead: 3,
    }

    r, err = RequestFromReader(reader)
    require.Error(t, err)

    //Missing End of Headers
    reader = &chunkReader{
        data: "GET / HTTP/1.1\r\nHost localhost:42069\r\n",
        numBytesPerRead: 8,
    }

    r, err = RequestFromReader(reader)
    require.Error(t, err)

}


func TestParseBody(t *testing.T) {
    // Test: Standard Body
    reader := &chunkReader{
        data: "POST /submit HTTP/1.1\r\n" +
        "Host: localhost:42069\r\n" +
        "Content-Length: 13\r\n" +
        "\r\n" +
        "hello world!\n",
        numBytesPerRead: 3,
    }
    r, err := RequestFromReader(reader)
    require.NoError(t, err)
    require.NotNil(t, r)
    assert.Equal(t, "hello world!\n", string(r.Body))

    //Test different order of header parameters
    //TODO: BREAKS FOR HIGHER CHUNK SIZES
    reader = &chunkReader{
        data: "POST /submittest HTTP/1.1\r\n" +
        "Content-Length: 13\r\n" +
        "Host: localhost:42069\r\n" +
        "\r\n" +
        "hello world!\n",
        numBytesPerRead: len(reader.data),
    }
    fmt.Println("-----------------------------------------------------------------------------------")
    fmt.Println(reader.numBytesPerRead)
    fmt.Println(r)
    r, err = RequestFromReader(reader)
    require.NoError(t, err)
    require.NotNil(t, r)
    assert.Equal(t, "13", r.Headers["content-length"])
    assert.Equal(t, "localhost:42069", r.Headers["host"])
    assert.Equal(t, "hello world!\n", string(r.Body))

    //Test: Empty Body Content-Legnth 0
    reader = &chunkReader{
        data: "POST /submit HTTP/1.1\r\n" +
        "Host: localhost:42069\r\n" +
        "Content-Length: 0\r\n" +
        "\r\n",
        numBytesPerRead: 10, 
    }
    r, err = RequestFromReader(reader)
    require.NoError(t, err)
    require.NotNil(t, r)
    assert.Equal(t, "", string(r.Body))

    //Test: Empty Body no reported content-length
    reader = &chunkReader{
        data: "POST /submit HTTP/1.1\r\n" +
        "Host: localhost:42069\r\n" +
        "\r\n",
        numBytesPerRead: 10,
    }
    r, err = RequestFromReader(reader)
    require.NoError(t, err)
    require.NotNil(t, r)
    assert.Equal(t, "", string(r.Body))

    // Test: Body shorter than reported content length
    reader = &chunkReader{
        data: "POST /submit HTTP/1.1\r\n" +
        "Host: localhost:42069\r\n" +
        "Content-Length: 20\r\n" +
        "\r\n" +
        "partial content",
        numBytesPerRead: 3,
    }
    fmt.Println("-------------------------------------------------")
    r, err = RequestFromReader(reader)
    require.Error(t, err)
    
    // Test: Body shorter than reported content length
    fmt.Println("-------------------------------------------------")
    reader = &chunkReader{
        data: "POST /submit HTTP/1.1\r\n" +
        "Host: localhost:42069\r\n" +
        "\r\n" +
        "body exists",
        numBytesPerRead: 3,
    }
    r, err = RequestFromReader(reader)
    require.NoError(t, err)
    require.NotNil(t, r)
    assert.Equal(t, "", string(r.Body))

}
