package api

import (
	"fmt"
	"log/slog"
	"net"
	"path/filepath"
	"reflect"
	"strings"
	"viiper/internal/server/usb"
	pusb "viiper/pkg/usb"
)

// DeviceStreamHandler returns a stream handler func that dynamically dispatches
// to device-specific handlers based on device type.
func DeviceStreamHandler(srv *usb.Server) StreamHandlerFunc {
	return func(conn net.Conn, dev *pusb.Device, logger *slog.Logger) error {
		defer conn.Close()

		if dev == nil || *dev == nil {
			return fmt.Errorf("nil device")
		}

		deviceType := inferDeviceType(*dev)
		reg := GetRegistration(deviceType)
		if reg == nil {
			return fmt.Errorf("no handler for device type: %s", deviceType)
		}
		handler := reg.StreamHandler()
		if err := handler(conn, dev, logger); err != nil {
			return err
		}
		return nil
	}
}

func inferDeviceType(dev any) string {
	if dev == nil {
		return ""
	}
	t := reflect.TypeOf(dev)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	pkg := t.PkgPath() // e.g., "viiper/pkg/device/xbox360"
	if pkg != "" {
		base := filepath.Base(pkg)
		if base != "." && base != string(filepath.Separator) {
			return strings.ToLower(base)
		}
	}
	return strings.ToLower(t.Name())
}
