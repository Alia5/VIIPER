package auth

import (
	"bufio"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	apitypes "github.com/Alia5/VIIPER/apitypes"
	apierror "github.com/Alia5/VIIPER/internal/server/api/error"
)

const (
	HandshakeMagic = "eVI1\x00"
	NonceSize      = 32
	authContext    = "VIIPER-Auth-v1"
)

// ReadClientNonce reads client nonce from handshake
// Expects handshake magic already consumed, reads only the 32-byte nonce
func ReadClientNonce(r io.Reader) (clientNonce []byte, err error) {
	clientNonce = make([]byte, NonceSize)
	if _, err = io.ReadFull(r, clientNonce); err != nil {
		return nil, fmt.Errorf("read client nonce: %w", err)
	}
	return clientNonce, nil
}

// WriteServerHandshake generates server nonce and sends response
// Sends: "OK\0" + server_nonce[32]
func WriteServerHandshake(w io.Writer) (serverNonce []byte, err error) {
	if w == nil {
		return nil, fmt.Errorf("write response: write on nil pointer")
	}
	serverNonce = make([]byte, NonceSize)
	if _, err = rand.Read(serverNonce); err != nil {
		return nil, fmt.Errorf("generate server nonce: %w", err)
	}

	response := append([]byte("OK\x00"), serverNonce...)
	if _, err = w.Write(response); err != nil {
		return nil, fmt.Errorf("write response: %w", err)
	}

	return serverNonce, nil
}

// IsAuthHandshake checks if the next bytes in reader match the handshake magic
func IsAuthHandshake(r *bufio.Reader) (bool, error) {
	b, err := r.Peek(len(HandshakeMagic))
	if err != nil {
		return false, err
	}
	return string(b) == HandshakeMagic, nil
}

// HandleAuthHandshake performs the authentication handshake
func HandleAuthHandshake(r *bufio.Reader, w io.Writer, key []byte, isClient bool) (clientNonce, serverNonce []byte, err error) {
	if r == nil {
		return nil, nil, fmt.Errorf("handshake: nil reader")
	}
	if len(key) == 0 {
		return nil, nil, fmt.Errorf("handshake: missing key")
	}

	if isClient {
		if w == nil {
			return nil, nil, fmt.Errorf("handshake: nil writer")
		}
		clientNonce = make([]byte, NonceSize)
		if _, err := rand.Read(clientNonce); err != nil {
			return nil, nil, fmt.Errorf("generate client nonce: %w", err)
		}

		mac := hmac.New(sha256.New, key)
		_, _ = mac.Write([]byte(authContext))
		_, _ = mac.Write(clientNonce)
		clientAuth := mac.Sum(nil)

		msg := append([]byte(HandshakeMagic), clientNonce...)
		msg = append(msg, clientAuth...)
		if _, err := w.Write(msg); err != nil {
			return nil, nil, fmt.Errorf("write handshake: %w", err)
		}

		respPrefix := make([]byte, 3)
		if _, err := io.ReadFull(r, respPrefix); err != nil {
			return nil, nil, fmt.Errorf("read handshake response: %w", err)
		}
		if string(respPrefix) != "OK\x00" {
			rest, _ := io.ReadAll(r)
			raw := append(respPrefix, rest...)
			line := strings.TrimSuffix(string(raw), "\n")

			var apiErr apitypes.ApiError
			if err := json.Unmarshal([]byte(line), &apiErr); err == nil && (apiErr.Status != 0 || apiErr.Title != "") {
				return nil, nil, &apiErr
			}
			return nil, nil, fmt.Errorf("invalid handshake response from server: %s", line)
		}

		serverNonce = make([]byte, NonceSize)
		if _, err := io.ReadFull(r, serverNonce); err != nil {
			return nil, nil, fmt.Errorf("read server nonce: %w", err)
		}
		return clientNonce, serverNonce, nil
	}

	_, err = r.Discard(len(HandshakeMagic))
	if err != nil {
		return nil, nil, fmt.Errorf("discard handshake magic: %w", err)
	}

	clientNonce, err = ReadClientNonce(r)
	if err != nil {
		return nil, nil, err
	}

	clientAuth := make([]byte, sha256.Size)
	if _, err := io.ReadFull(r, clientAuth); err != nil {
		return nil, nil, fmt.Errorf("read client auth: %w", err)
	}

	mac := hmac.New(sha256.New, key)
	_, _ = mac.Write([]byte(authContext))
	_, _ = mac.Write(clientNonce)
	expectedAuth := mac.Sum(nil)
	if !hmac.Equal(clientAuth, expectedAuth) {
		return nil, nil, apierror.ErrUnauthorized("invalid password")
	}

	serverNonce, err = WriteServerHandshake(w)
	if err != nil {
		return nil, nil, err
	}

	return clientNonce, serverNonce, nil
}
