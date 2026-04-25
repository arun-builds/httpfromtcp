package request

import (
	"io"
	
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
func TestParseHeaders(t *testing.T) {
	// Test: Standard Headers
	reader := &chunkReader{
		data:            "GET / HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
		numBytesPerRead: 3,
	}
	r, err := RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "localhost:42069", r.Headers.Get("host"))
	assert.Equal(t, "curl/7.81.0", r.Headers.Get("user-agent"))
	assert.Equal(t, "*/*", r.Headers.Get("accept"))

	// Test: Empty Headers
	reader = &chunkReader{
		data:            "GET / HTTP/1.1\r\n\r\n",
		numBytesPerRead: 5,
	}
	r, err = RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "", r.Headers.Get("host")) // Should be empty, not panic

	// Test: Malformed Header (missing colon)
	reader = &chunkReader{
		data:            "GET / HTTP/1.1\r\nHost localhost:42069\r\n\r\n",
		numBytesPerRead: 3,
	}
	_, err = RequestFromReader(reader)
	require.Error(t, err)

	// Test: Duplicate Headers
	reader = &chunkReader{
		data:            "GET / HTTP/1.1\r\nSet-Cookie: choc-chip\r\nSet-Cookie: oatmeal-raisin\r\n\r\n",
		numBytesPerRead: 4,
	}
	r, err = RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "choc-chip,oatmeal-raisin", r.Headers.Get("set-cookie"))

	// Test: Case Insensitive Headers
	reader = &chunkReader{
		data:            "GET / HTTP/1.1\r\nCaMel-CaSe: some-value\r\n\r\n",
		numBytesPerRead: 2,
	}
	r, err = RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	// Both should work assuming your Get() method uses strings.ToLower() internally
	assert.Equal(t, "some-value", r.Headers.Get("camel-case"))
	assert.Equal(t, "some-value", r.Headers.Get("CaMel-CaSe"))

	// Test: Missing End of Headers
	reader = &chunkReader{
		data:            "GET / HTTP/1.1\r\nHost: localhost:42069\r\n", // Missing final \r\n
		numBytesPerRead: 3,
	}
	_, err = RequestFromReader(reader)
	// It should trigger an error (like io.EOF) because it never reaches \r\n\r\n
	require.Error(t, err) 
}