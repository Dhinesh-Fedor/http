package headers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHeaderParse(t *testing.T) {
	// Test: Valid single header
	headers := NewHeaders()
	data := []byte("Host: localhost:42069\r\nLaaLaa:  baraaaa  \r\n\r\n")
	n, done, err := headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	host, ok := headers.Get("Host")
	assert.True(t,ok)
	assert.Equal(t, "localhost:42069", host)
	LaaLaa, ok := headers.Get("LaaLaa")
	assert.True(t,ok)
	assert.Equal(t, "baraaaa", LaaLaa)
	assert.Equal(t, 45, n)
	assert.True(t, done)

	// Test: Invalid spacing header
	headers = NewHeaders()
	data = []byte("H©st: localhost:42069\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)
	assert.False(t, done)
	
	headers = NewHeaders()
	data = []byte("Host: localhost:42069\r\nHost: localhost:42069\r\n" )
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	host1, ok := headers.Get("HOST")
	assert.True(t,ok)
	assert.Equal(t, "localhost:42069,localhost:42069", host1)
	assert.False(t, done)
}
