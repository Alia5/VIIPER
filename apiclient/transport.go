package apiclient

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/url"
	"strings"
	"time"
)

// Config controls low-level transport behavior such as timeouts.
type Config struct {
	DialTimeout  time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
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
func (c *Transport) Do(path string, payload any, pathParams map[string]string) (string, error) {
	return c.DoCtx(context.Background(), path, payload, pathParams)
}

// DoCtx is like Do but honors the provided context and configured timeouts.
func (c *Transport) DoCtx(ctx context.Context, path string, payload any, pathParams map[string]string) (string, error) {
	if c.mock != nil {
		return c.mock(path, payload, pathParams)
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
	d := &net.Dialer{Timeout: c.cfg.DialTimeout}
	conn, err := d.DialContext(ctx, "tcp", c.addr)
	if err != nil {
		return "", fmt.Errorf("dial: %w", err)
	}
	defer conn.Close()

	if tcpConn, ok := conn.(*net.TCPConn); ok {
		if err := tcpConn.SetNoDelay(true); err != nil {
			slog.Warn("failed to set TCP_NODELAY", "error", err)
		}
	}

	if c.cfg.WriteTimeout > 0 {
		_ = conn.SetWriteDeadline(time.Now().Add(c.cfg.WriteTimeout))
	}
	// Send request with null terminator
	if _, err := conn.Write(append(lineBytes, '\x00')); err != nil {
		return "", fmt.Errorf("write: %w", err)
	}
	if c.cfg.ReadTimeout > 0 {
		_ = conn.SetReadDeadline(time.Now().Add(c.cfg.ReadTimeout))
	}
	respBytes, err := io.ReadAll(conn)
	if err != nil && len(respBytes) == 0 {
		return "", fmt.Errorf("read: %w", err)
	}
	resp := string(respBytes)
	// Trim exactly one trailing newline if present.
	if strings.HasSuffix(resp, "\n") {
		resp = strings.TrimSuffix(resp, "\n")
	}
	return resp, nil
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
