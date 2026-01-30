package auth_test

import (
	"bufio"
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
	"io"
	"testing"

	"github.com/Alia5/VIIPER/internal/server/api/auth"
	apierror "github.com/Alia5/VIIPER/internal/server/api/error"
	"github.com/stretchr/testify/assert"
)

func TestReadClientNonce(t *testing.T) {
	type testCase struct {
		name          string
		input         []byte
		expectedNonce []byte
		expectedErr   error
	}

	validNonce := make([]byte, 32)
	for i := range validNonce {
		validNonce[i] = byte(i)
	}

	testCases := []testCase{
		{
			name:          "Valid nonce",
			input:         validNonce,
			expectedNonce: validNonce,
			expectedErr:   nil,
		},
		{
			name:          "Short input",
			input:         []byte{1, 2, 3},
			expectedNonce: nil,
			expectedErr:   fmt.Errorf("read client nonce: unexpected EOF"),
		},
		{
			name:          "Empty input",
			input:         []byte{},
			expectedNonce: nil,
			expectedErr:   fmt.Errorf("read client nonce: EOF"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			buf := bytes.NewBuffer(tc.input)
			nonce, err := auth.ReadClientNonce(buf)

			if tc.expectedErr != nil {
				assert.Error(t, err)
				assert.EqualError(t, err, tc.expectedErr.Error())
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tc.expectedNonce, nonce)
		})
	}
}

func TestWriteServerHandshake(t *testing.T) {
	type testCase struct {
		name        string
		writer      io.Writer
		expectedErr error
	}

	testCases := []testCase{
		{
			name:        "Success",
			writer:      bytes.NewBuffer(nil),
			expectedErr: nil,
		},
		{
			name:        "Err no writer",
			writer:      nil,
			expectedErr: fmt.Errorf("write response: write on nil pointer"),
		},
		{
			name: "Err closed writer",
			writer: func() io.Writer {
				_, w := io.Pipe()
				w.Close()
				return w
			}(),
			expectedErr: fmt.Errorf("write response: io: read/write on closed pipe"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			serverNonce, err := auth.WriteServerHandshake(tc.writer)

			if tc.expectedErr != nil {
				assert.Error(t, err)
				assert.EqualError(t, err, tc.expectedErr.Error())
				return
			}

			assert.NoError(t, err)
			assert.Len(t, serverNonce, 32)

			buf := tc.writer.(*bytes.Buffer)
			expectedResponse := buf.Bytes()
			assert.Equal(t, "OK\x00", string(expectedResponse[:3]))
			assert.Equal(t, serverNonce, expectedResponse[3:])
			assert.Len(t, expectedResponse, 35)
		})
	}
}

func TestIsAuthHandshake(t *testing.T) {
	type testCase struct {
		name           string
		input          *bufio.Reader
		expectedResult bool
		expectedErr    error
	}
	testCases := []testCase{
		{
			name:           "IS_AUTH",
			input:          bufio.NewReader(bytes.NewBuffer([]byte(auth.HandshakeMagic))),
			expectedResult: true,
			expectedErr:    nil,
		},
		{
			name:           "NOT_AUTH",
			input:          bufio.NewReader(bytes.NewBuffer([]byte("HEsdffdLLO\x00"))),
			expectedResult: false,
			expectedErr:    nil,
		},
		{
			name:           "ERR_INCOMPLETE",
			input:          bufio.NewReader(bytes.NewBuffer([]byte("eV"))),
			expectedResult: false,
			expectedErr:    fmt.Errorf("EOF"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := auth.IsAuthHandshake(tc.input)
			if tc.expectedErr != nil {
				assert.Error(t, err)
				assert.EqualError(t, err, tc.expectedErr.Error())
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tc.expectedResult, result)
		})
	}

}

func TestFullHandshake(t *testing.T) {

	type testCase struct {
		name        string
		reader      *bufio.Reader
		writer      io.Writer
		key         []byte
		isClient    bool
		expectedErr error
	}

	validKey, err := auth.DeriveKey("test123")
	assert.NoError(t, err)
	wrongKey, err := auth.DeriveKey("wrongpass")
	assert.NoError(t, err)

	clientNonce := make([]byte, 32)
	for i := range clientNonce {
		clientNonce[i] = byte(i)
	}
	mac := hmac.New(sha256.New, validKey)
	_, _ = mac.Write([]byte("VIIPER-Auth-v1"))
	_, _ = mac.Write(clientNonce)
	clientAuth := mac.Sum(nil)

	validHandshake := append([]byte(auth.HandshakeMagic), clientNonce...)
	validHandshake = append(validHandshake, clientAuth...)
	testCases := []testCase{
		{
			name:        "Successful Handshake",
			reader:      bufio.NewReader(bytes.NewBuffer(validHandshake)),
			writer:      bytes.NewBuffer(nil),
			key:         validKey,
			isClient:    false,
			expectedErr: nil,
		},
		{
			name:        "Err reading client nonce",
			reader:      bufio.NewReader(bytes.NewBuffer(append([]byte(auth.HandshakeMagic), []byte("short")...))),
			writer:      bytes.NewBuffer(nil),
			key:         validKey,
			isClient:    false,
			expectedErr: fmt.Errorf("read client nonce: unexpected EOF"),
		},
		{
			name:        "Err writing server handshake",
			reader:      bufio.NewReader(bytes.NewBuffer(validHandshake)),
			writer:      nil,
			key:         validKey,
			isClient:    false,
			expectedErr: fmt.Errorf("write response: write on nil pointer"),
		},
		{
			name:        "Err discarding handshake magic",
			reader:      bufio.NewReader(bytes.NewBuffer([]byte("sh"))),
			writer:      bytes.NewBuffer(nil),
			key:         validKey,
			isClient:    false,
			expectedErr: fmt.Errorf("discard handshake magic: EOF"),
		},
		{
			name:   "Err closed writer",
			reader: bufio.NewReader(bytes.NewBuffer(validHandshake)),
			writer: func() io.Writer {
				_, w := io.Pipe()
				w.Close()
				return w
			}(),
			key:         validKey,
			isClient:    false,
			expectedErr: fmt.Errorf("write response: io: read/write on closed pipe"),
		},
		{
			name:        "Err Tried unauthenticated handshake",
			reader:      bufio.NewReader(bytes.NewBuffer([]byte("NOT_A_HANDSHAKE"))),
			writer:      bytes.NewBuffer(nil),
			key:         validKey,
			isClient:    false,
			expectedErr: fmt.Errorf("read client nonce: unexpected EOF"),
		},
		{
			name:        "Err invalid password",
			reader:      bufio.NewReader(bytes.NewBuffer(validHandshake)),
			writer:      bytes.NewBuffer(nil),
			key:         wrongKey,
			isClient:    false,
			expectedErr: apierror.ErrUnauthorized("invalid password"),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			clientNonce, serverNonce, err := auth.HandleAuthHandshake(
				tc.reader,
				tc.writer,
				tc.key,
				tc.isClient,
			)
			if tc.expectedErr != nil {
				assert.Error(t, err)
				assert.EqualError(t, err, tc.expectedErr.Error())
				return
			}
			assert.NoError(t, err)
			assert.Len(t, clientNonce, 32)
			assert.Len(t, serverNonce, 32)
		})
	}

}
