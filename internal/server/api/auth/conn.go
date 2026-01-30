package auth

import (
	"bytes"
	"crypto/cipher"
	"encoding/binary"
	"io"
	"net"
	"sync"

	"golang.org/x/crypto/chacha20poly1305"
)

type Conn struct {
	net.Conn
	aead    cipher.AEAD
	sendCtr uint64
	recvBuf bytes.Buffer
	mu      sync.Mutex
}

const maxPacketSize = 2 * 1024 * 1024 // 2 MB

func WrapConn(conn net.Conn, sessionKey []byte) (net.Conn, error) {
	aead, err := chacha20poly1305.New(sessionKey)
	if err != nil {
		return nil, err
	}
	return &Conn{Conn: conn, aead: aead}, nil
}

func (s *Conn) Write(p []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	nonce := make([]byte, 12)
	binary.BigEndian.PutUint64(nonce[4:], s.sendCtr)
	s.sendCtr++

	ct := s.aead.Seal(nil, nonce, p, nil)
	length := uint32(len(nonce) + len(ct))

	var hdr [4]byte
	binary.BigEndian.PutUint32(hdr[:], length)

	if i, err := s.Conn.Write(hdr[:]); err != nil {
		return i, err
	}
	if i, err := s.Conn.Write(nonce); err != nil {
		return i, err
	}
	if i, err := s.Conn.Write(ct); err != nil {
		return i, err
	}

	return len(p), nil
}

func (s *Conn) Read(p []byte) (int, error) {
	if s.recvBuf.Len() == 0 {
		var hdr [4]byte
		if i, err := io.ReadFull(s.Conn, hdr[:]); err != nil {
			return i, err
		}
		length := binary.BigEndian.Uint32(hdr[:])
		if length > maxPacketSize {
			return 0, io.ErrUnexpectedEOF
		}

		pkt := make([]byte, length)
		if i, err := io.ReadFull(s.Conn, pkt); err != nil {
			return i, err
		}

		nonce := pkt[:12]
		ct := pkt[12:]

		pt, err := s.aead.Open(nil, nonce, ct, nil)
		if err != nil {
			return 0, err
		}

		s.recvBuf.Write(pt)
	}
	return s.recvBuf.Read(p)
}
