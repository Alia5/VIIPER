package auth_test

import (
	"errors"
	"net"
	"testing"

	"github.com/Alia5/VIIPER/internal/server/api/auth"
	"github.com/stretchr/testify/assert"
)

func TestConn(t *testing.T) {

	type testCase struct {
		name        string
		wrapConn    func(net.Conn, []byte) (net.Conn, error)
		setupFn     func(clientConn net.Conn, serverConn net.Conn) (clientKey []byte, serverKey []byte)
		input       []byte
		expected    []byte
		expectedErr error
	}

	testCases := []testCase{
		{
			name:     "valid read",
			wrapConn: auth.WrapConn,
			setupFn: func(clientConn, serverConn net.Conn) (clientKey []byte, serverKey []byte) {
				password := "test123"
				key, err := auth.DeriveKey(password)
				if err != nil {
					t.Fatalf("failed to derive key: %v", err)
				}
				return key, key
			},
			input:    []byte("Hello, World!"),
			expected: []byte("Hello, World!"),
		},
		{
			name:     "Differing Keys",
			wrapConn: auth.WrapConn,
			setupFn: func(clientConn, serverConn net.Conn) (clientKey []byte, serverKey []byte) {
				key, err := auth.DeriveKey("test123")
				if err != nil {
					t.Fatalf("failed to derive key: %v", err)
				}
				key2, err := auth.DeriveKey("123test")
				if err != nil {
					t.Fatalf("failed to derive key: %v", err)
				}
				return key, key2
			},
			input:       []byte("x"),
			expected:    nil,
			expectedErr: errors.New("chacha20poly1305: message authentication failed"),
		},
		{
			name:     "bad key length (client)",
			wrapConn: auth.WrapConn,
			setupFn: func(clientConn, serverConn net.Conn) (clientKey []byte, serverKey []byte) {
				key, err := auth.DeriveKey("test123")
				if err != nil {
					t.Fatalf("failed to derive key: %v", err)
				}
				return []byte{1, 2, 3}, key // invalid key length on client
			},
			input:       []byte("x"),
			expected:    nil,
			expectedErr: errors.New("chacha20poly1305: bad key length"),
		},
		{
			name:     "bad key length (server)",
			wrapConn: auth.WrapConn,
			setupFn: func(clientConn, serverConn net.Conn) (clientKey []byte, serverKey []byte) {
				key, err := auth.DeriveKey("test123")
				if err != nil {
					t.Fatalf("failed to derive key: %v", err)
				}
				return key, []byte{1, 2, 3} // invalid key length on server
			},
			input:       []byte("x"),
			expected:    nil,
			expectedErr: errors.New("chacha20poly1305: bad key length"),
		},
		{
			name:     "client closed before write",
			wrapConn: auth.WrapConn,
			setupFn: func(clientConn, serverConn net.Conn) (clientKey []byte, serverKey []byte) {
				key, err := auth.DeriveKey("test123")
				if err != nil {
					t.Fatalf("failed to derive key: %v", err)
				}
				_ = clientConn.Close()
				return key, key
			},
			input:       []byte("x"),
			expected:    nil,
			expectedErr: errors.New("use of closed network connection"),
		},
		{
			name:     "server closed before read",
			wrapConn: auth.WrapConn,
			setupFn: func(clientConn, serverConn net.Conn) (clientKey []byte, serverKey []byte) {
				key, err := auth.DeriveKey("test123")
				if err != nil {
					t.Fatalf("failed to derive key: %v", err)
				}
				_ = serverConn.Close()
				return key, key
			},
			input:    []byte("x"),
			expected: nil,
			// just check for error, linux/win differ
			expectedErr: errors.New(""),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			ln, err := net.Listen("tcp", "127.0.0.1:0")
			if err != nil {
				t.Fatalf("failed to start test server: %v", err)
			}
			clientConn, err := net.Dial("tcp", ln.Addr().String())
			if err != nil {
				t.Fatalf("failed to connect to test server: %v", err)
			}
			serverConn, err := ln.Accept()
			if err != nil {
				t.Fatalf("failed to accept connection: %v", err)
			}
			defer ln.Close()
			defer clientConn.Close()
			defer serverConn.Close()

			var clientKey, serverKey []byte
			if tc.setupFn != nil {
				clientKey, serverKey = tc.setupFn(clientConn, serverConn)
			}

			var wrappedServerConn net.Conn
			var wrappedClientConn net.Conn
			if tc.wrapConn != nil {
				wrappedServerConn, err = tc.wrapConn(serverConn, serverKey)
				if err != nil {
					if tc.expectedErr != nil {
						assert.ErrorContains(t, err, tc.expectedErr.Error())
					} else {
						t.Fatalf("failed to wrap server conn: %v", err)
					}
					return
				}
				wrappedClientConn, err = tc.wrapConn(clientConn, clientKey)
				if err != nil {
					if tc.expectedErr != nil {
						assert.ErrorContains(t, err, tc.expectedErr.Error())
					} else {
						t.Fatalf("failed to wrap client conn: %v", err)
					}
					return
				}
			}

			_, err = wrappedClientConn.Write(tc.input)
			if err != nil {
				if tc.expectedErr != nil {
					assert.ErrorContains(t, err, tc.expectedErr.Error())
				} else {
					t.Fatalf("failed to wrap client conn: %v", err)
				}
				return
			}
			buf := make([]byte, len(tc.expected))
			_, err = wrappedServerConn.Read(buf)
			if err != nil {
				if tc.expectedErr != nil {
					assert.ErrorContains(t, err, tc.expectedErr.Error())
				} else {
					t.Errorf("server read error: %v", err)
				}
				return
			}
			assert.Equal(t, tc.expected, buf)

		})
	}

}
