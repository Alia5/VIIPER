package api

import (
	"sync"

	"github.com/Alia5/VIIPER/device"
	"github.com/Alia5/VIIPER/usb"
)

// DeviceRegistration describes a device type, providing both device creation
// and stream handler registration.
type DeviceRegistration interface {
	// CreateDevice returns a new device instance of this type.
	CreateDevice(o *device.CreateOptions) (usb.Device, error)
	// StreamHandler returns the handler function for long-lived connections.
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

// ListDeviceTypes returns a list of all registered device type names.
func ListDeviceTypes() []string {
	deviceRegistryMu.RLock()
	defer deviceRegistryMu.RUnlock()
	types := make([]string, 0, len(deviceRegistry))
	for name := range deviceRegistry {
		types = append(types, name)
	}
	return types
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
	b := []byte(s)
	for i := range b {
		if b[i] >= 'A' && b[i] <= 'Z' {
			b[i] += 'a' - 'A'
		}
	}
	return string(b)
}
