package apiclient

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/url"
	"strings"
	"time"

	"github.com/Alia5/VIIPER/internal/server/api/auth"
	apierror "github.com/Alia5/VIIPER/internal/server/api/error"
)

// Config controls low-level transport behavior such as timeouts.
type Config struct {
	DialTimeout  time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	Password     string
}

func defaultConfig() Config {
	return Config{
		DialTimeout:  3 * time.Second,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}
}

// Transport is the low-level VIIPER management protocol implementation used by higher-level API clients.
// Request framing: `<path>[ SP <payload>] \x00` (null terminator). The payload may contain any data
// including newlines (e.g. pretty JSON, binary) because only \x00 ends the request.
// Response framing: server writes a single JSON (or empty success) line terminated by `\n` and then
// closes the connection. We therefore read until EOF (connection close) and trim a single trailing
// newline if present. Embedded newlines in the response (future multi-line responses) are preserved.
type Transport struct {
	addr string
	mock func(path string, payload any, pathParams map[string]string) (string, error)
	cfg  Config
}

// NewTransport creates a new low-level transport.
func NewTransport(addr string) *Transport { return NewTransportWithConfig(addr, nil) }

func NewTransportWithPassword(addr, password string) *Transport {
	cfg := defaultConfig()
	cfg.Password = password
	return NewTransportWithConfig(addr, &cfg)
}

// NewTransportWithConfig creates a new low-level transport with optional timeouts configuration.
func NewTransportWithConfig(addr string, cfg *Config) *Transport {
	c := defaultConfig()
	if cfg != nil {
		c = *cfg
	}
	return &Transport{addr: addr, cfg: c}
}

// NewMockTransport creates a transport that returns canned responses without real networking.
// The responder function receives the path, payload and path params and returns the raw line.
func NewMockTransport(responder func(path string, payload any, pathParams map[string]string) (string, error)) *Transport {
	return &Transport{addr: "mock", mock: responder, cfg: defaultConfig()}
}

// Extend Transport with optional mock callback (kept private to avoid external misuse).
// NOTE: This requires adding field; done by redefining struct above.

// Do sends a request and returns the exact single-line response (without trailing newline).
// Payload handling rules:
//
//	[]byte -> sent as-is
//	string -> UTF-8 bytes
//	struct/other -> JSON marshaled bytes
//	nil -> no payload appended
func (t *Transport) Do(path string, payload any, pathParams map[string]string) (string, error) {
	return t.DoCtx(context.Background(), path, payload, pathParams)
}

// DoCtx is like Do but honors the provided context and configured timeouts.
func (t *Transport) DoCtx(ctx context.Context, path string, payload any, pathParams map[string]string) (string, error) {
	if t.mock != nil {
		return t.mock(path, payload, pathParams)
	}
	fullPath := fillPath(path, pathParams)
	var lineBytes []byte
	if pb, ok := toPayloadBytes(payload); ok && len(pb) > 0 {
		lineBytes = append([]byte(fullPath+" "), pb...)
	} else {
		lineBytes = []byte(fullPath)
	}
	if err := ctx.Err(); err != nil {
		return "", fmt.Errorf("dial: %w", err)
	}
	d := &net.Dialer{Timeout: t.cfg.DialTimeout}
	conn, err := d.DialContext(ctx, "tcp", t.addr)
	if err != nil {
		return "", fmt.Errorf("dial: %w", err)
	}
	defer conn.Close()

	if tcpConn, ok := conn.(*net.TCPConn); ok {
		if err := tcpConn.SetNoDelay(true); err != nil {
			slog.Warn("failed to set TCP_NODELAY", "error", err)
		}
	}

	if t.cfg.WriteTimeout > 0 {
		_ = conn.SetWriteDeadline(time.Now().Add(t.cfg.WriteTimeout))
	}

	if t.cfg.Password != "" {
		key, err := auth.DeriveKey(t.cfg.Password)
		if err != nil {
			return "", err
		}
		r := bufio.NewReader(conn)
		clientNonce, serverNonce, err := auth.HandleAuthHandshake(r, conn, key, true)
		if err != nil {

			if strings.Contains(err.Error(), "read handshake response: EOF") {
				return "", apierror.ErrUnauthorized("invalid password")
			}
			return "", err
		}
		sessionKey := auth.DeriveSessionKey(key, serverNonce, clientNonce)
		conn, err = auth.WrapConn(conn, sessionKey)
		if err != nil {
			conn.Close()
			return "", err
		}
	}

	if _, err := conn.Write(append(lineBytes, '\x00')); err != nil {
		return "", fmt.Errorf("write: %w", err)
	}
	if t.cfg.ReadTimeout > 0 {
		_ = conn.SetReadDeadline(time.Now().Add(t.cfg.ReadTimeout))
	}
	respBytes, err := io.ReadAll(conn)
	if err != nil && len(respBytes) == 0 {
		return "", fmt.Errorf("read: %w", err)
	}
	resp := string(respBytes)

	return strings.TrimSuffix(resp, "\n"), nil
}

func fillPath(pattern string, params map[string]string) string {
	if len(params) == 0 {
		return strings.ToLower(pattern)
	}
	out := pattern
	for k, v := range params {
		esc := url.PathEscape(v)
		out = strings.ReplaceAll(out, "{"+k+"}", esc)
	}
	return strings.ToLower(out)
}

func toPayloadBytes(v any) ([]byte, bool) {
	if v == nil {
		return nil, true
	}
	switch t := v.(type) {
	case []byte:
		return t, true
	case string:
		return []byte(t), true
	default:
		b, err := json.Marshal(v)
		if err != nil {
			return nil, false
		}
		return b, true
	}
}
