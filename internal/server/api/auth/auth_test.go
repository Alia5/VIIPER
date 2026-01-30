package auth_test

import (
	"errors"
	"testing"

	"github.com/Alia5/VIIPER/internal/server/api/auth"
	"github.com/stretchr/testify/assert"
)

func TestGenKey(t *testing.T) {

	key, err := auth.GenerateKey()
	assert.NoError(t, err)
	assert.Len(t, key, auth.AutoGenKeyLength)
	assert.Regexp(t, "^[0-9A-Za-z]{16}$", key)

}

func BenchmarkGenKey(b *testing.B) {
	var key string
	var err error
	for b.Loop() {
		key, err = auth.GenerateKey()
	}
	assert.NoError(b, err)
	assert.Len(b, key, auth.AutoGenKeyLength)
}

func TestDeriveKey(t *testing.T) {

	type testCase struct {
		name        string
		password    string
		expectedKey []byte
		expectedErr error
	}

	testCases := []testCase{
		{
			name:        "Normal Password",
			password:    "password123",
			expectedKey: []byte{0x94, 0x50, 0x29, 0x55, 0x1, 0xd7, 0x3, 0xf, 0x4, 0x61, 0xf, 0x81, 0x6a, 0xdf, 0x43, 0x1c, 0xaf, 0x8f, 0xc8, 0x21, 0xd4, 0xc1, 0x2f, 0x2f, 0x21, 0x2c, 0x1b, 0xf8, 0x64, 0x46, 0x9, 0x82},
		},
		{
			name:        "Simple Password",
			password:    "1",
			expectedKey: []byte{0xfe, 0xdf, 0xdf, 0x4d, 0xab, 0xd2, 0x5d, 0x9f, 0xfd, 0x97, 0x96, 0xec, 0x76, 0xd2, 0xa2, 0xec, 0x2, 0x4f, 0xbf, 0xeb, 0x17, 0x8c, 0x6, 0x13, 0xed, 0x4f, 0x10, 0x9e, 0x4d, 0xef, 0xd1, 0xd2},
		},
		{
			name:        "empty password",
			password:    "",
			expectedKey: []byte{},
			expectedErr: errors.New("Password cannot be empty"),
		},
		{
			name:        "long password",
			password:    "dkfghdfg90d78h350ß8dgfjkdfg#---23489dfg!!!@!@#$$%&/()=",
			expectedKey: []byte{0xb4, 0xb9, 0xf5, 0x37, 0xa6, 0xac, 0x8a, 0x35, 0xc5, 0xe7, 0x1a, 0x90, 0xf9, 0x7e, 0xab, 0x38, 0x22, 0x83, 0xd8, 0xc6, 0xa, 0xcf, 0xbf, 0x7c, 0x3d, 0xd6, 0x6c, 0xd4, 0x35, 0x3b, 0x88, 0x2c},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			derivedKey, err := auth.DeriveKey(tc.password)
			if tc.expectedErr != nil {
				assert.Equal(t, tc.expectedErr, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tc.expectedKey, derivedKey)
		})
	}
}

func TestDeriveSessionKey(t *testing.T) {
	key := make([]byte, 32)
	serverNonce := make([]byte, 32)
	clientNonce := make([]byte, 32)

	for i := range key {
		key[i] = byte(i)
		serverNonce[i] = byte(i + 10)
		clientNonce[i] = byte(i + 20)
	}

	sessionKey := auth.DeriveSessionKey(key, serverNonce, clientNonce)
	assert.Len(t, sessionKey, 32)

	sessionKey2 := auth.DeriveSessionKey(key, serverNonce, clientNonce)
	assert.Equal(t, sessionKey, sessionKey2)

	clientNonce[0] = 99
	sessionKey3 := auth.DeriveSessionKey(key, serverNonce, clientNonce)
	assert.NotEqual(t, sessionKey, sessionKey3)
}
