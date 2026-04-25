package headers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHeaderParse(t *testing.T) {
	
		// Test: Valid single header ending with an empty line
		headers := NewHeaders()
		data := []byte("Host: localhost:42069\r\n\r\n")
		n, done, err := headers.Parse(data)
		require.NoError(t, err)
		require.NotNil(t, headers)
		host, ok := headers.Get("Host")
		require.True(t, ok)
		assert.Equal(t, "localhost:42069", host)
		assert.Equal(t, 25, n)
		assert.True(t, done)
	
		// Test: Mixed/Capital letters in header keys
		headers = NewHeaders()
		data = []byte("ConTent-TyPe: application/json\r\n\r\n")
		n, done, err = headers.Parse(data)
		require.NoError(t, err)

		assert.Equal(t, "application/json", headers.headers["content-type"])
		assert.True(t, done)
	
		// Test: Invalid character in header key (©)
		headers = NewHeaders()
		data = []byte("H©st: localhost:42069\r\n\r\n")
		n, done, err = headers.Parse(data)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "malformed header name") 
		assert.Equal(t, 0, n)
		assert.False(t, done)
	
		// Test: Invalid spacing header
		headers = NewHeaders()
		data = []byte("       Host : localhost:42069       \r\n\r\n")
		n, done, err = headers.Parse(data)
		require.Error(t, err)
		assert.Equal(t, 0, n)
		assert.False(t, done)

		// Test: Multiple headers with the same key combine with a comma
	headers = NewHeaders()
	
	headers.headers["set-person"] = "lane-loves-go"
	
	data = []byte("Set-Person: prime-loves-zig\r\nSet-Person: tj-loves-ocaml\r\n\r\n")
	n, done, err = headers.Parse(data)
	
	require.NoError(t, err)

	setPerson, ok := headers.Get("set-person")
	require.True(t, ok)
	assert.Equal(t, "lane-loves-go,prime-loves-zig,tj-loves-ocaml", setPerson)
	assert.True(t, done)
	}

