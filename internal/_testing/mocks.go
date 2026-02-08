package testing

import (
	"testing"

	"github.com/Alia5/VIIPER/device"
	"github.com/Alia5/VIIPER/internal/server/api"
	"github.com/Alia5/VIIPER/usb"
)

type mockRegistration struct {
	deviceName  string
	handlerFunc api.StreamHandlerFunc

	createFunc func(o *device.CreateOptions) (usb.Device, error)
}

func (m *mockRegistration) CreateDevice(o *device.CreateOptions) (usb.Device, error) {
	return m.createFunc(o)
}

func (m *mockRegistration) StreamHandler() api.StreamHandlerFunc {
	return m.handlerFunc
}

func CreateMockRegistration(
	t *testing.T,
	name string,
	cf func(o *device.CreateOptions) (usb.Device, error),
	h api.StreamHandlerFunc,
) api.DeviceRegistration {
	return &mockRegistration{
		deviceName:  name,
		handlerFunc: h,
		createFunc:  cf,
	}
}
