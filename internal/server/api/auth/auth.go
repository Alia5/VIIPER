package auth

import (
	"crypto/pbkdf2"
	"crypto/rand"
	"crypto/sha256"
	"errors"
)

const (
	AutoGenKeyLength = 16
	Base62Chars      = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	PBKDF2Iterations = 100000
	PBKDF2Salt       = "VIIPER-Key-v1"
)

// GenerateKey creates a random 16-char base62 key
func GenerateKey() (string, error) {
	randomBytes := make([]byte, AutoGenKeyLength)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}

	key := make([]byte, AutoGenKeyLength)
	for i, b := range randomBytes {
		key[i] = Base62Chars[int(b)%62]
	}

	return string(key), nil
}

// DeriveKey uses PBKDF2 to stretch any password to 32 bytes
func DeriveKey(password string) ([]byte, error) {
	if password == "" {
		return nil, errors.New("Password cannot be empty")
	}
	return pbkdf2.Key(
		sha256.New,
		password,
		[]byte(PBKDF2Salt),
		PBKDF2Iterations,
		32,
	)
}

// DeriveSessionKey creates unique session key from key and nonces
// SHA mixing is used for easier client implementations
func DeriveSessionKey(key, serverNonce, clientNonce []byte) []byte {
	h := sha256.New()
	h.Write(key)
	h.Write(serverNonce)
	h.Write(clientNonce)
	h.Write([]byte("VIIPER-Session-v1"))
	return h.Sum(nil)
}
