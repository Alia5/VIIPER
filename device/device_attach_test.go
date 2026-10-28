package device_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/Alia5/VIIPER/internal/server/api"
	"github.com/Alia5/VIIPER/internal/server/api/handler"
	"github.com/Alia5/VIIPER/viiperclient"
	"github.com/Alia5/VIIPER/virtualbus"
	"github.com/stretchr/testify/assert"

	viiperTesting "github.com/Alia5/VIIPER/_testing"
	"github.com/Alia5/VIIPER/device"

	_ "github.com/Alia5/VIIPER/internal/registry" // Register devices
)

func TestDeviceAttach(t *testing.T) {

	deviceTypes := api.ListDeviceTypes()
	assert.NotEmpty(t, deviceTypes)

	type testCase struct {
		deviceType string
	}

	cases := make([]testCase, len(deviceTypes))
	for i, dt := range deviceTypes {
		cases[i] = testCase{deviceType: dt}
	}

	for _, tc := range cases {
		t.Run(tc.deviceType, func(t *testing.T) {

			s := viiperTesting.NewTestServer(t)
			defer s.UsbServer.Close() // nolint
			defer s.ApiServer.Close() // nolint

			r := s.ApiServer.Router()
			r.Register("bus/{id}/add", handler.BusDeviceAdd(s.UsbServer, s.ApiServer))
			r.RegisterStream("bus/{busId}/{deviceid}", api.DeviceStreamHandler(s.UsbServer))

			if err := s.ApiServer.Start(); err != nil {
				t.Fatalf("Failed to start API server: %v", err)
			}
			b, err := virtualbus.NewWithBusID(1)
			if err != nil {
				t.Fatalf("Failed to create virtual bus: %v", err)
			}
			defer b.Close() // nolint
			err = s.UsbServer.AddBus(b)
			if err != nil {
				t.Fatalf("Failed to add bus to USB server: %v", err)
			}

			c := viiperclient.New(s.ApiServer.Addr())

			stream, addResp, err := c.AddDeviceAndConnect(context.Background(), b.BusID(), tc.deviceType, createOptions(t, tc.deviceType))
			if !assert.NoError(t, err) {
				t.Fatal()
			}
			assert.NotNil(t, stream)
			assert.NotNil(t, addResp)
			assert.Equal(t, tc.deviceType, addResp.Type)
			assert.Equal(t, b.BusID(), addResp.BusID)
			assert.Equal(t, "1", addResp.DevID)

			if stream != nil {
				defer stream.Close() //nolint:errcheck
			}

			usbipClient := viiperTesting.NewUsbIpClient(t, s.UsbServer.Addr())

			var devs []viiperTesting.Device
			ok := assert.Eventually(t, func() bool {
				list, err := usbipClient.ListDevices()
				if err != nil {
					return false
				}
				devs = list
				return len(devs) == 1
			}, 1*time.Second, 10*time.Millisecond)
			if !ok {
				return
			}

			imp, err := usbipClient.AttachDevice(devs[0].BusID)
			if !assert.NoError(t, err) {
				return
			}
			if !assert.NotNil(t, imp) {
				return
			}
			if imp.Conn != nil {
				defer imp.Conn.Close() // nolint
			}
			if !assert.NotNil(t, imp.Conn) {
				return
			}

		})
	}

}

// createOptions gives the device types that can't be made without them their create options: a hidproxy copies
// the descriptors it is given (here those of the Steam Controller puck).
func createOptions(t *testing.T, deviceType string) *device.CreateOptions {
	t.Helper()
	if deviceType != "hidproxy" {
		return nil
	}
	descriptors, err := os.ReadFile("hidproxy/testdata/puck.json")
	if err != nil {
		t.Fatal(err)
	}
	return &device.CreateOptions{DeviceSpecific: string(descriptors)}
}
