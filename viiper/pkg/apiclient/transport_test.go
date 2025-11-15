package apiclient

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToPayloadBytes(t *testing.T) {
	b, ok := toPayloadBytes(nil)
	assert.True(t, ok)
	assert.Nil(t, b)

	orig := []byte{0x01, 0x02, 0x03}
	b, ok = toPayloadBytes(orig)
	assert.True(t, ok)
	assert.Equal(t, orig, b)

	b, ok = toPayloadBytes("hello")
	assert.True(t, ok)
	assert.Equal(t, []byte("hello"), b)

	type S struct {
		A int    `json:"a"`
		B string `json:"b"`
	}
	b, ok = toPayloadBytes(S{A: 5, B: "x"})
	assert.True(t, ok)
	var s S
	assert.NoError(t, json.Unmarshal(b, &s))
	assert.Equal(t, 5, s.A)
	assert.Equal(t, "x", s.B)
}
