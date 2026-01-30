package apiclient_test

import (
	"bufio"
	"encoding/json"
	"errors"
	"log/slog"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/Alia5/VIIPER/apiclient"
	apitypes "github.com/Alia5/VIIPER/apitypes"
	"github.com/Alia5/VIIPER/internal/server/api/auth"

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
		var tmp [1]byte
		for {
			conn.SetReadDeadline(time.Now().Add(2 * time.Second))
			_, rerr := conn.Read(tmp[:])
			if rerr != nil {
				break
			}
			b := tmp[0]
			buf = append(buf, b)
			if b == '\x00' {
				break
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
			expectedLine: "echo\x00",
		},
		{
			name:         "empty string payload",
			payload:      "",
			expectedLine: "echo\x00",
		},
		{
			name:         "bytes payload",
			payload:      []byte("rawbytes"),
			expectedLine: "echo rawbytes\x00",
		},
		{
			name:         "string payload",
			payload:      "hello world",
			expectedLine: "echo hello world\x00",
		},
		{
			name:         "string payload with newline",
			payload:      "multi\nline",
			expectedLine: "echo multi\nline\x00",
		},
		{
			name:         "struct payload json marshaled",
			payload:      S{A: 7, B: "zzz"},
			validateJSON: true,
		},
		{
			name:         "multi-line JSON string payload",
			payload:      "{\n\"x\":1\n}",
			expectedLine: "echo {\n\"x\":1\n}\x00",
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
			expectedPrefix := "echo " + string(b) + "\x00"
			assert.Equal(t, expectedPrefix, *got, tc.name)
			line := strings.TrimSuffix(strings.TrimPrefix(*got, "echo "), "\x00")
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
		for {
			conn.SetReadDeadline(time.Now().Add(2 * time.Second))
			_, err := conn.Read(tmp)
			if err != nil {
				break
			}
			b := tmp[0]
			buf = append(buf, b)
			if b == '\x00' { // end of request
				break
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

func TestEncryptedTransport(t *testing.T) {
	type testCase struct {
		name          string
		password      string
		serverHandler func(t *testing.T, conn net.Conn)
		line          string
		expectedErr   error
	}

	echoHandler := func(t *testing.T, conn net.Conn) {
		defer conn.Close()
		r := bufio.NewReader(conn)

		key, err := auth.DeriveKey("test123")
		assert.NoError(t, err)

		clientNonce, serverNonce, err := auth.HandleAuthHandshake(r, conn, key, false)
		if err != nil {
			var apiErr apitypes.ApiError
			if errors.As(err, &apiErr) {
				b, err := json.Marshal(apiErr)
				if err != nil {
					slog.Error("failed to marshal api error", "error", err)
					return
				}
				_, _ = conn.Write(append(b, '\n'))
				return
			}
			return
		}

		sessionKey := auth.DeriveSessionKey(key, serverNonce, clientNonce)
		secureConn, err := auth.WrapConn(conn, sessionKey)
		assert.NoError(t, err)

		rr := bufio.NewReader(secureConn)
		line, err := rr.ReadString('\x00')
		if err != nil {
			return
		}

		_, err = secureConn.Write([]byte(line))
		assert.NoError(t, err)
	}

	cases := []testCase{
		{
			name:          "success",
			password:      "test123",
			serverHandler: echoHandler,
			line:          "echo hi",
		},
		{
			name:          "wrong password",
			password:      "wrongpass",
			serverHandler: echoHandler,
			expectedErr:   errors.New("401 Unauthorized: invalid password"),
		},
		{
			name:     "bad handshake response",
			password: "test123",
			serverHandler: func(t *testing.T, conn net.Conn) {
				defer conn.Close()
				_, _ = conn.Write([]byte("NO\x00" + strings.Repeat("x", 32)))
			},
			expectedErr: errors.New(""),
		},
		{
			name:     "server closes early",
			password: "test123",
			serverHandler: func(t *testing.T, conn net.Conn) {
				_ = conn.Close()
			},
			expectedErr: errors.New(""),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ln, err := net.Listen("tcp", "127.0.0.1:0")
			assert.NoError(t, err)
			defer ln.Close()

			go func() {
				conn, err := ln.Accept()
				if err != nil {
					return
				}
				tc.serverHandler(t, conn)
			}()

			client := apiclient.NewTransportWithPassword(ln.Addr().String(), tc.password)
			path, payload, _ := strings.Cut(tc.line, " ")
			out, err := client.Do(path, payload, nil)

			if tc.expectedErr != nil {
				assert.Error(t, err)
				assert.ErrorContains(t, err, tc.expectedErr.Error())
				return
			}

			assert.NoError(t, err)
			resp := strings.TrimSuffix(out, "\x00")
			assert.Equal(t, tc.line, resp)
		})
	}
}
