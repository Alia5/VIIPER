package log

import (
	"bytes"
	"fmt"
	"io"
	"sync"
	"time"
)

// RawLogger handles raw packet log with optional file output.
type RawLogger interface {
	Log(in bool, data []byte)
}

// rawLogger implements RawLogger with thread-safe log.
type rawLogger struct {
	w  io.Writer
	mu sync.Mutex
}

// NewRaw creates a new RawLogger. If writer is nil, returns a no-op logger.
func NewRaw(w io.Writer) RawLogger {
	return &rawLogger{w: w}
}

// Log emits a single-line raw packet log with timestamp and hex dump.
// in=true means client->server, in=false means server->client.
func (r *rawLogger) Log(in bool, data []byte) {
	if len(data) == 0 {
		return
	}
	if r.w == nil {
		return
	}

	dir := "S->C"
	if in {
		dir = "C->S"
	}

	var hexbuf bytes.Buffer
	const hexdigits = "0123456789abcdef"
	for i, b := range data {
		if i > 0 {
			hexbuf.WriteByte(' ')
		}
		hexbuf.WriteByte(hexdigits[b>>4])
		hexbuf.WriteByte(hexdigits[b&0x0f])
	}

	line := fmt.Sprintf("%s %s chunk: %d bytes, hex: %s\n",
		time.Now().Format("2006/01/02 15:04:05"),
		dir,
		len(data),
		hexbuf.String())

	r.mu.Lock()
	_, _ = r.w.Write([]byte(line))
	r.mu.Unlock()
}
