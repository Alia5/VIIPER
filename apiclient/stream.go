package apiclient

import (
	"bufio"
	"context"
	"encoding"
	"fmt"
	"io"
	"log/slog"
	"net"
	"sync"
	"time"

	apitypes "github.com/Alia5/VIIPER/apitypes"
	"github.com/Alia5/VIIPER/device"
	"github.com/Alia5/VIIPER/internal/server/api/auth"
)

// DeviceStream represents a bidirectional connection to a device stream.
type DeviceStream struct {
	conn   net.Conn
	BusID  uint32
	DevID  string
	closed bool

	readCancel context.CancelFunc
	readMu     sync.Mutex
}

// OpenStream connects to an existing device's stream channel.
// The device must already exist on the bus (use DeviceAdd first).
func (c *Client) OpenStream(ctx context.Context, busID uint32, devID string) (*DeviceStream, error) {
	addr := c.transport.addr
	if c.transport.mock != nil {
		return nil, fmt.Errorf("stream connections not supported with mock transport")
	}

	d := &net.Dialer{Timeout: c.transport.cfg.DialTimeout}
	conn, err := d.DialContext(ctx, "tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("dial: %w", err)
	}

	if tcpConn, ok := conn.(*net.TCPConn); ok {
		if err := tcpConn.SetNoDelay(true); err != nil {
			slog.Warn("failed to set TCP_NODELAY", "error", err)
		}
	}

	if c.transport.cfg.Password != "" {
		key, err := auth.DeriveKey(c.transport.cfg.Password)
		if err != nil {
			return nil, err
		}
		r := bufio.NewReader(conn)
		clientNonce, serverNonce, err := auth.HandleAuthHandshake(r, conn, key, true)
		if err != nil {
			return nil, err
		}
		sessionKey := auth.DeriveSessionKey(key, serverNonce, clientNonce)
		conn, err = auth.WrapConn(conn, sessionKey)
		if err != nil {
			conn.Close()
			return nil, err
		}
	}

	streamPath := fmt.Sprintf("bus/%d/%s\x00", busID, devID)
	if _, err := conn.Write([]byte(streamPath)); err != nil {
		conn.Close()
		return nil, fmt.Errorf("write stream path: %w", err)
	}

	ds := &DeviceStream{
		conn:  conn,
		BusID: busID,
		DevID: devID,
	}
	return ds, nil
}

// AddDeviceAndConnect creates a device on the specified bus and immediately connects to its stream.
// This is a convenience wrapper that combines DeviceAdd + OpenStream in one call.
func (c *Client) AddDeviceAndConnect(ctx context.Context, busID uint32, deviceType string, o *device.CreateOptions) (*DeviceStream, *apitypes.Device, error) {
	resp, err := c.DeviceAddCtx(ctx, busID, deviceType, o)
	if err != nil {
		return nil, nil, err
	}

	stream, err := c.OpenStream(ctx, busID, resp.DevId)
	if err != nil {
		return nil, resp, err
	}

	return stream, resp, nil
}

// Write sends raw bytes to the device stream (client → device input).
func (s *DeviceStream) Write(data []byte) (int, error) {
	if s.closed {
		return 0, fmt.Errorf("stream closed")
	}
	return s.conn.Write(data)
}

// WriteBinary marshals and sends a BinaryMarshaler to the device stream.
// This is the preferred way to send device input (e.g., xbox360.InputState, keyboard.InputState).
func (s *DeviceStream) WriteBinary(v encoding.BinaryMarshaler) error {
	if s.closed {
		return fmt.Errorf("stream closed")
	}
	data, err := v.MarshalBinary()
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}
	_, err = s.conn.Write(data)
	return err
}

// Read receives raw bytes from the device stream (device → client feedback).
// For event-driven reading, use StartReading() instead to avoid blocking/polling.
func (s *DeviceStream) Read(buf []byte) (int, error) {
	if s.closed {
		return 0, fmt.Errorf("stream closed")
	}
	return s.conn.Read(buf)
}

// StartReading begins asynchronously reading from the device stream in a background goroutine.
// You provide a decode function that reads exactly one message from the given *bufio.Reader
// and returns any value that implements encoding.BinaryUnmarshaler (the interface is only
// used for typing; StartReading does not call UnmarshalBinary itself).
//
// Example (xbox360 rumble, fixed 2 bytes):
//
//	rumbleCh, errCh := stream.StartReading(ctx, 10, func(r *bufio.Reader) (encoding.BinaryUnmarshaler, error) {
//	    var b [2]byte
//	    if _, err := io.ReadFull(r, b[:]); err != nil { return nil, err }
//	    msg := new(xbox360.XRumbleState)
//	    if err := msg.UnmarshalBinary(b[:]); err != nil { return nil, err }
//	    return msg, nil
//	})
func (s *DeviceStream) StartReading(ctx context.Context, chSize int, decode func(r *bufio.Reader) (encoding.BinaryUnmarshaler, error)) (<-chan encoding.BinaryUnmarshaler, <-chan error) {
	s.readMu.Lock()
	defer s.readMu.Unlock()

	if s.readCancel != nil {
		panic("StartReading called twice on the same stream")
	}

	msgCh := make(chan encoding.BinaryUnmarshaler, chSize)
	errCh := make(chan error, 1)

	readCtx, cancel := context.WithCancel(ctx)
	s.readCancel = cancel

	go func() {
		defer close(msgCh)
		defer close(errCh)
		defer cancel()

		r := bufio.NewReader(s.conn)
		for {
			select {
			case <-readCtx.Done():
				errCh <- readCtx.Err()
				return
			default:
			}

			if s.closed {
				errCh <- io.EOF
				return
			}

			msg, err := decode(r)
			if err != nil {
				errCh <- err
				return
			}

			select {
			case msgCh <- msg:
			case <-readCtx.Done():
				errCh <- readCtx.Err()
				return
			}
		}
	}()

	return msgCh, errCh
}

// SetReadDeadline sets the read deadline for the underlying connection.
func (s *DeviceStream) SetReadDeadline(t time.Time) error {
	return s.conn.SetReadDeadline(t)
}

// SetWriteDeadline sets the write deadline for the underlying connection.
func (s *DeviceStream) SetWriteDeadline(t time.Time) error {
	return s.conn.SetWriteDeadline(t)
}

// Close closes the stream connection and stops any background reading.
func (s *DeviceStream) Close() error {
	if s.closed {
		return nil
	}
	s.closed = true

	s.readMu.Lock()
	if s.readCancel != nil {
		s.readCancel()
	}
	s.readMu.Unlock()

	return s.conn.Close()
}
