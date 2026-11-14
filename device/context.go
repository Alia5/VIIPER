// Package device provides common interfaces and utilities for virtual USB devices.
package device

import (
	"context"
	"net"
	"time"

	"github.com/Alia5/VIIPER/usbip"
)

type contextKey string

const (
	ExportMetaKey contextKey = "exportMeta"
	ConnTimerKey  contextKey = "connTimer"
	ConnKey       contextKey = "conn"
)

func GetConn(ctx context.Context) net.Conn {
	conn, ok := ctx.Value(ConnKey).(net.Conn)
	if ok {
		return conn
	}
	return nil
}

// GetDeviceMeta extracts the device metadata from a device context.
// Returns nil if the context doesn't contain device metadata.
func GetDeviceMeta(ctx context.Context) *usbip.ExportMeta {
	if meta, ok := ctx.Value(ExportMetaKey).(*usbip.ExportMeta); ok {
		return meta
	}
	return nil
}

// GetConnTimer extracts the connection timer from a device context.
// Returns nil if the context doesn't contain the timer.
func GetConnTimer(ctx context.Context) *time.Timer {
	if timer, ok := ctx.Value(ConnTimerKey).(*time.Timer); ok {
		return timer
	}
	return nil
}
