package api

import (
	"sync"
	"viiper/pkg/usb"
)

// NOTE: Stream handlers now share the unified StreamHandler type (see router.go).
// They return an error to signal terminal failures; successful completion should
// generally return nil. Handlers still own the connection lifecycle.

// DeviceRegistration describes a device type, providing both device creation
// and stream handler registration.
type DeviceRegistration interface {
	// CreateDevice returns a new device instance of this type.
	CreateDevice() usb.Device
	// StreamHandler returns the handler function for long-lived connections.
	// The provided device (if non-nil) can be used to bind handler behavior, but handlers
	// should not rely on it being set at registration time.
	StreamHandler() StreamHandlerFunc
}

var (
	deviceRegistry   = make(map[string]DeviceRegistration)
	deviceRegistryMu sync.RWMutex
)

// RegisterDevice registers a device type for dynamic creation and handler dispatch.
// This should be called from device package init() functions.
// The name is case-insensitive and will be lowercased.
func RegisterDevice(name string, reg DeviceRegistration) {
	deviceRegistryMu.Lock()
	defer deviceRegistryMu.Unlock()
	deviceRegistry[toLower(name)] = reg
}

// GetRegistration retrieves a registered device handler by name for device creation.
// Returns nil if not found. Name lookup is case-insensitive.
func GetRegistration(name string) DeviceRegistration {
	deviceRegistryMu.RLock()
	defer deviceRegistryMu.RUnlock()
	return deviceRegistry[toLower(name)]
}

// GetStreamHandler retrieves the stream handler for a registered device type.
// Returns nil if not found. Name lookup is case-insensitive.
func GetStreamHandler(name string) StreamHandlerFunc {
	handler := GetRegistration(name)
	if handler == nil {
		return nil
	}
	return handler.StreamHandler()
}

func toLower(s string) string {
	// Simple ASCII lowercase for device type names
	b := []byte(s)
	for i := range b {
		if b[i] >= 'A' && b[i] <= 'Z' {
			b[i] += 'a' - 'A'
		}
	}
	return string(b)
}
