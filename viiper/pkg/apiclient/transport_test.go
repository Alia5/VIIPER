package apiclient_test

import (
	"encoding/json"
	"net"
	"strings"
	"testing"
	"time"

	"viiper/pkg/apiclient"

	"github.com/stretchr/testify/assert"
)

func startTestServer(t *testing.T, response string) (addr string, gotReqLine *string, closeFn func()) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	assert.NoError(t, err)
	got := new(string)
	go func() {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		var buf []byte
		newlineCount := 0
		var tmp [1]byte
		for {
			conn.SetReadDeadline(time.Now().Add(2 * time.Second))
			_, rerr := conn.Read(tmp[:])
			if rerr != nil {
				break
			}
			b := tmp[0]
			buf = append(buf, b)
			if b == '\n' {
				newlineCount++
				if newlineCount == 2 {
					break
				}
			} else {
				newlineCount = 0
			}
		}
		*got = string(buf)
		if response != "" {
			_, _ = conn.Write([]byte(response))
		}
	}()
	return ln.Addr().String(), got, func() { _ = ln.Close() }
}

func TestTransportPayloadEncoding(t *testing.T) {
	type S struct {
		A int    `json:"a"`
		B string `json:"b"`
	}
	type testCase struct {
		name         string
		payload      any
		expectedLine string // full request including terminator (for non-struct where deterministic)
		validateJSON bool   // whether to JSON-unmarshal payload part instead of direct equality
	}

	cases := []testCase{
		{
			name:         "nil payload",
			payload:      nil,
			expectedLine: "echo\n\n",
		},
		{
			name:         "empty string payload",
			payload:      "",
			expectedLine: "echo\n\n",
		},
		{
			name:         "bytes payload",
			payload:      []byte("rawbytes"),
			expectedLine: "echo rawbytes\n\n",
		},
		{
			name:         "string payload",
			payload:      "hello world",
			expectedLine: "echo hello world\n\n",
		},
		{
			name:         "string payload with newline",
			payload:      "multi\nline",
			expectedLine: "echo multi\nline\n\n",
		},
		{
			name:         "struct payload json marshaled",
			payload:      S{A: 7, B: "zzz"},
			validateJSON: true,
		},
		{
			name:         "multi-line JSON string payload",
			payload:      "{\n\"x\":1\n}",
			expectedLine: "echo {\n\"x\":1\n}\n\n",
		},
	}

	for _, tc := range cases {
		addr, got, closeFn := startTestServer(t, "ok\n")
		client := apiclient.NewTransport(addr)
		out, err := client.Do("echo", tc.payload, nil)
		closeFn()
		assert.NoError(t, err, tc.name)
		assert.Equal(t, "ok", out, tc.name)

		if tc.validateJSON {
			b, merr := json.Marshal(tc.payload)
			assert.NoError(t, merr, tc.name)
			expectedPrefix := "echo " + string(b) + "\n\n"
			assert.Equal(t, expectedPrefix, *got, tc.name)
			line := strings.TrimSuffix(strings.TrimPrefix(*got, "echo "), "\n\n")
			var s S
			assert.NoError(t, json.Unmarshal([]byte(line), &s), tc.name)
			assert.Equal(t, tc.payload, s, tc.name)
			continue
		}

		assert.Equal(t, tc.expectedLine, *got, tc.name)
	}
}

func TestTransportMultiLineResponse(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	assert.NoError(t, err)
	defer ln.Close()

	resp := "{\n  \"a\": 1,\n  \"b\": 2\n}\n" // multi-line + trailing newline

	go func() {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		buf := make([]byte, 0, 128)
		tmp := make([]byte, 1)
		newlineCount := 0
		for {
			conn.SetReadDeadline(time.Now().Add(2 * time.Second))
			_, err := conn.Read(tmp)
			if err != nil {
				break
			}
			b := tmp[0]
			buf = append(buf, b)
			if b == '\n' {
				newlineCount++
				if newlineCount == 2 { // end of request
					break
				}
			} else {
				newlineCount = 0
			}
		}
		_, _ = conn.Write([]byte(resp))
		conn.Close()
	}()

	client := apiclient.NewTransport(ln.Addr().String())
	out, err := client.Do("echo", nil, nil)
	assert.NoError(t, err)
	assert.Equal(t, "{\n  \"a\": 1,\n  \"b\": 2\n}", out)
}
