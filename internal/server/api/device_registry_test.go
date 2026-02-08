package api_test

import (
	"log/slog"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Alia5/VIIPER/device"
	th "github.com/Alia5/VIIPER/internal/_testing"
	"github.com/Alia5/VIIPER/internal/server/api"
	"github.com/Alia5/VIIPER/usb"
)

type mockDevice struct {
	name string
}

func (m *mockDevice) HandleTransfer(ep uint32, dir uint32, out []byte) []byte {
	return nil
}
func (m *mockDevice) GetDescriptor() *usb.Descriptor {
	return &usb.Descriptor{}
}
func (m *mockDevice) GetDeviceSpecificArgs() map[string]any {
	return map[string]any{}
}

func TestDeviceRegistry(t *testing.T) {

	// TODO: test with OPTS

	tests := []struct {
		name           string
		registerName   string
		lookupName     string
		shouldFind     bool
		expectedDevice string
	}{
		{
			name:           "register and retrieve exact match",
			registerName:   "testdevice",
			lookupName:     "testdevice",
			shouldFind:     true,
			expectedDevice: "testdevice",
		},
		{
			name:           "case insensitive lookup",
			registerName:   "TestDevice",
			lookupName:     "testdevice",
			shouldFind:     true,
			expectedDevice: "TestDevice",
		},
		{
			name:           "case insensitive lookup uppercase",
			registerName:   "mydevice",
			lookupName:     "MYDEVICE",
			shouldFind:     true,
			expectedDevice: "mydevice",
		},
		{
			name:         "lookup non-existent device",
			registerName: "device1",
			lookupName:   "device2",
			shouldFind:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testRegName := tt.name + "_" + tt.registerName

			handlerCalled := false
			mockHandler := func(conn net.Conn, dev *usb.Device, logger *slog.Logger) error {
				handlerCalled = true
				return nil
			}

			reg := th.CreateMockRegistration(
				t,
				tt.expectedDevice,
				func(o *device.CreateOptions) (usb.Device, error) { return &mockDevice{name: tt.expectedDevice}, nil },
				mockHandler,
			)

			api.RegisterDevice(testRegName, reg)

			testLookupName := tt.name + "_" + tt.lookupName
			retrieved := api.GetRegistration(testLookupName)

			if tt.shouldFind {
				assert.NotNil(t, retrieved, "expected to find registered device")
				if retrieved != nil {
					dev, err := retrieved.CreateDevice(nil)
					assert.NoError(t, err)
					mockDev, ok := dev.(*mockDevice)
					assert.True(t, ok, "expected mockDevice type")
					if ok {
						assert.Equal(t, tt.expectedDevice, mockDev.name)
					}

					var dv usb.Device = &mockDevice{name: tt.expectedDevice}
					handler := retrieved.StreamHandler()
					assert.NotNil(t, handler, "expected handler to be non-nil")
					if handler != nil {
						_ = handler(nil, &dv, slog.Default())
						assert.True(t, handlerCalled, "expected handler to be called")
					}
				}
			} else {
				assert.Nil(t, retrieved, "expected not to find device")
			}
		})
	}
}

func TestGetStreamHandler(t *testing.T) {

	// TODO: test with OPTS

	tests := []struct {
		name         string
		registerName string
		lookupName   string
		shouldFind   bool
	}{
		{
			name:         "get stream handler for registered device",
			registerName: "streamtest1",
			lookupName:   "streamtest1",
			shouldFind:   true,
		},
		{
			name:         "get stream handler case insensitive",
			registerName: "StreamTest2",
			lookupName:   "streamtest2",
			shouldFind:   true,
		},
		{
			name:         "get stream handler for non-existent device",
			registerName: "streamtest3",
			lookupName:   "nonexistent",
			shouldFind:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handlerCalled := false
			mockHandler := func(conn net.Conn, dev *usb.Device, logger *slog.Logger) error {
				handlerCalled = true
				return nil
			}

			reg := th.CreateMockRegistration(
				t,
				tt.registerName,
				func(o *device.CreateOptions) (usb.Device, error) { return &mockDevice{name: tt.registerName}, nil },
				mockHandler,
			)

			testRegName := tt.name + "_" + tt.registerName
			api.RegisterDevice(testRegName, reg)

			testLookupName := tt.name + "_" + tt.lookupName
			handler := api.GetStreamHandler(testLookupName)

			if tt.shouldFind {
				assert.NotNil(t, handler, "expected to find stream handler")
				if handler != nil {
					_ = handler(nil, nil, slog.Default())
					assert.True(t, handlerCalled, "expected handler to be called")
				}
			} else {
				assert.Nil(t, handler, "expected not to find stream handler")
			}
		})
	}
}
