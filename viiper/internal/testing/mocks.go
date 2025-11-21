package testing

import (
	"testing"
	"viiper/internal/server/api"
	"viiper/pkg/device"
	"viiper/pkg/usb"
)

type mockRegistration struct {
	deviceName  string
	handlerFunc api.StreamHandlerFunc

	createFunc func(o *device.CreateOptions) usb.Device
}

func (m *mockRegistration) CreateDevice(o *device.CreateOptions) usb.Device {
	return m.createFunc(o)
}

func (m *mockRegistration) StreamHandler() api.StreamHandlerFunc {
	return m.handlerFunc
}

func CreateMockRegistration(
	t *testing.T,
	name string,
	cf func(o *device.CreateOptions) usb.Device,
	h api.StreamHandlerFunc,
) api.DeviceRegistration {
	return &mockRegistration{
		deviceName:  name,
		handlerFunc: h,
		createFunc:  cf,
	}
}
