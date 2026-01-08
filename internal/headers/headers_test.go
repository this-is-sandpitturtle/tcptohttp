package headers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//func TestRequestLineParse(t *testing.T) {

func TestParseHeaders(t *testing.T) {
    //Valid Single Header
    headers := NewHeaders()
    data := []byte("Host: localhost:42069\r\n\r\n")
    n, done, err := headers.Parse(data)
    require.NoError(t, err)
    require.NotNil(t, headers)
    assert.Equal(t, "localhost:42069", headers["host"])
    assert.Equal(t, 25, n)
    assert.True(t, done)
    
    //Valid Header with extra whitespace
    headers = NewHeaders()
    data = []byte("       Host: localhost:42069       \r\n\r\n")
    n, done, err = headers.Parse(data)
    require.NoError(t, err)
    require.NotNil(t, headers)
    assert.Equal(t, "localhost:42069", headers["host"])
    assert.Equal(t, 39, n)
    assert.True(t, done)

    //Valid 2 Headers
    headers = NewHeaders()
    data = []byte("Host: localhost:42069\r\n")
    n, done, err = headers.Parse(data)
    require.NoError(t, err)
    require.NotNil(t, headers)
    assert.Equal(t, "localhost:42069", headers["host"])
    assert.Equal(t, 23, n)
    assert.False(t, done)

    data = []byte("Accept: */*\r\n\r\n")
    n, done, err = headers.Parse(data)
    require.NoError(t, err)
    require.NotNil(t, headers)
    assert.Equal(t, "*/*", headers["accept"])
    assert.Equal(t, 15, n)
    assert.True(t, done)
    assert.Equal(t, 2, len(headers))

    // Valid done
    data = []byte("\r\n and some other things")
    n, done, err = headers.Parse(data)
    require.NoError(t, err)
    require.NotNil(t, headers)
    assert.Equal(t, "localhost:42069", headers["host"])
    assert.Equal(t, "*/*", headers["accept"])
    assert.True(t, done)
    assert.Equal(t, 2, n)

    //Multiple Values for Headers
    headers = NewHeaders()
    headers["multiple-values"] = "go"
    data = []byte("multiple-values: rust\r\n\r\n")
    n, done, err = headers.Parse(data)
    require.NoError(t, err)
    require.NotNil(t, headers)
    assert.Equal(t, "go, rust", headers["multiple-values"])
    assert.Equal(t, 25, n)
    assert.True(t, done)
    assert.Equal(t, 1, len(headers))

    //Invalid spacing Header
    headers = NewHeaders()
    data = []byte("       Host : localhost:42069       \r\n\r\n")
    n, done, err = headers.Parse(data)
    require.Error(t, err)
    assert.Equal(t, 0, n)
    assert.False(t, done)

    //Invalid Header Signs
    headers = NewHeaders()
    data = []byte("H@st: localhost:42069       \r\n\r\n")
    n, done, err = headers.Parse(data)
    require.Error(t, err)
    assert.Equal(t, 0, n)
    assert.False(t, done)
    
    //Invalid Header Signs
    headers = NewHeaders()
    data = []byte(": localhost:42069       \r\n\r\n")
    n, done, err = headers.Parse(data)
    require.Error(t, err)
    assert.Equal(t, 0, n)
    assert.False(t, done)

    //Invalid Header Signs
    headers = NewHeaders()
    data = []byte("just some random bs\r\n")
    n, done, err = headers.Parse(data)
    require.Error(t, err)
    assert.Equal(t, 0, n)
    assert.False(t, done)

}
